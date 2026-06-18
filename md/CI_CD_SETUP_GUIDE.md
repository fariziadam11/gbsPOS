# 🚀 CI/CD Setup Guide - GBS POS-CMS API

## ✅ Status Setup

### ✔️ Completed Steps

1. **✅ CI/CD Files Created**
   - GitHub Actions workflow (`.github/workflows/ci-cd.yml`)
   - GitLab CI configuration (`.gitlab-ci.yml`)
   - Linting configuration (`.golangci.yml`)
   - Deployment script (`scripts/deploy.sh`)
   - Production Docker Compose (`docker-compose.prod.yml`)
   - Nginx configuration (`nginx/nginx.conf`)
   - Environment template (`.env.production.example`)

2. **✅ Files Committed and Pushed**
   - All CI/CD files committed to repository
   - Pushed to `main` branch on GitHub
   - Commit: `feat: add comprehensive CI/CD pipeline`

3. **✅ Documentation Created**
   - `CI_CD_DOCUMENTATION.md` - Complete pipeline documentation
   - `ARCHITECTURE_DIAGRAM.md` - System architecture diagrams
   - `PROJECT_STATUS.md` - Project status and API documentation
   - `USE_CASE_DIAGRAM.md` - Use case diagrams

---

## 📋 Next Steps - GitHub Actions Setup

### Step 1: Add GitHub Secrets

Go to your GitHub repository: https://github.com/fariziadam11/gbs-pos-cms-api

Navigate to: **Settings → Secrets and variables → Actions → New repository secret**

Add the following secrets:

#### Staging Environment
```
Name: STAGING_SSH_KEY
Value: [Your staging server SSH private key]

Name: STAGING_HOST
Value: [Your staging server IP or hostname]
Example: staging.gbs-pos.com or 192.168.1.100

Name: STAGING_USER
Value: [SSH username for staging]
Example: ubuntu or deploy
```

#### Production Environment
```
Name: PRODUCTION_SSH_KEY
Value: [Your production server SSH private key]

Name: PRODUCTION_HOST
Value: [Your production server IP or hostname]
Example: api.gbs-pos.com or 192.168.1.200

Name: PRODUCTION_USER
Value: [SSH username for production]
Example: ubuntu or deploy
```

#### Notifications (Optional)
```
Name: SLACK_WEBHOOK
Value: [Your Slack webhook URL]
Example: https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

### Step 2: Generate SSH Keys (if needed)

If you don't have SSH keys yet:

```bash
# Generate SSH key pair
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/gbs-deploy

# Copy public key to servers
ssh-copy-id -i ~/.ssh/gbs-deploy.pub user@staging-server
ssh-copy-id -i ~/.ssh/gbs-deploy.pub user@production-server

# Copy private key content for GitHub Secret
cat ~/.ssh/gbs-deploy
# Copy the entire output including -----BEGIN and -----END lines
```

### Step 3: Prepare Deployment Servers

On both staging and production servers:

```bash
# 1. Install Docker and Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# 2. Create project directory
sudo mkdir -p /opt/gbs-pos-cms-api
sudo chown $USER:$USER /opt/gbs-pos-cms-api

# 3. Clone repository
cd /opt/gbs-pos-cms-api
git clone https://github.com/fariziadam11/gbs-pos-cms-api.git .

# 4. Create .env.production file
cp .env.production.example .env.production
nano .env.production
# Fill in the production values

# 5. Create backup directory
mkdir -p /backup
```

### Step 4: Test GitHub Actions

```bash
# Create a test branch
git checkout -b test/ci-pipeline

# Make a small change
echo "# CI/CD Test" >> README.md

# Commit and push
git add README.md
git commit -m "test: trigger CI/CD pipeline"
git push origin test/ci-pipeline

# Create Pull Request on GitHub
# Check Actions tab to see pipeline running
```

---

## 📋 GitLab CI Setup (Alternative)

If using GitLab instead of GitHub:

### Step 1: Add GitLab CI Variables

Go to: **Settings → CI/CD → Variables → Add variable**

Add the same variables as GitHub secrets above.

**Important**: Mark sensitive variables as:
- ✅ Protected
- ✅ Masked

### Step 2: Enable Container Registry

Go to: **Settings → General → Visibility → Container Registry**
- Enable Container Registry

### Step 3: Test Pipeline

```bash
# Push to trigger pipeline
git push origin main

# Check CI/CD → Pipelines to see status
```

---

## 🧪 Local Testing

### Test Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linting
cd gbs-pos-api
golangci-lint run --timeout=5m

cd ../gbs-cms-api
golangci-lint run --timeout=5m
```

### Test Unit Tests

```bash
# Start PostgreSQL for testing
docker run -d --name postgres-test \
  -e POSTGRES_DB=gbs_pos_test \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:15-alpine

# Set environment variables
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/gbs_pos_test?sslmode=disable"
export JWT_SECRET="test-secret-key-minimum-32-characters-long"

# Run tests
cd gbs-pos-api
go test -v -race -coverprofile=coverage.out ./...

cd ../gbs-cms-api
go test -v -race -coverprofile=coverage.out ./...

# View coverage
go tool cover -html=coverage.out
```

### Test Docker Build

```bash
# Build images
docker build -f gbs-pos-api/Dockerfile -t gbs-pos-api:test .
docker build -f gbs-cms-api/Dockerfile -t gbs-cms-api:test .

# Test run
docker-compose up -d

# Check health
curl http://localhost:8080/health
curl http://localhost:8081/health
```

### Test Deployment Script

```bash
# Make script executable (Linux/Mac)
chmod +x scripts/deploy.sh

# Test staging deployment (dry run)
./scripts/deploy.sh staging

# Test production deployment (dry run)
./scripts/deploy.sh production
```

---

## 🔍 Monitoring Pipeline

### GitHub Actions

1. Go to **Actions** tab in your repository
2. Click on latest workflow run
3. View logs for each job:
   - Lint & Test
   - Security Scan
   - Build Images
   - Deploy Staging/Production

### GitLab CI

1. Go to **CI/CD → Pipelines**
2. Click on latest pipeline
3. View logs for each stage

---

## 🐛 Troubleshooting

### Issue: Linting Errors

**Current Status**: Some linting issues detected in code

**Solution**:
```bash
# Auto-fix formatting issues
cd gbs-pos-api
gofmt -w .

# Auto-fix some linting issues
golangci-lint run --fix

# Commit fixes
git add .
git commit -m "fix: resolve linting issues"
git push
```

### Issue: SSH Connection Failed

**Solution**:
```bash
# Test SSH connection manually
ssh -i ~/.ssh/gbs-deploy user@server-ip

# Check SSH key permissions
chmod 600 ~/.ssh/gbs-deploy

# Verify key is added to server
cat ~/.ssh/gbs-deploy.pub
# Compare with ~/.ssh/authorized_keys on server
```

### Issue: Docker Build Failed

**Solution**:
```bash
# Check Dockerfile syntax
docker build -f gbs-pos-api/Dockerfile .

# Check for missing dependencies
cd gbs-pos-api
go mod tidy
go mod verify
```

### Issue: Tests Failed

**Solution**:
```bash
# Run tests locally with verbose output
go test -v ./...

# Check database connection
psql -h localhost -U postgres -d gbs_pos_test

# Check environment variables
echo $DATABASE_URL
echo $JWT_SECRET
```

---

## 📊 Pipeline Status

### Current Status: ⚠️ Partially Complete

- ✅ CI/CD files created and pushed
- ✅ Documentation complete
- ⚠️ GitHub Secrets need to be added
- ⚠️ Deployment servers need to be prepared
- ⚠️ Some linting issues need to be fixed
- ⏳ Pipeline not yet tested end-to-end

### To Complete Setup:

1. **Add GitHub Secrets** (5 minutes)
   - STAGING_SSH_KEY, STAGING_HOST, STAGING_USER
   - PRODUCTION_SSH_KEY, PRODUCTION_HOST, PRODUCTION_USER
   - SLACK_WEBHOOK (optional)

2. **Prepare Servers** (15 minutes per server)
   - Install Docker
   - Clone repository
   - Configure environment variables

3. **Fix Linting Issues** (10 minutes)
   - Run `gofmt -w .`
   - Run `golangci-lint run --fix`
   - Commit and push

4. **Test Pipeline** (5 minutes)
   - Create test branch
   - Push changes
   - Verify pipeline runs successfully

**Total Estimated Time**: ~1 hour

---

## 🎯 Quick Start Commands

```bash
# 1. Fix linting issues
cd gbs-pos-api && gofmt -w . && cd ..
cd gbs-cms-api && gofmt -w . && cd ..
git add . && git commit -m "fix: format code" && git push

# 2. Create develop branch for staging
git checkout -b develop
git push origin develop

# 3. Test local deployment
./scripts/deploy.sh staging

# 4. Monitor pipeline
# Go to GitHub Actions tab
```

---

## 📞 Support

For help with CI/CD setup:
- 📧 Email: devops@gbs-pos.com
- 💬 Slack: #devops-support
- 📖 Documentation: `CI_CD_DOCUMENTATION.md`

---

## 📝 Checklist

### Pre-Deployment Checklist

- [ ] GitHub Secrets added
- [ ] SSH keys generated and distributed
- [ ] Staging server prepared
- [ ] Production server prepared
- [ ] Linting issues fixed
- [ ] Tests passing locally
- [ ] Docker images building successfully
- [ ] Environment variables configured
- [ ] Backup strategy in place
- [ ] Monitoring configured

### Post-Deployment Checklist

- [ ] Pipeline runs successfully
- [ ] Staging deployment works
- [ ] Production deployment works (manual approval)
- [ ] Health checks passing
- [ ] Rollback tested
- [ ] Team notified
- [ ] Documentation updated

---

**Last Updated**: 2026-05-29
**Status**: Ready for GitHub Secrets configuration
**Next Action**: Add GitHub Secrets and prepare deployment servers
