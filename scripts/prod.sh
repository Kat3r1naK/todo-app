#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [ -f .env.deploy ]; then
  set -a
  # shellcheck disable=SC1091
  . ./.env.deploy
  set +a
fi

REMOTE="${REMOTE:-}"
REMOTE_DIR="${REMOTE_DIR:-/srv/todo-app}"
PLATFORM="${PLATFORM:-linux/amd64}"
IMAGE_TAG="${IMAGE_TAG:-$(git rev-parse --short HEAD)}"
API_IMAGE="todo-app-api:${IMAGE_TAG}"
WEB_IMAGE="todo-app-web:${IMAGE_TAG}"
AGENT_IMAGE="todo-app-agent:${IMAGE_TAG}"
COMPOSE="docker compose --env-file .env.prod -f docker-compose.prod.yml"

require_remote() {
  if [ -z "$REMOTE" ]; then
    echo "请先执行 cp .env.deploy.example .env.deploy 并填写 REMOTE=user@host" >&2
    exit 1
  fi
}

build() {
  echo "==> 本地构建 ${API_IMAGE} / ${WEB_IMAGE} / ${AGENT_IMAGE} (${PLATFORM})"
  docker build --platform "$PLATFORM" --target prod -t "$API_IMAGE" ./apps/server
  docker build --platform "$PLATFORM" --target prod -t "$WEB_IMAGE" ./apps/web
  docker build --platform "$PLATFORM" -t "$AGENT_IMAGE" ./apps/agent
}

deploy() {
  require_remote
  echo "==> 检查远端目录 ${REMOTE}:${REMOTE_DIR}"
  if ! ssh "$REMOTE" "mkdir -p '$REMOTE_DIR' && test -f '$REMOTE_DIR/.env.prod'"; then
    echo "远端缺少 $REMOTE_DIR/.env.prod，首次部署请先在服务器上创建" >&2
    exit 1
  fi
  build
  scp docker-compose.prod.yml "$REMOTE:$REMOTE_DIR/docker-compose.prod.yml"
  echo "==> 传输镜像"
  docker save "$API_IMAGE" "$WEB_IMAGE" "$AGENT_IMAGE" | gzip | ssh "$REMOTE" 'gunzip | docker load'
  echo "==> 远端启动"
  ssh "$REMOTE" "cd '$REMOTE_DIR' && IMAGE_TAG='$IMAGE_TAG' $COMPOSE up -d"
  echo "==> 部署完成"
}

remote_compose() {
  require_remote
  ssh "$REMOTE" "cd '$REMOTE_DIR' && IMAGE_TAG='$IMAGE_TAG' $COMPOSE $1"
}

case "${1:-}" in
  build) build ;;
  deploy) deploy ;;
  logs) remote_compose "logs -f --tail=100 api web agent" ;;
  ps) remote_compose "ps" ;;
  down) remote_compose "down" ;;
  *)
    echo "用法: bash ./scripts/prod.sh {deploy|build|logs|ps|down}" >&2
    exit 1
    ;;
esac
