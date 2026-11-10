# todo-app

一个前后端分离的待办看板应用:支持多账号数据隔离、拖拽排序、便签和主题切换。使用 monorepo 管理,基于 Docker 开发与部署。

## 功能

- 账号登录(JWT,7 天会话)与数据隔离,支持免登链接
- 看板三列(待办 / 进行中 / 已完成),卡片一键勾选完成、子任务清单、跨列拖拽排序、置顶、优先级、搜索
- 任务支持子任务,卡片上可直接逐条勾选并显示完成进度(如 2/3)
- 每日统计:今日完成 / 新增数量与近 7 天完成趋势
- 任务备注支持 Markdown 编辑与预览,方便沉淀排查过程和解决方案
- 已完成任务可归档,归档抽屉支持搜索、查看备注、恢复到看板
- 便签抽屉(5 种颜色)与 5 套界面主题(深色 / 暖阳 / 薄荷 / 樱花 / 纸墨)
- 管理端:创建账号、重置密码、删除账号、任务数统计
- 响应式布局:桌面三列并排,移动端按状态分栏切换

## 技术栈

| 层 | 技术 |
| --- | --- |
| 前端 | Vue 3 + TypeScript + Vite + Element Plus + Pinia + Vue Router + vuedraggable |
| 后端 | Go + Gin + GORM + Wire |
| 存储 | MySQL 8 + Redis 7 |
| 部署 | Docker Compose(开发 Vite HMR + air 热重载 / 生产 nginx + 多阶段镜像) |

## 目录结构

```
todo-app/
├── apps/
│   ├── server/                 # Go 服务(monorepo 后端)
│   │   ├── config/             # 配置加载
│   │   ├── internal/
│   │   │   ├── domain/         # 领域模型
│   │   │   ├── errs/           # 业务错误
│   │   │   ├── repository/     # 仓储层(dao + cache)
│   │   │   ├── service/        # 业务逻辑
│   │   │   └── web/            # Gin 路由 / 中间件 / Handler
│   │   ├── ioc/                # Wire 依赖注入
│   │   └── pkg/                # jwt / password 等通用包
│   └── web/                    # Vue 3 前端
│       └── src/
│           ├── components/      # 看板、任务、便签、账号管理
│           ├── stores/          # Pinia
│           ├── views/           # 登录 / 看板
│           └── types.ts         # 领域与接口类型
├── docker-compose.yml          # 开发环境
├── docker-compose.prod.yml     # 生产环境
├── Makefile
└── .env.example
```

## 本地开发(Docker)

```bash
cp .env.example .env
make dev-build
```

- 前端:http://localhost:5173
- API:http://localhost:8080/healthz
- 默认管理员:`admin / admin`(登录后请立即修改密码)

前端支持 Vite HMR,后端使用 air 热重载,MySQL 和 Redis 数据通过 Docker Volume 持久化。

## 生产部署

```bash
cp .env.example .env.prod
# 修改 .env.prod 中的数据库密码、Redis 密码、JWT_SECRET
# 可用 openssl rand -hex 32 生成 JWT_SECRET
make down
COMPOSE_PARALLEL_LIMIT=1 make prod
```

生产档只暴露 web 端口(默认 `8088:80`),nginx 托管静态资源并将 `/api` 反向代理到 Go 服务;MySQL 与 Redis 仅在容器内网中访问。

## 安全提示

- 请勿提交 `.env` / `.env.prod`,仓库已默认忽略
- 上线后立即修改默认管理员密码,并使用强密码与独立 JWT 密钥
- 公网访问建议使用 HTTPS(域名 + Caddy / 云负载均衡),并限制安全组来源 IP
