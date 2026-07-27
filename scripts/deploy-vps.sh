#!/bin/bash

# ============================================
# GBS POS-CMS Auto Deploy Script
# ============================================
# Local: Build & push Docker images ke GHCR
# VPS:   Pull & restart containers
#
# Usage: ./deploy.sh [pos|cms|all] [production]
# Default: deploy all services ke production
# ============================================

set -e

# === Configuration ===
REGISTRY="ghcr.io/fariziadam11"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# VPS Configuration (sesuaikan dengan server kamu)
VPS_HOST="159.89.204.100"
VPS_USER="adam"
VPS_PATH="/opt/gbs/gbs-pos-cms-api"
VPS_ENV_FILE="/opt/gbs/.env"

# === Colors ===
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# === Functions ===
log_info()    { echo -e "${BLUE}[INFO]${NC}    $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARN]${NC}    $1"; }
log_error()   { echo -e "${RED}[ERROR]${NC}   $1"; }
log_step()    { echo -e "${CYAN}[STEP]${NC}     $1"; }

check_vps_config() {
    if [ -z "$VPS_HOST" ]; then
        log_error "VPS_HOST belum diset!"
        echo "Set variabel environment atau edit script ini:"
        echo "  export VPS_HOST=your-server-ip"
        echo "  export VPS_USER=root  # atau user lain"
        exit 1
    fi
}

# === Parse Arguments ===
SERVICE="${1:-all}"
ENV="${2:-production}"

case "$SERVICE" in
    pos)    SERVICES=("pos") ;;
    cms)    SERVICES=("cms") ;;
    all)    SERVICES=("pos" "cms") ;;
    *)      echo "Usage: $0 [pos|cms|all] [production]"
            exit 1 ;;
esac

# === Main ===
log_info "=========================================="
log_info "  GBS POS-CMS Deploy Script"
log_info "=========================================="
log_info "Service : ${SERVICES[*]}"
log_info "Target  : $VPS_USER@$VPS_HOST"
log_info "Path    : $VPS_PATH"
log_info "=========================================="
echo

# ----------------------------
# 1. Build & Push Images
# ----------------------------
log_step "Building Docker images..."

for svc in "${SERVICES[@]}"; do
    log_info "Building gbs-$svc-api..."

    case "$svc" in
        pos) DOCKERFILE="gbs-pos-api/Dockerfile" ;;
        cms) DOCKERFILE="gbs-cms-api/Dockerfile" ;;
    esac

    IMAGE_TAG="$REGISTRY/gbs-$svc-api:latest"

    docker build -t "$IMAGE_TAG" -f "$DOCKERFILE" .
    log_success "Built: $IMAGE_TAG"

    docker push "$IMAGE_TAG"
    log_success "Pushed: $IMAGE_TAG"
    echo
done

# ----------------------------
# 2. Deploy ke VPS
# ----------------------------
check_vps_config

log_step "Deploying to VPS..."

ssh "$VPS_USER@$VPS_HOST" << 'EOF'
    set -e
    sudo su -

    cd $VPS_PATH

    echo "[INFO] Pulling latest images..."
    docker compose -f docker-compose.prod.yml --env-file $VPS_ENV_FILE pull

    echo "[INFO] Restarting containers..."
    docker compose -f docker-compose.prod.yml --env-file $VPS_ENV_FILE up -d --force-recreate

    echo "[INFO] Cleaning unused images..."
    docker image prune -f

    echo "[INFO] Container status:"
    docker compose -f docker-compose.prod.yml --env-file $VPS_ENV_FILE ps
EOF

# ----------------------------
# 3. Health Check
# ----------------------------
log_step "Running health check..."

sleep 5

HEALTHY=true
for svc in "${SERVICES[@]}"; do
    case "$svc" in
        pos) PORT=8080 ;;
        cms) PORT=8081 ;;
    esac

    if curl -sf "http://$VPS_HOST:$PORT/health" > /dev/null 2>&1; then
        log_success "gbs-$svc-api is healthy"
    else
        log_error "gbs-$svc-api health check FAILED"
        HEALTHY=false
    fi
done

echo
if [ "$HEALTHY" = true ]; then
    log_success "=========================================="
    log_success "  Deploy completed successfully! 🚀"
    log_success "=========================================="
else
    log_error "Some services failed health check"
    exit 1
fi
