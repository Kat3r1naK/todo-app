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
| 部署 | Docker Compose(开发 Vite HMR + air 热重载 / 生产:本地多阶段构建 + SSH 分发) |

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
├── scripts/prod.sh             # 本地构建 + SSH 上传部署
├── docker-compose.yml          # 开发环境
├── docker-compose.prod.yml     # 生产环境(引用本地构建的镜像)
├── Makefile
├── .env.deploy.example         # 部署目标配置示例
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

## 生产部署(本地构建 + SSH 上传)

镜像在本地构建,通过 SSH 传到服务器运行。服务器只需要 Docker 和 Compose 插件,不需要 Go/Node/make,也不需要拉源码;本地需要 Docker 以及 ssh/scp/gzip(Git Bash 自带)。

服务器首次准备:

```bash
ssh root@<服务器> 'mkdir -p /srv/todo-app'
```

然后在服务器的 `/srv/todo-app/.env.prod` 中写入生产配置(参考 `.env.example`),至少修改数据库密码、Redis 密码和 `JWT_SECRET`(可用 `openssl rand -hex 32` 生成),并放行 `WEB_PORT_PROD`(默认 8088)。

本地配置并发布:

```bash
cp .env.deploy.example .env.deploy
# 填写 REMOTE=user@host、REMOTE_DIR=/srv/todo-app

bash ./scripts/prod.sh deploy    # 构建镜像 → SSH 上传 → 远端启动(等价 make prod-deploy)
```

镜像以当前 git 提交的短 SHA 为标签,只上传 api / web 两个镜像;MySQL 和 Redis 由服务器自行拉取,数据持久化在 Docker Volume 中。

日常运维:

```bash
bash ./scripts/prod.sh logs      # 远端日志
bash ./scripts/prod.sh ps        # 容器状态
bash ./scripts/prod.sh down      # 停止远端服务
```

更新发布重复执行 `bash ./scripts/prod.sh deploy` 即可。公网访问建议前置 Caddy / 云负载均衡提供 HTTPS;旧镜像会留在服务器上,可按需用 `docker image prune` 清理。

## 安全提示

- 请勿提交 `.env` / `.env.prod` / `.env.deploy`,仓库已默认忽略
- 上线后立即修改默认管理员密码,并使用强密码与独立 JWT 密钥
- 公网访问建议使用 HTTPS(域名 + Caddy / 云负载均衡),并限制安全组来源 IP
