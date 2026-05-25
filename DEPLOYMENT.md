# GBS POS & CMS — Production Deployment Guide

This guide covers deploying the GBS backend (POS API + CMS API + PostgreSQL) to a **single VPS** using **Docker Compose** and **Cloudflare Tunnel** (for zero-config HTTPS and no open ports).

**Last updated:** 2026-05-25

---

## Architecture

```
Cloudflare (TLS + DNS + DDoS)
    |
    v
Cloudflare Tunnel (outbound-only, no open ports)
    |
    v
VPS (Ubuntu 22.04 LTS)
    |-- Docker Compose
    |   |-- postgres:15-alpine  (:5432 internal)
    |   |-- gbs-pos-api         (:8080 container, mapped to host :8082, 127.0.0.1 only)
    |   `-- gbs-cms-api         (:8081, 127.0.0.1 only)
    `-- cloudflared
        |-- api-pos.yourdomain.com  -> localhost:8082
        `-- api-cms.yourdomain.com  -> localhost:8081
```

**Why Cloudflare Tunnel?**
- **Zero exposed ports** — outbound-only connection, no firewall rules needed.
- **Automatic HTTPS** — Cloudflare handles TLS certificates (no certbot, no reverse proxy).
- **DDoS protection** — built into Cloudflare's edge.
- **Cost-effective** — runs on a $5–12/mo VPS.

---

## Prerequisites

### 1. Domain & DNS
- A domain managed by **Cloudflare** (free tier is fine).
- Two A/CNAME records pointing to Cloudflare:
  - `api-pos.yourdomain.com`
  - `api-cms.yourdomain.com`

### 2. VPS
- **OS:** Ubuntu 22.04 LTS (or 24.04)
- **Specs:** 1–2 vCPU, 2–4 GB RAM, 20–40 GB SSD
- **Access:** SSH key auth (disable password login)

### 3. GitHub Account
- For GitHub Container Registry (`ghcr.io`) and GitHub Actions CI/CD.

---

## Step 1: Prepare the VPS

SSH into your VPS and run:

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker & Docker Compose
sudo apt install -y docker.io docker-compose-plugin
sudo usermod -aG docker $USER
newgrp docker

# Verify
docker --version
docker compose version

# Create deployment directory
sudo mkdir -p /opt/gbs
sudo chown $USER:$USER /opt/gbs
```

---

## Step 2: Install Cloudflare Tunnel

Follow the [official Cloudflare guide](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) to:

1. Install `cloudflared`:
```bash
wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared-linux-amd64.deb
```

2. Authenticate:
```bash
cloudflared tunnel login
# This opens a browser to authenticate with your Cloudflare account
```

3. Create the tunnel:
```bash
cloudflared tunnel create gbs-backend
# Note the Tunnel ID output
```

4. Create the config file:
```bash
mkdir -p ~/.cloudflared
```

Create `~/.cloudflared/config.yml`:
```yaml
tunnel: <YOUR_TUNNEL_ID>
credentials-file: /home/ubuntu/.cloudflared/<YOUR_TUNNEL_ID>.json

ingress:
  - hostname: api-pos.yourdomain.com
    service: http://localhost:8082
  - hostname: api-cms.yourdomain.com
    service: http://localhost:8081
  - service: http_status:404
```

5. Route DNS to the tunnel (in Cloudflare Dashboard):
- Go to **Zero Trust** -> **Networks** -> **Tunnels**
- Find `gbs-backend` -> **Configure**
- Add public hostnames:
  - `api-pos.yourdomain.com` -> `http://localhost:8082`
  - `api-cms.yourdomain.com` -> `http://localhost:8081`

6. Run the tunnel as a service:
```bash
sudo cloudflared service install
sudo systemctl start cloudflared
sudo systemctl enable cloudflared
```

---

## Step 3: Configure Production Secrets

On the VPS, create `/opt/gbs/.env`:

```bash
cd /opt/gbs
cat > .env << 'EOF'
# PostgreSQL
POSTGRES_USER=gbs_prod
POSTGRES_PASSWORD=<generate_a_32+_char_random_password>

# JWT (must be >= 32 characters; 64 hex chars recommended)
JWT_SECRET=<generate_a_64_char_random_hex>
EOF
```

**Secure the file:**
```bash
chmod 600 /opt/gbs/.env
```

**Never commit this file to Git.** It is already ignored via `.gitignore`.

---

## Step 4: Deploy the Stack

### Option A: First-Time Deploy (Manual — Build Locally)

On the VPS:

```bash
cd /opt/gbs

# Clone your repo (or use SCP to copy docker-compose.prod.yml)
git clone https://github.com/YOUR_USERNAME/YOUR_REPO.git .

# IMPORTANT: create .env at /opt/gbs/.env (see Step 3 above)
# Run from repo root so --env-file resolves correctly
docker compose -f docker-compose.prod.yml --env-file .env up -d --build

# Verify
docker compose -f docker-compose.prod.yml ps
docker logs -f gbs-pos-api
docker logs -f gbs-cms-api
```

**Note:** First deploy builds images locally because GHCR images don't exist yet. After CI/CD runs (Step 5), future deploys will pull pre-built images.

### Option B: Automated Deploy (GitHub Actions)

After setting up CI/CD (Step 5), update `docker-compose.prod.yml` to use pre-built images:

```yaml
  pos-api:
    image: ghcr.io/YOUR_USERNAME/gbs-pos-api:latest
    # build: ./gbs-pos-api   # comment out or remove

  cms-api:
    image: ghcr.io/YOUR_USERNAME/gbs-cms-api:latest
    # build: ./gbs-cms-api   # comment out or remove
```

Then every push to `main` will auto-deploy.

---

## Step 5: Set Up GitHub Actions CI/CD

### 5.1 Enable GitHub Container Registry

In your repo settings:
- Go to **Settings** -> **Packages and pages** -> **Package settings**
- Ensure "Inherit access from source repository" is enabled

### 5.2 Add Repository Secrets

Go to **Settings** -> **Secrets and variables** -> **Actions** -> **New repository secret**:

| Secret Name | Value |
|-------------|-------|
| `VPS_HOST` | Your VPS IP address (e.g., `123.45.67.89`) |
| `VPS_USER` | SSH username (e.g., `ubuntu`) |
| `VPS_SSH_KEY` | Contents of your **private** SSH key (`~/.ssh/id_rsa`) |

### 5.3 CI/CD Workflow

The workflow file `.github/workflows/deploy.yml` is already in this repo. It will:

1. **Test** — run `go test ./...` in both API modules
2. **Build** — build Docker images for both APIs
3. **Push** — push to `ghcr.io/YOUR_USERNAME/gbs-pos-api:latest` and `ghcr.io/YOUR_USERNAME/gbs-cms-api:latest`
4. **Deploy** — SSH into the VPS, pull new images, and restart

### 5.4 Trigger a Deploy

```bash
git push origin main
```

Then monitor the run at: `https://github.com/YOUR_USERNAME/YOUR_REPO/actions`

---

## Step 6: Verify the Deployment

### Health Checks

```bash
# POS API
curl https://api-pos.yourdomain.com/health
# Expected: "ok"

# CMS API
curl https://api-cms.yourdomain.com/health
# Expected: "ok"
```

### Login Test

```bash
# POS API login
curl -X POST https://api-pos.yourdomain.com/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
# Expected: JSON with token

# CMS API login
curl -X POST https://api-cms.yourdomain.com/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
# Expected: JSON with token
```

### Database

```bash
docker exec -it gbs-postgres psql -U gbs_prod -d gbs_pos -c "\dt"
```

---

## Step 7: Backup Strategy

Create `/opt/gbs/backup.sh`:

```bash
#!/bin/bash
set -e

BACKUP_DIR="/backups/gbs"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p "$BACKUP_DIR"

# Database dump
docker exec gbs-postgres pg_dump -U gbs_prod gbs_pos \
  > "$BACKUP_DIR/gbs_pos_$DATE.sql"

# Uploads archive
tar czf "$BACKUP_DIR/cms-uploads_$DATE.tar.gz" \
  -C /var/lib/docker/volumes/gbs_cms-uploads/_data .

# Keep only last 7 days
find "$BACKUP_DIR" -type f -mtime +7 -delete

echo "Backup completed: $DATE"
```

Make it executable and schedule daily:

```bash
chmod +x /opt/gbs/backup.sh
sudo crontab -e
# Add:
0 3 * * * /opt/gbs/backup.sh >> /var/log/gbs-backup.log 2>&1
```

**Optional:** Sync backups to Cloudflare R2 / AWS S3:
```bash
# Add to backup.sh
rclone sync /backups/gbs remote:my-backup-bucket/gbs
```

---

## Step 8: Monitoring (Optional but Recommended)

### Netdata (Free, Zero-Config)

```bash
wget -O /tmp/netdata-kickstart.sh https://my-netdata.io/kickstart.sh
sh /tmp/netdata-kickstart.sh --stable-channel
```

Add to `~/.cloudflared/config.yml`:
```yaml
  - hostname: monitor.yourdomain.com
    service: http://localhost:19999
```

Access at: `https://monitor.yourdomain.com`

---

## Maintenance

### Update the Stack

```bash
cd /opt/gbs
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
```

### View Logs

```bash
# All services
docker compose -f docker-compose.prod.yml logs -f

# Single service
docker logs -f gbs-pos-api
docker logs -f gbs-cms-api
docker logs -f gbs-postgres
```

### Restart a Service

```bash
docker compose -f docker-compose.prod.yml restart pos-api
docker compose -f docker-compose.prod.yml restart cms-api
```

### Database Migrations (Manual)

If you add new migration files:

```bash
# Migrations run automatically on API startup
# To run manually:
docker exec gbs-pos-api /app/gbs-pos-api migrate up
```

---

## Security Checklist

- [ ] SSH key auth only (`PasswordAuthentication no` in `/etc/ssh/sshd_config`)
- [ ] `.env` file has `chmod 600`
- [ ] `POSTGRES_PASSWORD` is strong (32+ random chars)
- [ ] `JWT_SECRET` is strong (64 hex chars)
- [ ] APIs bind to `127.0.0.1` only (not `0.0.0.0`)
- [ ] Cloudflare Tunnel is the only ingress (no exposed ports)
- [ ] Automatic security updates enabled (`sudo apt install unattended-upgrades`)
- [ ] Backups are running and tested

---

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| `JWT_SECRET is required` | `.env` not loaded | Ensure `.env` exists in `/opt/gbs` and `docker-compose.prod.yml` references it |
| `connection refused` to API | Service not ready | Wait for `depends_on` healthcheck; check `docker logs` |
| `401 Unauthorized` | Token missing or expired | Ensure `Authorization: Bearer <token>` header is sent |
| Video preview not loading | CMS download endpoint 401 | Fixed in `VideoPlayer.vue` — fetches blob via Axios with auth |
| `migration failed` | Schema conflict | Run `docker exec gbs-postgres psql -U gbs_prod -d gbs_pos` and inspect |
| Cloudflare 502 error | Tunnel not running | `sudo systemctl status cloudflared` |

---

## Files Created for Deployment

| File | Purpose |
|------|---------|
| `docker-compose.prod.yml` | Production Docker Compose (secrets via `.env`, health checks, restart policies) |
| `.github/workflows/deploy.yml` | GitHub Actions CI/CD (test -> build -> push -> deploy) |
| `gbs-pos-api/.dockerignore` | Prevents secrets/build artifacts from leaking into images |
| `gbs-cms-api/.dockerignore` | Same |
| `gbs-pos-api/cmd/server/main.go` | Added `/health` endpoint for Docker health checks |
| `gbs-cms-api/cmd/server/main.go` | Added `/health` endpoint for Docker health checks |
| `DEPLOYMENT.md` | This guide |

---

## Next Steps After Deployment

1. **Update `cms-web/.env.production`** to point to your Cloudflare domains:
   ```
   VITE_API_BASE_URL=https://api-cms.yourdomain.com
   ```

2. **Build and deploy the frontend** (optional — can be served from the same VPS via Caddy/nginx or deployed to Vercel/Netlify):
   ```bash
   cd cms-web
   bun run build
   # Copy dist/ to your static web server
   ```

3. **Set up log aggregation** (optional): Use `docker logs` with `journald` or ship to CloudWatch/Loki.

4. **Enable rate limiting** on Cloudflare (free tier includes basic rate limiting).
