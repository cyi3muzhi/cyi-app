#!/usr/bin/env bash
# ====================================================================
# 本地开发一键环境准备（Linux / macOS 版）
#
# 功能：
#   1. 检查 Docker，必要时从 .env.example 生成 .env
#   2. 启动基础设施（MySQL / Redis / Consul / NATS / Jaeger / RustFS 等）
#   3. 等待 Consul 就绪
#   4. 把各服务 configs/*.yaml 转成本地地址与端口后写入 Consul KV
#
# 用法：
#   bash scripts/dev-setup.sh
#   bash scripts/dev-setup.sh --skip-infra
#   bash scripts/dev-setup.sh --consul=127.0.0.1:8500
#
# 说明：
#   - Consul 由 Docker 提供，本机无需安装
#   - 重复执行是安全的，配置会覆盖写入
# ====================================================================

set -euo pipefail

SKIP_INFRA=0
CONSUL_ADDR="${CONSUL_ADDR:-127.0.0.1:8500}"
for arg in "$@"; do
  case "$arg" in
    --skip-infra) SKIP_INFRA=1 ;;
    --consul=*)   CONSUL_ADDR="${arg#*=}" ;;
  esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_DIR="$PROJECT_ROOT/deploy/docker-compose"
APP_DIR="$PROJECT_ROOT/app"
ENV_FILE="$COMPOSE_DIR/.env"

step() { printf "\n==> %s\n" "$1"; }
ok()   { printf "  [OK] %s\n" "$1"; }
warn() { printf "  [..] %s\n" "$1"; }
err()  { printf "  [!!] %s\n" "$1" >&2; }

# --------------------------------------------------------------------
# 1. 检查 Docker 与 Compose
# --------------------------------------------------------------------
step "检查 Docker"
if ! command -v docker >/dev/null 2>&1; then
  err "未安装 Docker，请先安装 Docker Desktop 或 Docker Engine"
  exit 1
fi
if ! docker version >/dev/null 2>&1; then
  err "Docker 未运行，请先启动 Docker"
  exit 1
fi
if docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE=(docker-compose)
else
  err "未找到 docker compose / docker-compose"
  exit 1
fi
ok "Docker 正常"

# --------------------------------------------------------------------
# 2. 准备 .env
# --------------------------------------------------------------------
if [ ! -f "$ENV_FILE" ]; then
  cp "$COMPOSE_DIR/.env.example" "$ENV_FILE"
  ok "已从 .env.example 生成 .env"
fi

# 从 .env 读取端口，读不到时使用默认值
port() {
  local name="$1" default="$2" val
  val="$(grep -E "^${name}=" "$ENV_FILE" 2>/dev/null | tail -1 | cut -d= -f2- | tr -d '\r' || true)"
  if [ -n "${val:-}" ]; then printf '%s' "$val"; else printf '%s' "$default"; fi
}

# --------------------------------------------------------------------
# 3. 启动基础设施
# --------------------------------------------------------------------
if [ "$SKIP_INFRA" -eq 0 ]; then
  step "启动基础设施（首次拉取镜像可能较慢）"
  ( cd "$COMPOSE_DIR" && "${COMPOSE[@]}" -f docker-compose.infra.yml --env-file .env -p cyi-app up -d )
  ok "基础设施已启动"
else
  warn "已跳过基础设施启动"
fi

# --------------------------------------------------------------------
# 4. 等待 Consul 就绪
# --------------------------------------------------------------------
step "等待 Consul 就绪"
ready=0
for _ in $(seq 1 30); do
  if curl -sf "http://$CONSUL_ADDR/v1/status/leader" >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 2
done
if [ "$ready" -ne 1 ]; then
  err "Consul 未就绪，请检查：docker compose -f deploy/docker-compose/docker-compose.infra.yml ps"
  exit 1
fi
ok "Consul 已就绪：http://$CONSUL_ADDR"

# --------------------------------------------------------------------
# 5. 写入公共配置（launch.json 会读取 /local/common）
# --------------------------------------------------------------------
put_config() {
  local key="$1" content="$2"
  printf '%s' "$content" | curl -sf -X PUT --data-binary @- "http://$CONSUL_ADDR/v1/kv/$key" >/dev/null
}

put_config "local/common/config.yaml" "{}"
ok "已写入公共配置 local/common/config.yaml"

# --------------------------------------------------------------------
# 6. 端口映射（容器内端口 -> 本机端口），需与 .env 保持一致
# --------------------------------------------------------------------
map_port() {
  local rel="$1" src="$2"
  case "$rel:$src" in
    "cin/interface:8000")     port INTERFACE_HTTP_PORT 8010 ;;
    "cin/interface:9000")     port INTERFACE_GRPC_PORT 9010 ;;
    "cin/admin:8001")         port ADMIN_HTTP_PORT 8080 ;;
    "cin/admin:9001")         port ADMIN_GRPC_PORT 9091 ;;
    "service/user:8000")      printf '8100' ;;
    "service/user:9000")      port USER_GRPC_PORT 9100 ;;
    "service/order:8003")     port ORDER_HTTP_PORT 9104 ;;
    "service/order:9003")     port ORDER_GRPC_PORT 9101 ;;
    "service/product:8000")   printf '8102' ;;
    "service/product:9000")   port PRODUCT_GRPC_PORT 9102 ;;
    "service/marketing:8000") printf '8103' ;;
    "service/marketing:9000") port MARKETING_GRPC_PORT 9103 ;;
    "service/chat:8005")      printf '8105' ;;
    "service/chat:9005")      port CHAT_GRPC_PORT 9105 ;;
    "blades:8000")            port BLADES_HTTP_PORT 30800 ;;
    "blades:9000")            port BLADES_GRPC_PORT 30900 ;;
    "task:8000")              port TASK_PORT 10100 ;;
    "task:9000")              printf '10101' ;;
    *) printf '%s' "$src" ;;
  esac
}

# --------------------------------------------------------------------
# 7. 配置内容转换
#    host 先换占位符，再换端口，最后还原真实地址，避免误伤 OSS 的 9000
# --------------------------------------------------------------------
transform() {
  local rel="$1" content="$2" src dst
  local h_consul h_mysql h_redis h_nats h_jaeger h_rustfs h_dtm h_blades
  h_consul="127.0.0.1:$(port CONSUL_PORT 8500)"
  h_mysql="127.0.0.1:$(port MYSQL_PORT 3306)"
  h_redis="127.0.0.1:$(port REDIS_PORT 6379)"
  h_nats="127.0.0.1:$(port NATS_PORT 4222)"
  h_jaeger="127.0.0.1:$(port JAEGER_OTLP_GRPC_PORT 4318)"
  h_rustfs="127.0.0.1:$(port RUSTFS_API_PORT 9001)"
  h_dtm="http://127.0.0.1:$(port DTM_PORT 36789)"
  h_blades="127.0.0.1:$(port BLADES_HTTP_PORT 30800)"

  content="${content//consul:8500/__HOST_CONSUL__}"
  content="${content//mysql:3306/__HOST_MYSQL__}"
  content="${content//redis:6379/__HOST_REDIS__}"
  content="${content//nats:4222/__HOST_NATS__}"
  content="${content//jaeger:4318/__HOST_JAEGER__}"
  content="${content//rustfs:9000/__HOST_RUSTFS__}"
  content="${content//http:\/\/dtm:36789/__HOST_DTM__}"
  content="${content//blades:8000/__HOST_BLADES__}"

  for src in 8000 8001 8003 8005 9000 9001 9003 9005; do
    dst="$(map_port "$rel" "$src")"
    if [ "$dst" != "$src" ]; then
      content="${content//:$src/:$dst}"
    fi
  done

  content="${content//__HOST_CONSUL__/$h_consul}"
  content="${content//__HOST_MYSQL__/$h_mysql}"
  content="${content//__HOST_REDIS__/$h_redis}"
  content="${content//__HOST_NATS__/$h_nats}"
  content="${content//__HOST_JAEGER__/$h_jaeger}"
  content="${content//__HOST_RUSTFS__/$h_rustfs}"
  content="${content//__HOST_DTM__/$h_dtm}"
  content="${content//__HOST_BLADES__/$h_blades}"

  printf '%s' "$content"
}

# --------------------------------------------------------------------
# 8. 遍历服务配置并写入 Consul
# --------------------------------------------------------------------
step "导入服务配置到 Consul KV"

count=0
while IFS= read -r -d '' file; do
  dir="$(dirname "$file")"
  case "$dir" in
    */configs) ;;
    *) continue ;;
  esac
  base="$(basename "$file")"
  case "$base" in
    config.yaml|auth.yaml|register.yaml|transaction.yaml) ;;
    *) continue ;;
  esac

  rel="${dir#"$APP_DIR"/}"
  content="$(cat "$file")"
  content="$(transform "$rel" "$content")"
  key="local/cyi-app/$rel/$base"
  put_config "$key" "$content"
  ok "$key"
  count=$((count + 1))
done < <(find "$APP_DIR" -type f -name "*.yaml" -print0)

printf "\n配置导入完成，共 %d 个文件。\n" "$count"
printf "\n下一步：\n"
printf "  1. 在 VS Code 打开「运行和调试」，选择 user_service 等配置按 F5 启动\n"
printf "  2. Consul 控制台：http://%s/ui\n" "$CONSUL_ADDR"
printf "  3. 若服务启动报配置缺失，重新执行本脚本即可\n\n"
printf "提示：生产部署请勿使用本脚本，容器内使用服务名访问基础设施。\n"
