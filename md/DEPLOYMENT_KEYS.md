# 🔐 Deployment Keys & Setup Instructions

## ✅ SSH Keys Generated

SSH key pair telah dibuat di: `C:\Users\ThinkPad X13 G1\.ssh\`

- **Private Key**: `gbs-deploy` (untuk GitHub Secrets)
- **Public Key**: `gbs-deploy.pub` (untuk server)

---

## 📋 Step 1: Add GitHub Secrets (5 minutes)

### 1.1 Buka GitHub Repository Settings

Buka: https://github.com/fariziadam11/gbs-pos-cms-api/settings/secrets/actions

### 1.2 Add Secrets

Klik **"New repository secret"** untuk setiap secret berikut:

#### **STAGING_SSH_KEY**

```
Name: STAGING_SSH_KEY
Value: [Copy isi file C:\Users\ThinkPad X13 G1\.ssh\gbs-deploy]
```

**Cara copy private key:**
```powershell
Get-Content "$env:USERPROFILE\.ssh\gbs-deploy" | Set-Clipboard
```
Paste ke GitHub Secret value field.

---

#### **STAGING_HOST**

```
Name: STAGING_HOST
Value: [IP atau hostname staging server Anda]
```

**Contoh:**
- `staging.gbs-pos.com` (jika punya domain)
- `192.168.1.100` (jika pakai IP lokal)
- `staging-server.example.com`

---

#### **STAGING_USER**

```
Name: STAGING_USER
Value: [Username SSH di staging server]
```

**Contoh:**
- `ubuntu` (untuk Ubuntu server)
- `deploy` (jika buat user khusus)
- `root` (tidak disarankan)

---

#### **PRODUCTION_SSH_KEY**

```
Name: PRODUCTION_SSH_KEY
Value: [Copy isi file C:\Users\ThinkPad X13 G1\.ssh\gbs-deploy]
```

**Gunakan SSH key yang sama untuk production.**

---

#### **PRODUCTION_HOST**

```
Name: PRODUCTION_HOST
Value: [IP atau hostname production server Anda]
```

**Contoh:**
- `api.gbs-pos.com`
- `192.168.1.200`
- `prod-server.example.com`

---

#### **PRODUCTION_USER**

```
Name: PRODUCTION_USER
Value: [Username SSH di production server]
```

**Contoh:**
- `ubuntu`
- `deploy`

---

#### **SLACK_WEBHOOK** (Optional)

```
Name: SLACK_WEBHOOK
Value: [Slack webhook URL]
```

**Cara mendapatkan Slack webhook:**
1. Buka: https://api.slack.com/messaging/webhooks
2. Create New App
3. Enable Incoming Webhooks
4. Add New Webhook to Workspace
5. Copy webhook URL

---

## 📋 Step 2: Prepare Servers (30 minutes)

### 2.1 Copy Public Key ke Servers

**Public Key Anda:**
```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIAHBZBT9fDhkIWxyXM2S4SB8AjX5djuY6WlWYfftj81 github-actions-gbs-deploy
```

**Cara copy ke server:**

#### Option A: Menggunakan ssh-copy-id (Linux/Mac)
```bash
ssh-copy-id -i ~/.ssh/gbs-deploy.pub user@staging-server
ssh-copy-id -i ~/.ssh/gbs-deploy.pub user@production-server
```

#### Option B: Manual (Windows/All)
```bash
# 1. SSH ke server
ssh user@staging-server

# 2. Tambahkan public key
mkdir -p ~/.ssh
echo "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIAHBZBT9fDhkIWxyXM2S4SB8AjX5djuY6WlWYfftj81 github-actions-gbs-deploy" >> ~/.ssh/authorized_keys
chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys

# 3. Exit dan test
exit
ssh -i C:\Users\ThinkPad X13 G1\.ssh\gbs-deploy user@staging-server
```

Ulangi untuk production server.

---

### 2.2 Install Docker di Server

**Jalankan di staging dan production server:**

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker-compose --version

# Logout and login again for group changes
exit
```

---

### 2.3 Setup Project Directory

**Jalankan di staging dan production server:**

```bash
# Create project directory
sudo mkdir -p /opt/gbs-pos-cms-api
sudo chown $USER:$USER /opt/gbs-pos-cms-api

# Clone repository
cd /opt/gbs-pos-cms-api
git clone https://github.com/fariziadam11/gbs-pos-cms-api.git .

# Create backup directory
mkdir -p /backup

# Create environment file
cp .env.production.example .env.production
```

---

### 2.4 Configure Environment Variables

**Edit `.env.production` di server:**

```bash
nano /opt/gbs-pos-cms-api/.env.production
```

**Isi dengan nilai production:**

```env
# Database Configuration
DATABASE_URL=postgres://postgres:YOUR_SECURE_PASSWORD@postgres:5432/gbs_pos?sslmode=disable
POSTGRES_DB=gbs_pos
POSTGRES_USER=postgres
POSTGRES_PASSWORD=YOUR_SECURE_PASSWORD_HERE

# JWT Configuration (MUST be at least 32 characters)
JWT_SECRET=YOUR_SECURE_JWT_SECRET_KEY_MINIMUM_32_CHARACTERS_LONG_HERE

# API Configuration
PORT_POS=8080
PORT_CMS=8081
ENV=production
LOG_LEVEL=info

# Upload Configuration
UPLOAD_DIR=/uploads/ads

# CORS Configuration
ALLOWED_ORIGINS=https://pos.gbs.com,https://cms.gbs.com
```

**Generate secure passwords:**
```bash
# Generate secure password
openssl rand -base64 32

# Generate JWT secret
openssl rand -hex 32
```

---

### 2.5 Test Docker Setup

**Test di server:**

```bash
cd /opt/gbs-pos-cms-api

# Pull images
docker-compose pull

# Start services
docker-compose up -d

# Check status
docker-compose ps

# Check logs
docker-compose logs -f

# Test APIs
curl http://localhost:8080/health
curl http://localhost:8081/health

# Stop services
docker-compose down
```

---

## 📋 Step 3: Test Pipeline (5 minutes)

### 3.1 Create Develop Branch

**Di komputer lokal:**

```bash
cd C:\laragon\www\gbs_pos_cms\gbs-pos-cms-api

# Create and push develop branch
git checkout -b develop
git push origin develop
```

### 3.2 Trigger Pipeline

**Make a test change:**

```bash
# Add test file
echo "# Pipeline Test" > PIPELINE_TEST.md
git add PIPELINE_TEST.md
git commit -m "test: trigger CI/CD pipeline"
git push origin develop
```

### 3.3 Monitor Pipeline

1. **Buka GitHub Actions:**
   https://github.com/fariziadam11/gbs-pos-cms-api/actions

2. **Lihat workflow run:**
   - Klik pada workflow run terbaru
   - Monitor setiap job:
     - ✅ Lint & Test
     - ✅ Security Scan
     - ✅ Build Images
     - ✅ Deploy Staging (jika secrets sudah di-add)

3. **Check logs:**
   - Klik pada setiap job untuk melihat detail logs
   - Pastikan semua job berhasil (green checkmark)

### 3.4 Verify Deployment

**Jika deploy ke staging berhasil:**

```bash
# Check staging API
curl https://staging-api.gbs-pos.com/health
# atau
curl http://STAGING_IP:8080/health

# Test login
curl -X POST https://staging-api.gbs-pos.com/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

---

## 🎯 Quick Reference

### **Copy Private Key to Clipboard**
```powershell
Get-Content "$env:USERPROFILE\.ssh\gbs-deploy" | Set-Clipboard
```

### **Copy Public Key to Clipboard**
```powershell
Get-Content "$env:USERPROFILE\.ssh\gbs-deploy.pub" | Set-Clipboard
```

### **Test SSH Connection**
```powershell
ssh -i "$env:USERPROFILE\.ssh\gbs-deploy" user@server-ip
```

### **View Private Key**
```powershell
Get-Content "$env:USERPROFILE\.ssh\gbs-deploy"
```

### **View Public Key**
```powershell
Get-Content "$env:USERPROFILE\.ssh\gbs-deploy.pub"
```

---

## ✅ Checklist

### GitHub Secrets Setup
- [ ] STAGING_SSH_KEY added
- [ ] STAGING_HOST added
- [ ] STAGING_USER added
- [ ] PRODUCTION_SSH_KEY added
- [ ] PRODUCTION_HOST added
- [ ] PRODUCTION_USER added
- [ ] SLACK_WEBHOOK added (optional)

### Staging Server Setup
- [ ] SSH key copied to server
- [ ] Docker installed
- [ ] Docker Compose installed
- [ ] Project directory created
- [ ] Repository cloned
- [ ] Environment variables configured
- [ ] Docker services tested

### Production Server Setup
- [ ] SSH key copied to server
- [ ] Docker installed
- [ ] Docker Compose installed
- [ ] Project directory created
- [ ] Repository cloned
- [ ] Environment variables configured
- [ ] Docker services tested

### Pipeline Testing
- [ ] Develop branch created
- [ ] Test commit pushed
- [ ] Pipeline triggered
- [ ] All jobs passed
- [ ] Staging deployment successful
- [ ] APIs responding

---

## 🆘 Troubleshooting

### SSH Connection Failed

**Problem:** Cannot connect to server

**Solution:**
```bash
# Check SSH key permissions
icacls "$env:USERPROFILE\.ssh\gbs-deploy"

# Test connection with verbose
ssh -v -i "$env:USERPROFILE\.ssh\gbs-deploy" user@server

# Check if public key is in authorized_keys on server
ssh user@server "cat ~/.ssh/authorized_keys"
```

### Docker Installation Failed

**Problem:** Docker not installing

**Solution:**
```bash
# Remove old Docker
sudo apt remove docker docker-engine docker.io containerd runc

# Install prerequisites
sudo apt install apt-transport-https ca-certificates curl software-properties-common

# Try installation again
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
```

### Pipeline Failed

**Problem:** GitHub Actions pipeline failing

**Solution:**
1. Check GitHub Actions logs
2. Verify all secrets are added correctly
3. Test SSH connection manually
4. Check server is accessible from internet
5. Verify Docker is running on server

---

## 📞 Need Help?

- 📖 Read: `CI_CD_DOCUMENTATION.md`
- 📖 Read: `CI_CD_SETUP_GUIDE.md`
- 💬 Contact: devops@gbs-pos.com

---

**Generated**: 2026-05-29
**SSH Key Location**: `C:\Users\ThinkPad X13 G1\.ssh\gbs-deploy`
**Repository**: https://github.com/fariziadam11/gbs-pos-cms-api
