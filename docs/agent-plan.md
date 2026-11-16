# 平台 Agent 建设方案（基于 Pi 核心）

- 状态：Phase 0/1 已实现（2026-11-12），Phase 2/3 待做
- 日期：2026-11-12
- 范围：为 todo-app 增加一个平台内 AI 助手（对话式，自主选择工具操作平台数据）

## 1. 目标与非目标

**目标**

- 用户在 Web 面板中对话，agent 自主决定调用哪些平台工具，完成查询、创建、修改、归档、统计等任务。
- agent 以用户身份行动，不越权；写操作可确认；全链路可审计。
- 架构分层清晰、每层可单独解释，可用于面试叙事。

**非目标（明确不做）**

- 通用 coding agent（不暴露文件系统 / shell 工具）。
- 多 agent 协作、RAG / 向量库、MCP Server、低代码平台。
- 本阶段不做后台无人值守 agent（所有操作有实时用户在场）。

## 2. 总体架构

```
Browser（React 聊天面板）
   │  POST /agent/chat（SSE 响应，携带 Authorization: Bearer <用户 JWT>）
   ▼
web（nginx）
   ├─ /api/   → api:8080（Go，现有）
   └─ /agent/ → agent:3001（新增，SSE 反代）
   ▼
agent（Node 22 + TS + Hono）
   ├─ Pi SDK（createAgentSession）：agent 循环、工具调用、事件流
   ├─ customTools：平台工具，内部用 HTTP 调 Go API（透传用户 JWT）
   ├─ 会话快照：Redis；同 thread 串行执行
   ├─ 写操作两阶段审批（pending action）
   ├─ 审计日志（Phase 2）
   └─ 限额：步数 / Token / 超时
   ▼
api（Go + Gin，业务唯一事实来源）
   ▼
MySQL（业务数据） / Redis（会话、缓存）
```

**部署形态**：`docker-compose` 增加一个 `agent` 服务，仅在 compose 内网可达，不暴露公网端口；对外统一走 web 的 nginx。

## 3. 技术选型

| 项 | 选择 | 理由 |
|---|---|---|
| Agent 内核 | Pi SDK（`@earendil-works/pi-coding-agent`，底层 `pi-ai` / `pi-agent-core`） | 循环 + 工具调用 + 多 provider 抽象现成；极简、可读、MIT |
| 运行时 | Node.js 22 LTS + TypeScript | Pi 生态要求 Node ≥ 22；与前端同语言 |
| HTTP 服务 | Hono + `@hono/node-server` | 轻量，SSE 支持直接 |
| 模型接入 | OpenAI 兼容接口，默认 DeepSeek，env 可切换 | 成本可控、国内可用；由 `pi-ai` 统一抽象 |
| 前端事件流 | SSE + AG-UI 1.0（`@ag-ui/core` + `@ag-ui/encoder`） | 标准事件、schema 校验；前端按 messageId / toolCallId 关联 |
| 前端消息组件 | 自绘气泡与输入框（Element Plus + 主题变量） | 与现有视觉一致，克制描边风格，不引入第三方 AI 组件库 |
| 会话/审批/锁 | Redis（复用现有实例） | 无需新增基础设施 |
| 审计 | Phase 2：Go 内部端点 + MySQL 表 | 数据归属清晰，可追溯 |
| 明确不用 | LangGraph、MCP Server、向量库、多 agent | 当前规模下均为负收益 |

## 4. Agent 如何调用 Go（核心）

### 4.1 方式：内网 HTTP + 用户 JWT 透传

agent 不直连数据库，也不使用服务账号，而是**以用户身份**调用现有 Go REST API：

1. 前端从 localStorage 取出 `todo.token`，在每次 Agent 请求中携带
   `Authorization: Bearer <token>`。
2. agent 服务**不解析、不签发、不持久化**该 JWT，只在单次 run 的内存上下文中保存。
3. 开跑前先调用 Go 的 `GET /api/auth/me` 预检：token 无效或过期直接返回 401，不进入模型调用；Go 不可用时返回 502。
4. 工具的 HTTP 客户端把同一个 `Authorization` 头透传给 Go API。
5. Go 侧现有的 `authMiddleware.Required()` 照常校验 token（HS256，Claims.Subject = userID），业务逻辑与权限零改动。

**为什么这样设计**

- 不复用服务账号：个人助手必须以用户身份行动，服务账号无法表达“这个用户只能操作自己的任务”。
- 不直连 MySQL：权限、业务校验、审计都会分裂成两份；Go 是唯一事实来源。
- agent 不验签，但开跑前用 `/auth/me` 预检：无效 token 不消耗模型调用；验签算法与 JWT_SECRET 仍然只存在于 Go，Node 不复制验签逻辑。
- 预检不缓存：每次对话多一次轻量内部请求，换取失效 token 立即被挡住。

### 4.2 请求时序

```
Browser                agent                     Go API                 MySQL
   │ POST /agent/chat     │                         │                     │
   │ + Bearer JWT         │                         │                     │
   ├─────────────────────►│                         │                     │
   │                      │ GET /api/auth/me 预检   │                     │
   │                      ├────────────────────────►│ 校验 JWT            │
   │                      │◄──────── user ──────────┤                     │
   │                      │ 401 则直接返回，不调模型 │                     │
   │                      │ 创建 run（保存 JWT 到内存）│                     │
   │                      │ 模型推理 → 决定调工具     │                     │
   │                      │ GET /api/tasks          │                     │
   │                      │ + Bearer JWT            │                     │
   │                      ├────────────────────────►│ 校验 JWT → 查询      │
   │                      │                         ├────────────────────►│
   │                      │◄────── {"tasks": [...]}─┤◄────────────────────┤
   │                      │ 工具结果回填 → 再推理      │                     │
   │◄─ SSE events ────────┤                         │                     │
   │  (text / tool / done)│                         │                     │
```

### 4.3 工具实现示例（TypeScript）

> 工具定义的确切字段以 Pi 的 `AgentTool` 类型为准，以下为结构与写法示意。

```ts
// src/http.ts —— 统一的 Go API 客户端
const BASE = process.env.TODO_API_URL ?? 'http://localhost:8080'

export async function callApi<T>(
  token: string,
  path: string,
  init: { method?: string; body?: unknown } = {},
): Promise<T> {
  const res = await fetch(`${BASE}/api${path}`, {
    method: init.method ?? 'GET',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`, // 只在内存中透传，禁止写日志
    },
    body: init.body === undefined ? undefined : JSON.stringify(init.body),
    signal: AbortSignal.timeout(5000),
  })
  const data = await res.json().catch(() => null)
  if (!res.ok) {
    const message =
      data && typeof data === 'object' && 'error' in data
        ? String((data as { error: unknown }).error)
        : `HTTP ${res.status}`
    throw new Error(message) // 由调用方转为工具结果交给模型
  }
  return data as T
}
```

```ts
// src/tools/listTasks.ts —— 只读工具
export const listTasksTool = {
  name: 'list_tasks',
  description: '查询当前用户的任务列表，可按状态/优先级/是否归档筛选。返回精简字段。',
  parameters: {
    type: 'object',
    properties: {
      status: { type: 'string', enum: ['todo', 'doing', 'done'] },
      includeArchived: { type: 'boolean', default: false },
    },
  },
  async execute(args: { status?: string; includeArchived?: boolean }, ctx: { token: string }) {
    const { tasks } = await callApi<{ tasks: any[] }>(ctx.token, '/tasks')
    const filtered = tasks.filter(
      (t) =>
        (args.includeArchived || !t.archived_at) &&
        (!args.status || t.status === args.status),
    )
    return filtered.slice(0, 50).map((t) => ({
      id: t.id,
      title: t.title,
      status: t.status,
      priority: t.priority,
      subtaskDone: t.subtasks.filter((s: any) => s.done).length,
      subtaskTotal: t.subtasks.length,
      archived: !!t.archived_at,
    }))
  },
}
```

### 4.4 工具 ↔ Go API 映射（基于现有路由）

| 工具 | Go 接口 | 说明 |
|---|---|---|
| `list_tasks` | `GET /api/tasks` | 现有接口无筛选参数，agent 侧过滤；返回上限 50 条、精简字段 |
| `get_task` | `GET /api/tasks` | 按 ID 取完整任务（描述、备注、子任务），修改前读取 |
| `create_task` | `POST /api/tasks` | 字段：`title` / `description` / `notes` / `subtasks[]` / `status` / `priority` |
| `update_task` | `PATCH /api/tasks/:id` | 字段：`title` / `description` / `notes` / `subtasks[]` / `status` / `priority` / `pinned` |
| `complete_task` | `GET /api/tasks` + `PATCH /api/tasks/:id` | 按标题关键词完成；重名时返回候选让用户确认，不擅自修改 |
| `list_notes` | `GET /api/notes` | 便签实体（content / color） |
| `create_note` | `POST /api/notes` | 颜色枚举：yellow / green / blue / pink / purple |
| `update_note` | `PATCH /api/notes/:id` | — |
| `delete_note` | `DELETE /api/notes/:id` | 危险操作，需审批 |
| `get_daily_stats` | 无独立接口 | MVP 由 agent 汇总 `GET /api/tasks` 计算；Phase 2 可加 `GET /api/tasks/stats` |
| `reorder_tasks` | `POST /api/tasks/reorder` | V2 再做，涉及列内顺序，低优先 |

**两个必须注意的模型细节**

1. **子任务没有独立端点**：`subtasks` 是任务上的 JSON 数组，修改单个子任务需要“先 GET 整个任务 → 改数组 → PATCH 整个数组”。工具内部实现这个读改写，不要暴露给模型。
2. **两种 “notes” 不要混淆**：`task.notes` 是任务内的 Markdown 长文本；`/api/notes` 是独立便签实体。工具命名和描述必须区分清楚（如 `update_task_notes` vs `create_sticky_note`），否则模型必然调错。

### 4.5 错误处理与幂等

- Go 错误统一为 `{"error": "..."}`；工具捕获后**把错误文本作为工具结果返回给模型**，让模型决定重试或调整参数，而不是直接中断 run。
- 网络错误、超时（5s）同样转为工具结果；连续失败 N 次后终止 run 并返回错误事件。
- 写操作幂等：审批流程中每个 pending action 生成唯一 `action_id`，确认后只允许执行一次（Redis `SETNX` 去重）；Phase 2 可在 Go 侧支持 `Idempotency-Key`。
- 禁止把 JWT、模型 API Key 写入日志或审计记录。

### 4.6 Go 侧需要的改动

- **MVP（Phase 0/1）：零改动**，直接复用现有 `/api` 路由。
- **Phase 2**：新增 `POST /internal/agent/audit`，使用独立服务令牌（env `AGENT_SERVICE_TOKEN`，请求头 `X-Service-Token`），写入 `agent_audit_logs` 表。
- **可选**：`GET /api/tasks/stats`（每日完成/创建统计），减少 agent 每次拉全量任务。
- 需要信任链说明：agent 服务本身不鉴权用户，只透传；Go 对每个请求独立校验。

### 4.7 部署与代理配置

`docker-compose.prod.yml` 新增：

```yaml
  agent:
    image: todo-app-agent:${IMAGE_TAG:-latest}
    restart: always
    environment:
      PORT: "3001"
      TODO_API_URL: "http://api:8080"
      DEEPSEEK_API_KEY: ${DEEPSEEK_API_KEY}
      MODEL_PROVIDER: ${MODEL_PROVIDER:-deepseek}
      MODEL_ID: ${MODEL_ID:-deepseek-v4-pro}
      AGENT_DATA_DIR: /data
      SETTINGS_STORE: "redis"
      REDIS_ADDR: "cache:6379"
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      REDIS_DB: "0"
      MODEL_SETTINGS_SECRET: ${MODEL_SETTINGS_SECRET}
    depends_on:
      api:
        condition: service_healthy
      cache:
        condition: service_healthy
```

不暴露 `ports`，仅内网。`apps/web/nginx.conf` 增加 SSE 反代：

```nginx
location /agent/ {
    proxy_pass http://agent:3001;
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_set_header Host $host;
    proxy_buffering off;          # SSE 必须关闭缓冲
    proxy_read_timeout 300s;
}
```

本地开发：`TODO_API_URL=http://localhost:8080`，直接 `npm run dev`。

### 4.8 用户自定义模型（BYOK）

用户可以在助手悬浮窗右上角的设置里配置自己的模型，保存后下一条消息生效。

**接口（agent，均需登录预检）**

| 接口 | 说明 |
|---|---|
| `GET /agent/model-options` | 返回 pi-ai 模型目录（provider / model），供前端下拉选择 |
| `GET /agent/model-settings` | 返回当前配置，Key 只返回掩码（如 `****1234`）；未配置时 `usingDefault: true` |
| `PUT /agent/model-settings` | `{ provider, modelId, apiKey? }` 保存；`apiKey` 省略时保留旧 Key；`{ clearApiKey: true }` 清除并回落到平台默认 |

**存储**

- Redis key：`agent:model-settings:{userID}`（AOF 持久化）；本地开发可 `SETTINGS_STORE=memory`。
- API Key 用 AES-256-GCM 加密后存储，密钥来自 `MODEL_SETTINGS_SECRET`；明文不落盘、不写日志、不返回前端。
- Redis 模式下缺少 `MODEL_SETTINGS_SECRET` 时拒绝保存并返回 503。

**运行时**

- 每次 run 读取该用户配置；未配置则使用平台默认（`MODEL_PROVIDER` / `MODEL_ID` + `DEEPSEEK_API_KEY`）。
- 每个用户一个 `ModelRuntime`：`setRuntimeApiKey(provider, key)` 设置 Key，配置变更后用 `session.setModel(model)` 切换模型，不清空会话。
- 没有可用 Key 时返回 `RUN_ERROR`：「尚未配置模型 API Key，请点击助手右上角的设置按钮填写」。
- 模型调用失败（401 / invalid key 等）统一转为「模型服务鉴权失败，请检查 API Key」，原始错误只写服务端日志。

**环境变量**：`SETTINGS_STORE=redis|memory`、`REDIS_ADDR`、`REDIS_PASSWORD`、`REDIS_DB`、`MODEL_SETTINGS_SECRET`（Redis 模式必填）；平台默认模型用 `MODEL_PROVIDER` / `MODEL_ID` / `DEEPSEEK_API_KEY`。

## 5. Pi 接入方式

- Node ≥ 22 LTS；安装 SDK（以官方 npm 包为准，锁定版本）。
- 关键：**压制 coding-agent 默认行为**——清空技能发现、只注册白名单工具、覆盖系统提示、会话交给自建存储。

```ts
const resourceLoader = new DefaultResourceLoader({
  cwd: process.cwd(),
  agentDir: getAgentDir(),
  systemPromptOverride: () => PLATFORM_SYSTEM_PROMPT,
  appendSystemPromptOverride: () => [],
  skillsOverride: () => ({ skills: [], diagnostics: [] }),
})
await resourceLoader.reload()

const { session } = await createAgentSession({
  resourceLoader,
  customTools: todoTools,
  tools: todoTools.map((t) => t.name),
  sessionManager: SessionManager.inMemory(),
})

session.subscribe((event) => {
  // 映射为 AG-UI 事件后写入 SSE
})

await session.prompt(userText)
```

- 工具执行需要拿到当前用户 JWT：通过 run 级上下文注入（每个请求创建独立 run 对象，工具闭包或执行上下文读取），禁止用全局变量。
- Pi SDK 较新且迭代快：锁版本、通读 `packages/agent` 源码；`AgentTool` 的具体字段以源码类型为准。

## 6. SSE 事件协议（AG-UI 1.0）

服务端使用 `@ag-ui/core` 的事件类型与 `EventSchema` 校验，`@ag-ui/encoder` 编码为标准 SSE 帧（只含 `data: {...}`），前端按 `messageId` / `toolCallId` 关联消息与工具步骤。

| 事件 | 关键字段 | 前端行为 |
|---|---|---|
| `RUN_STARTED` | threadId, runId, protocolVersion | 开始态 |
| `TEXT_MESSAGE_START` | messageId, role | 创建助手消息 |
| `TEXT_MESSAGE_CONTENT` | messageId, delta | 追加流式文本 |
| `TEXT_MESSAGE_END` | messageId | 关闭消息 |
| `TOOL_CALL_START` | toolCallId, toolCallName, parentMessageId | 显示工具步骤 |
| `TOOL_CALL_ARGS` | toolCallId, delta（JSON 文本） | 记录参数 |
| `TOOL_CALL_END` | toolCallId | 参数流结束 |
| `TOOL_CALL_RESULT` | messageId, toolCallId, content | 工具结果 |
| `CUSTOM`（`agent.tool_status`） | toolCallId, ok, durationMs | 步骤完成 / 失败状态 |
| `RUN_FINISHED` | threadId, runId | 结束态 |
| `RUN_ERROR` | message, code | 错误提示 |

尚未使用的规范能力：`STATE_*`、`MESSAGES_SNAPSHOT`、`STEP_*` 与审批中断（HITL），计划在审批与状态同步阶段接入。

## 7. 会话、并发、审批、限额、审计

- **Redis Key 设计**
  - `agent:session:{userID}:{threadID}`：会话消息快照（JSON）
  - `agent:pending:{actionID}`：待确认的写操作（含参数摘要、过期时间）
  - `agent:lock:{threadID}`：同 thread 互斥锁
- **并发**：同一 thread 任意时刻只允许一个 run（单实例用进程内队列 + Mutex；多实例加 Redis 锁）。
- **审批流程**：写工具返回 `PendingAction` → 前端确认卡 → `POST /agent/confirm` → 校验归属与过期时间 → 执行 → 记审计。
- **限额**：单 run 最多 8 步工具调用；单 run token 预算；总超时 60s；按用户限流。
- **审计字段**：user_id、thread_id、run_id、tool、参数（脱敏）、结果状态、耗时、token 用量。

## 8. 阶段计划与验收

| 阶段 | 内容 | 验收标准 |
|---|---|---|
| Phase 0 | agent 服务骨架、healthz、SSE 回显、compose/nginx 接入 | 浏览器能看到流式回显事件 |
| Phase 1 | Pi SDK 接入、`list_tasks` / `list_notes` / `get_daily_stats`、聊天面板 | “今天有哪些任务”能正确查询并流式回答 |
| Phase 2 | 写工具、两阶段审批、审计 | 创建/修改/归档需确认；审计可查；删除需二次确认 |
| Phase 3 | 会话持久化、限额、eval 集（15~20 条）、README | 重启后可恢复会话；eval 全绿；文档完整 |

## 9. 风险与边界

- **Pi SDK 年轻**：API 可能变动，锁版本、读源码、封装一层自己的 `AgentRuntime` 接口隔离。
- **上下文爆炸**：工具返回必须裁剪（条数上限、字段精简、长文本截断）；后续再做摘要压缩。
- **多用户并发**：thread 级串行是硬要求，否则会话状态会互相覆盖。
- **安全**：JWT 只在内存；写操作审批；提示词不是安全边界；工具返回内容（用户笔记）按数据处理，防注入。
- **成本**：按 run 计 token 预算，限制最大步数，日志记录 usage。

## 10. 前置知识清单

按优先级：

1. **Agent 原理**：循环与终止条件、tool calling 机制（schema、tool_call_id、结果回填）、消息角色、上下文预算与压缩、流式事件时序。
2. **Pi 用法**：`pi-ai` 事件（`start` / `text_*` / `toolcall_*` / `done` / `error`）、`createAgentSession` / `DefaultResourceLoader` / `customTools` / `SessionManager`；重点读源码 `agent-loop`、`stream-fn`、`AgentTool`。
3. **服务端工程**：SSE、AG-UI 事件、Redis（会话/锁/去重）、同 thread 串行、Docker 化 Node 服务。
4. **安全**：用户委派 vs 服务身份、最小权限、两阶段审批、审计、prompt injection。
5. **工具与模型**：工具粒度与描述质量、OpenAI 兼容接口的流式 tool_calls、按 run 算成本。
6. **评测**：固定用例集、trace、失败定位（没调工具 / 调错工具 / 循环打转）。

当前不需要学：LangGraph、多 agent 编排、RAG/向量库、K8s、MCP Server 开发。

## 11. 参考资料

- Pi SDK 接入示例：https://m.toutiao.com/article/7673696349073818147/
- Pi 生产级 Harness 解析（pi-ai stream / Agent Loop）：https://jishuzhan.net/article/2100533957306273793
- Pi Agent 架构与关键功能全解析：https://jishuzhan.net/article/2093954607039565825
- Pi 源码与架构深度分析：https://m.blog.csdn.net/m0_58426858/article/details/164189556
- AG-UI 1.0 事件流规范解读：https://blog.csdn.net/m0_61531676/article/details/170247650
- AG-UI 官方仓库：https://github.com/ag-ui-protocol/ag-ui
- 火山方舟 Agents 工具与权限策略（custom tools 参考）：https://docs.volcengine.com/docs/ark/tools?lang=zh

## 12. 实施进度（2026-11-12）

**已完成（Phase 0/1）**

- `apps/agent`：Hono 服务、Pi SDK 运行时（`createAgentSession` + `ModelRuntime` + `DefaultResourceLoader`）、只读与写操作工具、AG-UI 1.0 事件流（官方 core/encoder + schema 校验）、同 thread 串行锁、按 userID + thread 隔离会话。
- 写操作（Phase 2 部分）：`create_task` / `update_task` / `complete_task` / `get_task`，支持一句话创建任务、按关键词完成任务；重名任务返回候选让用户确认；删除与归档暂不开放。
- 用户自定义模型（BYOK）：悬浮窗「模型设置」入口、pi-ai 模型目录下拉、AES-256-GCM 加密存储（Redis / 内存回退）、按用户切换 provider / model / Key；未配置时回落到平台默认并给出配置引导。
- 前端：`AgentChatWindow.vue` 悬浮窗（自绘消息流 + 输入框，Element Plus + 主题变量），点击看板 3D 形象打开，可拖动、Esc 关闭，流式文本与工具步骤渲染。
- 部署：agent `Dockerfile`、dev/prod compose 服务、nginx SSE 反代、`prod.sh` 构建与上传 agent 镜像。
- 安全：开跑前 `GET /api/auth/me` 预检（无效 token 401 且不消耗模型，Go 不可用 502）；错误信息脱敏（用户只看到中文提示，完整错误仅写服务端日志）；JWT 只在单次 run 内存中透传。

**已验证**

- 前后端 type-check 与生产构建通过。
- SSE：`RUN_STARTED` → `RUN_ERROR`（无 Key 时友好提示）；未登录返回 401。
- 鉴权预检：无效 token 直接 401（无 SSE、不进入模型）；Go API 不可用返回 502；有效 token 正常进入 SSE。
- 工具层（mock Go API）：`list_tasks` 归档过滤与字段精简、`get_daily_stats` 统计、`list_notes`、401 透传全部正确；写工具创建/详情/修改/按关键词完成/重复完成/重名歧义确认全部符合预期。
- 模型设置：默认配置、保存（Key 掩码）、省略 Key 保留旧值、清除回落默认、无效模型 400、未登录 401；AES-GCM 加密往返正确、错误密钥被拒绝；运行时日志确认按用户选择 `deepseek/deepseek-flash (custom)`，假 Key 返回「模型服务鉴权失败」提示。
- AG-UI：映射层 12 类事件全部通过官方 `EventSchema` 校验；SSE 帧为纯 `data:` JSON；浏览器端按 messageId / toolCallId 消费正常。
- 前端：模型设置界面类型检查通过（provider/model 下拉、Key 输入、保存/清除）。
- 浏览器：看板加载、点击 3D 形象打开悬浮窗、拖动定位、窗内发送消息、SSE 错误态渲染正常、发送后输入框自动清空；模型设置面板可打开并显示平台默认配置，通过界面保存 Key 后重新打开显示 `****5678` 掩码。

**待验证（需要真实环境）**

- 真实 `DEEPSEEK_API_KEY` 或其他真实模型 Key 下的完整对话与工具调用链路（当前假 Key 只验证了鉴权失败路径）。
- Docker compose 构建与部署（本机 Docker 未运行）。

**本地运行**

```bash
# Go API + MySQL + Redis（已有 docker compose）
docker compose up -d db cache api

# agent（填好 apps/agent/.env 中的 DEEPSEEK_API_KEY）
cd apps/agent
cp .env.example .env
npm install
npm run dev

# web（开发代理已配置 /api 与 /agent）
cd apps/web
npm run dev
```

**下一步（Phase 2 剩余）**

- 写操作审批（当前按用户要求为直接执行）与 Go 侧审计端点；
- 便签写工具与删除/归档操作（需审批后开放）。

