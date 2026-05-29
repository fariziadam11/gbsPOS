# CI/CD Pipeline Documentation

## 📋 Overview

Dokumentasi lengkap untuk CI/CD pipeline GBS POS-CMS API yang mendukung GitHub Actions dan GitLab CI.

---

## 🏗️ Pipeline Architecture

```mermaid
graph LR
    A[Code Push] --> B[Lint & Test]
    B --> C[Security Scan]
    C --> D[Build Images]
    D --> E{Branch?}
    E -->|develop| F[Deploy Staging]
    E -->|main| G[Deploy Production]
    G --> H[Health Check]
    H -->|Failed| I[Rollback]
    H -->|Success| J[Notify Success]
```

---

## 🔄 Pipeline Stages

### 1. **Lint & Test Stage**

**Tujuan**: Memastikan kualitas kode dan menjalankan unit tests

**Jobs**:
- `lint-pos-api` - Linting untuk POS API menggunakan golangci-lint
- `lint-cms-api` - Linting untuk CMS API menggunakan golangci-lint
- `test-pos-api` - Unit tests untuk POS API dengan coverage report
- `test-cms-api` - Unit tests untuk CMS API dengan coverage report

**Tools**:
- golangci-lint v1.55+
- Go test dengan race detector
- Coverage reporting ke Codecov

**Triggers**:
- Push ke branch `main` atau `develop`
- Pull Request ke branch `main` atau `develop`

---

### 2. **Security Scan Stage**

**Tujuan**: Mendeteksi vulnerabilities dan security issues

**Jobs**:
- `security-trivy` - Scan filesystem untuk vulnerabilities
- `security-gosec` - Static security analysis untuk Go code

**Tools**:
- Trivy (Aqua Security)
- gosec (Golang Security Checker)

**Output**:
- SARIF reports untuk GitHub Security
- JSON reports untuk GitLab Security Dashboard

---

### 3. **Build Images Stage**

**Tujuan**: Build dan push Docker images ke registry

**Jobs**:
- `build-pos-api` - Build POS API Docker image
- `build-cms-api` - Build CMS API Docker image

**Features**:
- Multi-platform builds (linux/amd64, linux/arm64)
- Layer caching untuk build yang lebih cepat
- Automatic tagging (branch, SHA, latest)
- Image vulnerability scanning

**Registry**:
- GitHub: `ghcr.io`
- GitLab: `registry.gitlab.com`

---

### 4. **Deploy Staging Stage**

**Tujuan**: Deploy ke staging environment untuk testing

**Triggers**:
- Automatic pada push ke branch `develop`

**Steps**:
1. SSH ke staging server
2. Pull latest code
3. Pull Docker images
4. Restart containers dengan `docker-compose`
5. Health check
6. Cleanup unused resources

**Environment**:
- URL: `https://staging-api.gbs-pos.com`
- Auto-deploy: Yes
- Approval: Not required

---

### 5. **Deploy Production Stage**

**Tujuan**: Deploy ke production environment

**Triggers**:
- Manual approval pada push ke branch `main`

**Steps**:
1. **Backup database** (PostgreSQL dump)
2. SSH ke production server
3. Pull latest code
4. Pull Docker images
5. Restart containers dengan `docker-compose`
6. Health check (15 seconds wait)
7. Cleanup unused resources
8. Send notification (Slack/Email)

**Environment**:
- URL: `https://api.gbs-pos.com`
- Auto-deploy: No (Manual approval required)
- Approval: Required

**Rollback**:
- Automatic rollback jika health check gagal
- Manual rollback via workflow dispatch

---

## 🔐 Required Secrets

### GitHub Actions Secrets

```bash
# Staging Environment
STAGING_SSH_KEY          # SSH private key untuk staging server
STAGING_HOST             # Hostname staging server
STAGING_USER             # Username SSH staging

# Production Environment
PRODUCTION_SSH_KEY       # SSH private key untuk production server
PRODUCTION_HOST          # Hostname production server
PRODUCTION_USER          # Username SSH production

# Notifications
SLACK_WEBHOOK            # Slack webhook URL untuk notifikasi
```

### GitLab CI Variables

```bash
# Staging Environment
STAGING_SSH_KEY          # SSH private key (Protected, Masked)
STAGING_HOST             # Hostname staging server
STAGING_USER             # Username SSH staging

# Production Environment
PRODUCTION_SSH_KEY       # SSH private key (Protected, Masked)
PRODUCTION_HOST          # Hostname production server
PRODUCTION_USER          # Username SSH production

# Registry
CI_REGISTRY              # GitLab Container Registry URL (auto)
CI_REGISTRY_USER         # Registry username (auto)
CI_REGISTRY_PASSWORD     # Registry password (auto)
```

---

## 🚀 Deployment Flow

### Staging Deployment

```bash
# 1. Create feature branch
git checkout -b feature/new-feature

# 2. Make changes and commit
git add .
git commit -m "feat: add new feature"

# 3. Push to develop
git checkout develop
git merge feature/new-feature
git push origin develop

# 4. Pipeline automatically deploys to staging
# 5. Test on https://staging-api.gbs-pos.com
```

### Production Deployment

```bash
# 1. Merge develop to main
git checkout main
git merge develop

# 2. Tag release (optional)
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 3. Push to main
git push origin main

# 4. Go to GitHub Actions / GitLab CI
# 5. Approve production deployment
# 6. Monitor deployment progress
# 7. Verify on https://api.gbs-pos.com
```

---

## 🔄 Rollback Procedures

### Automatic Rollback

Pipeline akan otomatis rollback jika:
- Health check gagal setelah deployment
- Container gagal start
- Database migration error

### Manual Rollback

#### GitHub Actions

```bash
# 1. Go to Actions tab
# 2. Select "CI/CD Pipeline" workflow
# 3. Click "Run workflow"
# 4. Select "rollback" option
# 5. Confirm rollback
```

#### GitLab CI

```bash
# 1. Go to CI/CD > Pipelines
# 2. Find latest pipeline
# 3. Click "rollback:production" job
# 4. Click "Play" button
# 5. Confirm rollback
```

#### Manual via Script

```bash
# SSH to production server
ssh user@production-server

# Navigate to project directory
cd /opt/gbs-pos-cms-api

# Run rollback script
./scripts/deploy.sh rollback
```

---

## 🧪 Testing Pipeline Locally

### Test Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linting
cd gbs-pos-api
golangci-lint run

cd ../gbs-cms-api
golangci-lint run
```

### Test Unit Tests

```bash
# Set environment variables
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/gbs_pos_test?sslmode=disable"
export JWT_SECRET="test-secret-key-minimum-32-characters-long"

# Run tests
cd gbs-pos-api
go test -v -race -coverprofile=coverage.out ./...

cd ../gbs-cms-api
go test -v -race -coverprofile=coverage.out ./...
```

### Test Docker Build

```bash
# Build POS API
docker build -f gbs-pos-api/Dockerfile -t gbs-pos-api:test .

# Build CMS API
docker build -f gbs-cms-api/Dockerfile -t gbs-cms-api:test .

# Test run
docker-compose up -d
```

### Test Security Scan

```bash
# Install Trivy
# Windows: choco install trivy
# Mac: brew install trivy
# Linux: apt-get install trivy

# Run Trivy scan
trivy fs --severity HIGH,CRITICAL .

# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run gosec
gosec ./gbs-pos-api/...
gosec ./gbs-cms-api/...
```

---

## 📊 Monitoring & Notifications

### Health Checks

Pipeline melakukan health check pada endpoint:
- POS API: `http://localhost:8080/health`
- CMS API: `http://localhost:8081/health`

### Notifications

**Slack Integration**:
- ✅ Deployment success
- ❌ Deployment failure
- 🔄 Rollback triggered
- ⚠️ Security vulnerabilities found

**Email Notifications**:
- Configured via GitHub/GitLab settings
- Sent to repository maintainers

---

## 🛠️ Troubleshooting

### Pipeline Fails at Lint Stage

**Problem**: Linting errors

**Solution**:
```bash
# Run linting locally
golangci-lint run --fix

# Commit fixes
git add .
git commit -m "fix: linting issues"
git push
```

### Pipeline Fails at Test Stage

**Problem**: Unit tests failing

**Solution**:
```bash
# Run tests locally with verbose output
go test -v ./...

# Fix failing tests
# Commit and push
```

### Pipeline Fails at Build Stage

**Problem**: Docker build errors

**Solution**:
```bash
# Check Dockerfile syntax
docker build -f gbs-pos-api/Dockerfile .

# Check for missing dependencies
go mod tidy
```

### Deployment Fails

**Problem**: SSH connection or deployment errors

**Solution**:
```bash
# Verify SSH key is correct
ssh -i ~/.ssh/id_rsa user@server

# Check server disk space
df -h

# Check Docker status
docker ps
docker-compose ps
```

### Health Check Fails

**Problem**: Services not responding after deployment

**Solution**:
```bash
# Check container logs
docker logs gbs-pos-cms-api-pos-api-1
docker logs gbs-pos-cms-api-cms-api-1

# Check database connection
docker exec -it gbs-pos-cms-api-postgres-1 psql -U postgres -d gbs_pos

# Restart services
docker-compose restart
```

---

## 📈 Performance Optimization

### Build Time Optimization

1. **Use Layer Caching**
   - GitHub Actions: `cache-from: type=gha`
   - GitLab CI: Cache Go modules

2. **Parallel Jobs**
   - Lint and test jobs run in parallel
   - Build jobs run in parallel

3. **Conditional Execution**
   - Only build on main/develop branches
   - Skip unnecessary jobs on PRs

### Deployment Time Optimization

1. **Pre-built Images**
   - Use pre-built images from registry
   - Avoid rebuilding on deployment

2. **Rolling Updates**
   - Use `docker-compose up -d` for zero-downtime
   - Health checks before marking as complete

---

## 🔒 Security Best Practices

1. **Secrets Management**
   - Never commit secrets to repository
   - Use GitHub Secrets / GitLab Variables
   - Rotate secrets regularly

2. **Image Scanning**
   - Scan images for vulnerabilities
   - Block deployment if critical issues found

3. **Access Control**
   - Require approval for production deployments
   - Limit SSH access to deployment servers
   - Use separate keys for staging/production

4. **Audit Logging**
   - Log all deployments
   - Track who deployed what and when
   - Keep deployment history

---

## 📝 Maintenance

### Regular Tasks

**Weekly**:
- Review security scan reports
- Check test coverage trends
- Monitor deployment success rate

**Monthly**:
- Update dependencies
- Review and rotate secrets
- Clean up old Docker images
- Review and optimize pipeline

**Quarterly**:
- Update CI/CD tools (golangci-lint, Trivy, etc.)
- Review and update deployment procedures
- Conduct disaster recovery drills

---

## 🆘 Emergency Procedures

### Complete System Failure

1. **Immediate Actions**:
   ```bash
   # Rollback to last known good version
   ./scripts/deploy.sh rollback
   
   # Check all services
   docker-compose ps
   docker-compose logs
   ```

2. **Restore from Backup**:
   ```bash
   # Restore database
   docker exec -i gbs-pos-cms-api-postgres-1 psql -U postgres gbs_pos < backup.sql
   
   # Restart services
   docker-compose restart
   ```

3. **Notify Team**:
   - Alert on-call engineer
   - Update status page
   - Communicate with stakeholders

---

## 📚 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitLab CI Documentation](https://docs.gitlab.com/ee/ci/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Go Testing Guide](https://go.dev/doc/tutorial/add-a-test)
- [golangci-lint Linters](https://golangci-lint.run/usage/linters/)

---

## 📞 Support

Untuk bantuan dengan CI/CD pipeline:
- 📧 Email: devops@gbs-pos.com
- 💬 Slack: #devops-support
- 📖 Wiki: https://wiki.gbs-pos.com/cicd

---

**Last Updated**: 2026-05-29
**Version**: 1.0.0
