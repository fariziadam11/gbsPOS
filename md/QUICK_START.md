# 🚀 Quick Start Guide - CI/CD Setup

## ✅ Status: Script Fixed!

PowerShell script telah diperbaiki dan siap digunakan.

---

## 📋 Step-by-Step Instructions

### **Step 1: Run Helper Script** (5 minutes)

```powershell
cd C:\laragon\www\gbs_pos_cms\gbs-pos-cms-api
.\scripts\setup-github-secrets.ps1
```

**Menu Options:**
1. **Copy STAGING_SSH_KEY** - Copy private key untuk staging
2. **Copy PRODUCTION_SSH_KEY** - Copy private key untuk production
3. **View Public Key** - Lihat public key untuk server
4. **Generate Secure Password** - Generate password untuk database
5. **Generate JWT Secret** - Generate JWT secret
6. **Open GitHub Secrets Page** - Buka halaman GitHub Secrets
7. **Show All Secrets Template** - Lihat template semua secrets
0. **Exit** - Keluar

---

### **Step 2: Add GitHub Secrets** (5 minutes)

Script akan membuka halaman ini (atau buka manual):
https://github.com/fariziadam11/gbs-pos-cms-api/settings/secrets/actions

**Tambahkan 6 secrets ini:**

| Secret Name | How to Get | Example Value |
|-------------|------------|---------------|
| `STAGING_SSH_KEY` | Script option 1 | [Private key content] |
| `STAGING_HOST` | Your server | `192.168.1.100` or `staging.example.com` |
| `STAGING_USER` | Your server | `ubuntu` or `deploy` |
| `PRODUCTION_SSH_KEY` | Script option 2 | [Private key content] |
| `PRODUCTION_HOST` | Your server | `192.168.1.200` or `api.example.com` |
| `PRODUCTION_USER` | Your server | `ubuntu` or `deploy` |

**Optional:**
| Secret Name | How to Get |
|-------------|------------|
| `SLACK_WEBHOOK` | https://api.slack.com/messaging/webhooks |

---

### **Step 3: Test Without Servers** (5 minutes)

Jika Anda **belum punya server**, Anda tetap bisa test pipeline:

```powershell
# Trigger pipeline
git checkout develop
echo "# Test Pipeline" > TEST.md
git add TEST.md
git commit -m "test: trigger CI/CD pipeline"
git push origin develop
```

**Cek hasil:**
https://github.com/fariziadam11/gbs-pos-cms-api/actions

Pipeline akan:
- ✅ Run linting
- ✅ Run tests
- ✅ Run security scan
- ✅ Build Docker images
- ⏸️ Skip deployment (karena belum ada server)

---

### **Step 4: Setup Servers** (Optional - 30 minutes)

Jika Anda punya server staging/production:

#### 4.1 Copy Public Key ke Server

```powershell
# Copy public key (script option 3)
.\scripts\setup-github-secrets.ps1
# Pilih option 3

# SSH ke server
ssh user@your-server

# Add public key
mkdir -p ~/.ssh
nano ~/.ssh/authorized_keys
# Paste public key, save and exit

chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys
exit
```

#### 4.2 Install Docker di Server

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify
docker --version
docker-compose --version

# Logout and login again
exit
```

#### 4.3 Setup Project di Server

```bash
# Create directory
sudo mkdir -p /opt/gbs-pos-cms-api
sudo chown $USER:$USER /opt/gbs-pos-cms-api

# Clone repository
cd /opt/gbs-pos-cms-api
git clone https://github.com/fariziadam11/gbs-pos-cms-api.git .

# Create backup directory
mkdir -p /backup

# Setup environment
cp .env.production.example .env.production
nano .env.production
```

#### 4.4 Configure Environment

Generate secure values menggunakan script (option 4 & 5):

```env
DATABASE_URL=postgres://postgres:YOUR_PASSWORD@postgres:5432/gbs_pos?sslmode=disable
POSTGRES_PASSWORD=YOUR_PASSWORD_FROM_SCRIPT
JWT_SECRET=YOUR_JWT_SECRET_FROM_SCRIPT
```

#### 4.5 Test Docker

```bash
cd /opt/gbs-pos-cms-api

# Start services
docker-compose up -d

# Check status
docker-compose ps

# Test APIs
curl http://localhost:8080/health
curl http://localhost:8081/health

# Stop
docker-compose down
```

---

## 🎯 What Happens After Setup?

### **When you push to `develop` branch:**
```
Code Push → Lint → Test → Security Scan → Build → Deploy to Staging
```

### **When you push to `main` branch:**
```
Code Push → Lint → Test → Security Scan → Build → Wait for Approval → Deploy to Production
```

---

## 📊 Monitoring

### **GitHub Actions**
https://github.com/fariziadam11/gbs-pos-cms-api/actions

### **Check Pipeline Status**
- Green checkmark = Success
- Red X = Failed
- Yellow circle = Running
- Gray circle = Waiting

### **View Logs**
Click on workflow run → Click on job → View detailed logs

---

## 🆘 Troubleshooting

### **Script Error**
```powershell
# If script fails, check PowerShell version
$PSVersionTable.PSVersion

# Should be 5.1 or higher
```

### **SSH Connection Failed**
```powershell
# Test SSH connection
ssh -i "$env:USERPROFILE\.ssh\gbs-deploy" user@server

# If fails, check:
# 1. Public key is in ~/.ssh/authorized_keys on server
# 2. Permissions are correct (700 for .ssh, 600 for authorized_keys)
# 3. Server is accessible from your network
```

### **Pipeline Failed**
1. Check GitHub Actions logs
2. Verify all secrets are added correctly
3. Check secret names match exactly (case-sensitive)
4. Verify SSH key format is correct (includes BEGIN and END lines)

---

## 📚 Full Documentation

For detailed information, read:

1. **SETUP_COMPLETE.md** - Complete setup guide
2. **DEPLOYMENT_KEYS.md** - SSH keys and server setup
3. **CI_CD_SUMMARY.md** - Quick reference
4. **CI_CD_DOCUMENTATION.md** - Technical documentation

---

## ✅ Checklist

### Minimal Setup (No Servers)
- [ ] Run helper script
- [ ] Add 6 GitHub Secrets
- [ ] Push to develop branch
- [ ] Check GitHub Actions

**Time: ~15 minutes**

### Full Setup (With Servers)
- [ ] Run helper script
- [ ] Add 6 GitHub Secrets
- [ ] Copy public key to servers
- [ ] Install Docker on servers
- [ ] Setup project on servers
- [ ] Configure environment
- [ ] Test Docker
- [ ] Push to develop branch
- [ ] Verify deployment

**Time: ~45 minutes**

---

## 🎉 Success!

Once setup is complete:
- ✅ Every push to `develop` auto-deploys to staging
- ✅ Every push to `main` requires approval for production
- ✅ Automatic rollback on failure
- ✅ Health checks after deployment
- ✅ Slack/Email notifications

---

**Repository**: https://github.com/fariziadam11/gbs-pos-cms-api

**Next Command**:
```powershell
.\scripts\setup-github-secrets.ps1
```

Good luck! 🚀
