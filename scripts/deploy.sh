#!/bin/bash

# GBS POS-CMS API Deployment Script
# Usage: ./scripts/deploy.sh [staging|production]

set -e

ENVIRONMENT=${1:-staging}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_requirements() {
    log_info "Checking requirements..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed"
        exit 1
    fi
    
    log_success "All requirements met"
}

backup_database() {
    log_info "Creating database backup..."
    
    BACKUP_DIR="$PROJECT_DIR/backups"
    mkdir -p "$BACKUP_DIR"
    
    BACKUP_FILE="$BACKUP_DIR/gbs_pos_$(date +%Y%m%d_%H%M%S).sql"
    
    docker exec gbs-pos-cms-api-postgres-1 pg_dump -U postgres gbs_pos > "$BACKUP_FILE"
    
    if [ $? -eq 0 ]; then
        log_success "Database backup created: $BACKUP_FILE"
    else
        log_error "Database backup failed"
        exit 1
    fi
}

pull_latest_code() {
    log_info "Pulling latest code from Git..."
    
    cd "$PROJECT_DIR"
    
    if [ "$ENVIRONMENT" == "production" ]; then
        git pull origin main
    else
        git pull origin develop
    fi
    
    log_success "Code updated"
}

build_images() {
    log_info "Building Docker images..."
    
    cd "$PROJECT_DIR"
    docker-compose build --no-cache
    
    log_success "Images built successfully"
}

run_migrations() {
    log_info "Running database migrations..."
    
    # Migrations are handled automatically by the application
    log_success "Migrations will run on container startup"
}

deploy_containers() {
    log_info "Deploying containers..."
    
    cd "$PROJECT_DIR"
    docker-compose up -d --force-recreate
    
    log_success "Containers deployed"
}

health_check() {
    log_info "Running health checks..."
    
    sleep 10
    
    # Check POS API
    if curl -f http://localhost:8080/health &> /dev/null; then
        log_success "POS API is healthy"
    else
        log_error "POS API health check failed"
        return 1
    fi
    
    # Check CMS API
    if curl -f http://localhost:8081/health &> /dev/null; then
        log_success "CMS API is healthy"
    else
        log_error "CMS API health check failed"
        return 1
    fi
    
    log_success "All services are healthy"
}

cleanup() {
    log_info "Cleaning up unused Docker resources..."
    
    docker system prune -af
    
    log_success "Cleanup completed"
}

rollback() {
    log_warning "Rolling back to previous version..."
    
    cd "$PROJECT_DIR"
    git reset --hard HEAD~1
    docker-compose up -d --force-recreate
    
    sleep 10
    
    if health_check; then
        log_success "Rollback successful"
    else
        log_error "Rollback failed"
        exit 1
    fi
}

# Main deployment flow
main() {
    log_info "Starting deployment to $ENVIRONMENT environment..."
    
    check_requirements
    
    if [ "$ENVIRONMENT" == "production" ]; then
        log_warning "Deploying to PRODUCTION environment"
        read -p "Are you sure you want to continue? (yes/no): " confirm
        if [ "$confirm" != "yes" ]; then
            log_info "Deployment cancelled"
            exit 0
        fi
        
        backup_database
    fi
    
    pull_latest_code
    build_images
    deploy_containers
    
    if health_check; then
        cleanup
        log_success "Deployment to $ENVIRONMENT completed successfully! 🚀"
    else
        log_error "Health check failed. Starting rollback..."
        rollback
        exit 1
    fi
}

# Handle script arguments
case "$1" in
    staging|production)
        main
        ;;
    rollback)
        rollback
        ;;
    *)
        echo "Usage: $0 {staging|production|rollback}"
        exit 1
        ;;
esac
