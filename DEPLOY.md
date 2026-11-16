# 部署文档（todo-app）

前后端分离待办看板的生产部署说明。采用 **本地构建镜像 → SSH 上传 → 服务器只运行容器** 的模式，服务器不需要 Go/Node/make，也不需要拉源码。

## 1. 部署架构

```
本地 Windows (Docker Desktop)
  ├─ docker build  →  todo-app-api:<tag> / todo-app-web:<tag>
  ├─ docker save | gzip | ssh docker load   （只传 api / web / agent 三个镜像）
  └─ scp docker-compose.prod.yml + ssh compose up -d
            │
            ▼
云服务器 (Ubuntu, Docker + Compose)
  └─ docker compose 运行 5 个容器：
       web   nginx:1.27      0.0.0.0:8088 → 80   （唯一对外端口）
       api   Go 服务         8080（仅内网，经 nginx /api 反代）
       agent Node agent 服务 3001（仅内网，经 nginx /agent 反代，SSE）
       db    MySQL 8.4       3306（仅内网，数据卷 db_data）
       cache Redis 7         6379（仅内网，数据卷 redis_data）
```

- 前端请求 `/api/*` 由 nginx 反向代理到 `api:8080`，其余路径返回 SPA 静态资源。
- MySQL / Redis / API 都不对公网暴露，只有 web 的 8088 端口对外。
- 镜像标签默认取当前 git 提交短 SHA（`git rev-parse --short HEAD`）。

## 2. 环境信息

| 项 | 值 |
| --- | --- |
| 服务器 | `<服务器IP>`（腾讯云 Ubuntu） |
| SSH 用户 | `ubuntu`（已配免密登录，别名 `todo-prod`） |
| 部署目录 | `/srv/todo-app` |
| 服务器配置 | `/srv/todo-app/.env.prod`（权限 600，随机强密码，**不在仓库中**） |
| 对外端口 | `8088`（`WEB_PORT_PROD`） |
| 本地部署配置 | `.env.deploy`（已被 git 忽略） |

## 3. 前置条件

**本地：**
- Docker Desktop（已启动）
- Git Bash（自带 `ssh` / `scp` / `gzip`）
- 国内网络需解决 Docker Hub 直连超时（见第 8 节）

**服务器：**
- Docker + Compose v2 插件（`docker compose version` 可用）
- `ubuntu` 用户在 `docker` 组，且 sudo 免密
- 已创建 `/srv/todo-app/.env.prod`

## 4. 首次部署（已完成，记录可复现步骤）

```bash
# 1) 服务器免密登录（本地公钥装到服务器）
ssh-copy-id -i ~/.ssh/id_ed25519.pub ubuntu@<服务器IP>

# 2) 服务器创建部署目录与生产配置（密码随机生成）
ssh todo-prod 'sudo mkdir -p /srv/todo-app && sudo chown ubuntu:ubuntu /srv/todo-app'
ssh todo-prod 'cd /srv/todo-app && umask 077 && cat > .env.prod <<EOF
MYSQL_ROOT_PASSWORD=$(openssl rand -hex 16)
MYSQL_DATABASE=todo
MYSQL_USER=todo
MYSQL_PASSWORD=$(openssl rand -hex 16)
REDIS_PASSWORD=$(openssl rand -hex 16)
JWT_SECRET=$(openssl rand -hex 32)
WEB_PORT_PROD=8088
EOF'

# 3) 本地部署配置
cp .env.deploy.example .env.deploy   # 填 REMOTE / REMOTE_DIR / PLATFORM

# 4) 构建 + 上传 + 启动
bash ./scripts/prod.sh deploy
```

`.env.prod` 各变量含义见 `.env.example`；生产 compose 只用到
`MYSQL_ROOT_PASSWORD / MYSQL_DATABASE / MYSQL_USER / MYSQL_PASSWORD / REDIS_PASSWORD / JWT_SECRET / WEB_PORT_PROD / DEEPSEEK_API_KEY / MODEL_PROVIDER / MODEL_ID / MODEL_SETTINGS_SECRET`。

> 部署 agent 前，需在服务器 `/srv/todo-app/.env.prod` 中补充 `DEEPSEEK_API_KEY`（平台默认模型，可选）与 `MODEL_SETTINGS_SECRET`（用户自定义模型 Key 的加密密钥，必填，建议 `openssl rand -hex 32`）；可选 `MODEL_PROVIDER`（默认 deepseek）与 `MODEL_ID`（默认 deepseek-v4-pro）。

## 5. 日常更新发布

改完代码后重复执行：

```bash
bash ./scripts/prod.sh deploy
```

脚本会重新构建镜像（新 SHA 标签）→ 上传 → 远端 `compose up -d` 自动重建变更的容器。
数据保存在 `db_data` / `redis_data` 卷中，重建容器不会丢数据。

## 6. 运维命令

```bash
bash ./scripts/prod.sh ps      # 容器状态（应均为 healthy / Up）
bash ./scripts/prod.sh logs    # 远端 api / web 日志（Ctrl+C 退出）
bash ./scripts/prod.sh down    # 停止远端服务（保留数据卷）
bash ./scripts/prod.sh build   # 只本地构建，不上传
```

等价 Makefile 目标（需本机有 make）：`make prod-deploy / prod-ps / prod-logs / prod-down`。

## 7. 云安全组端口清单（重要）

在腾讯云控制台 → 轻量/CVM → 安全组（或防火墙）中配置：

| 端口 | 协议 | 用途 | 是否放行 | 建议来源 |
| --- | --- | --- | --- | --- |
| **8088** | TCP | Web 访问（当前唯一对外服务） | **必须放行** | `0.0.0.0/0`（或限定你的 IP） |
| 22 | TCP | SSH 部署 / 运维 | 已放行 | 建议限定你的 IP |
| 80 | TCP | HTTP（以后加域名 / Caddy 时） | 可选 | `0.0.0.0/0` |
| 443 | TCP | HTTPS（以后加域名 / Caddy 时） | 可选 | `0.0.0.0/0` |
| 3306 | TCP | MySQL | **禁止放行** | 仅容器内网 |
| 6379 | TCP | Redis | **禁止放行** | 仅容器内网 |
| 8080 | TCP | Go API | **禁止放行** | 仅容器内网（经 nginx 反代） |

> 服务器自身 `ufw` 当前为 inactive，无需额外配置；公网能否访问只取决于云安全组。
> 放行 8088 后访问：`http://<服务器IP>:8088`

## 8. 常见问题

**Q1：本地构建报 `registry-1.docker.io ... connection timed out`**
国内直连 Docker Hub 超时。本项目已用 DaoCloud 镜像源预拉基础镜像并打本地标签绕过：

```bash
M=docker.m.daocloud.io/library
for img in golang:1.26-alpine node:22-alpine nginx:1.27-alpine alpine:3.20; do
  docker pull "$M/$img" && docker tag "$M/$img" "$img"
done
```

同时两个 Dockerfile 已去掉 `# syntax=docker/dockerfile:1` 行（该行会强制从 Docker Hub 拉前端镜像）。
基础镜像拉过一次后本地有缓存，后续构建不再需要联网拉取。

**Q2：公网打不开，但服务器本地 `curl 127.0.0.1:8088` 正常**
云安全组没放行 8088，按第 7 节放行即可。

**Q3：`/api` 返回 401**
正常，接口需要登录鉴权，说明 nginx → api 反代已通。

**Q4：前端构建 npm 慢 / 失败**
可临时给 web 构建加国内 npm 源（`registry.npmmirror.com`）；当前官方源可达，未启用。

## 9. 验证

```bash
# 服务器本地
ssh todo-prod 'curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8088/'        # 期望 200
ssh todo-prod 'curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8088/api/tasks' # 期望 401

# 公网（放行安全组后）
curl -I http://<服务器IP>:8088/
```

浏览器打开 `http://<服务器IP>:8088`，默认管理员 `admin / admin`，**登录后立即修改密码**。

## 10. 安全注意事项

- 默认管理员 `admin / admin` 上线后必须立即改密。
- 服务器登录密码曾在部署对话中出现，建议尽快修改，或改用纯密钥登录并禁用密码登录。
- `.env` / `.env.prod` / `.env.deploy` 均已被 git 忽略，切勿提交；`.env.prod` 权限保持 600。
- 公网建议前置 Caddy / 云负载均衡提供 HTTPS（放行 80/443），不要长期裸 HTTP。
- 定期备份数据库：

```bash
ssh todo-prod 'cd /srv/todo-app && docker compose --env-file .env.prod -f docker-compose.prod.yml \
  exec -T db sh -c "mysqldump -uroot -p\$MYSQL_ROOT_PASSWORD \$MYSQL_DATABASE" > backup-$(date +%F).sql'
```

## 11. 回滚

镜像按 git SHA 打标签，旧镜像仍留在服务器。回滚到上一版本：

```bash
ssh todo-prod 'cd /srv/todo-app && IMAGE_TAG=<旧SHA> docker compose --env-file .env.prod -f docker-compose.prod.yml up -d'
```

旧镜像可用 `ssh todo-prod 'docker images | grep todo-app'` 查看；清理无用镜像用 `docker image prune`（谨慎）。
