# GBS CMS Website Deployment Guide

This guide covers deploying the `cms-web` frontend to production on **`cms.armmada.id`**, assuming the backend is already deployed per [`DEPLOYMENT.md`](./DEPLOYMENT.md) (POS API + CMS API + PostgreSQL on a single VPS behind Cloudflare Tunnel).

The frontend is a standard **Vite + Vue 3 SPA**. It is built into a static `dist/` folder and can be served by any static web server.

**Last updated:** 2026-06-15

---

## Architecture

```
Cloudflare (TLS + DNS)
    |
    v
Cloudflare Tunnel (outbound-only)
    |
    v
VPS (Ubuntu 22.04)
    |-- Docker Compose (backend services)
    |   |-- postgres:15-alpine
    |   |-- gbs-pos-api  (:8082 on host)
    |   |-- gbs-cms-api  (:8081 on host)
    |   `-- gbs-cms-web  (:3000 on host)   <-- Option B
    |
    `-- Caddy / nginx (Option A) on :80
```

Public hostnames served through the tunnel:

| Hostname | Destination | Purpose |
|----------|-------------|---------|
| `api-pos.armmada.id` | `http://localhost:8082` | POS API |
| `api-cms.armmada.id` | `http://localhost:8081` | CMS API |
| `cms.armmada.id` | `http://localhost:80` (Option A) or `http://localhost:3000` (Option B) | CMS website |

---

## Prerequisites

- The backend is already running and healthy (`api-pos.armmada.id/health` and `api-cms.armmada.id/health` return `ok`).
- [`bun`](https://bun.sh/) is installed on your development machine.
- A Cloudflare DNS record for `cms.armmada.id` exists (or will be created via the tunnel).

---

## Step 1: Configure Production Environment Variables

Vite embeds `VITE_*` variables at **build time**, so you must set them before building.

Create `cms-web/.env.production`:

```ini
VITE_API_BASE_URL=https://api-cms.armmada.id
VITE_POS_API_BASE_URL=https://api-pos.armmada.id
```

**Do not commit this file.** It is added to `.gitignore` automatically.

---

## Step 2: Update Backend CORS

The backend must allow the new frontend origin. This has already been updated in `gbs-common/middleware/cors.go`:

```go
AllowOrigins: []string{
    "https://cms.armmada.id",
    "https://cms.gbs.com",
    "http://localhost:5173",
    "http://localhost:3000",
},
```

Because `gbs-common` is shared by both APIs via a `replace` directive, you must **rebuild and redeploy both API images** after this change:

```bash
cd /path/to/pos-cms

# Build both API images
make build

# Or with Docker directly:
docker build -t ghcr.io/YOUR_GITHUB_USERNAME/gbs-pos-api:latest -f gbs-pos-api/Dockerfile .
docker build -t ghcr.io/YOUR_GITHUB_USERNAME/gbs-cms-api:latest -f gbs-cms-api/Dockerfile .

# Push and redeploy on the VPS
docker push ghcr.io/YOUR_GITHUB_USERNAME/gbs-pos-api:latest
docker push ghcr.io/YOUR_GITHUB_USERNAME/gbs-cms-api:latest
```

On the VPS:

```bash
cd /opt/gbs
docker compose -f docker-compose.prod.yml --env-file .env pull
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

---

## Option A — Manual Static Build + Caddy (Recommended for First Deploy)

This is the fastest way to get the site live. You build locally, copy the static files to the VPS, and serve them with Caddy.

### A.1 Build the frontend locally

```bash
cd cms-web
bun install
bun run build
```

This produces a `dist/` folder containing the production SPA.

### A.2 Copy `dist/` to the VPS

```bash
# Create the target directory on the VPS
ssh ubuntu@YOUR_VPS_IP "sudo mkdir -p /opt/gbs/cms-web/dist && sudo chown -R \$USER:\$USER /opt/gbs/cms-web"

# Copy the build output
scp -r cms-web/dist/* ubuntu@YOUR_VPS_IP:/opt/gbs/cms-web/dist/
```

### A.3 Install Caddy on the VPS

```bash
ssh ubuntu@YOUR_VPS_IP

sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https

curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
  | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg

curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
  | sudo tee /etc/apt/sources.list.d/caddy-stable.list

sudo apt update
sudo apt install -y caddy
```

### A.4 Configure Caddy

Create `/opt/gbs/Caddyfile`:

```caddyfile
{
    # Cloudflare handles TLS; serve plain HTTP on localhost only
    auto_https off
}

cms.armmada.id {
    root * /opt/gbs/cms-web/dist
    encode gzip
    try_files {path} /index.html
    file_server
}
```

Run Caddy with this config:

```bash
sudo caddy run --config /opt/gbs/Caddyfile
```

Or install it as a service:

```bash
sudo cp /opt/gbs/Caddyfile /etc/caddy/Caddyfile
sudo systemctl reload caddy
sudo systemctl enable caddy
```

### A.5 Add the public hostname to Cloudflare Tunnel

Edit `~/.cloudflared/config.yml` on the VPS:

```yaml
tunnel: <YOUR_TUNNEL_ID>
credentials-file: /home/ubuntu/.cloudflared/<YOUR_TUNNEL_ID>.json

ingress:
  - hostname: api-pos.armmada.id
    service: http://localhost:8082
  - hostname: api-cms.armmada.id
    service: http://localhost:8081
  - hostname: cms.armmada.id
    service: http://localhost:80
  - service: http_status:404
```

Restart the tunnel:

```bash
sudo systemctl restart cloudflared
```

### A.6 Updating the site

Whenever you release a new version:

```bash
cd cms-web
bun install
bun run build
scp -r dist/* ubuntu@YOUR_VPS_IP:/opt/gbs/cms-web/dist/
```

No server restart is required; Caddy serves the files directly from disk.

---

## Option B — Dockerized Frontend (Better for CI/CD)

This option builds the frontend into a Docker image and runs it as part of the existing Docker Compose stack. The image has already been wired into `docker-compose.prod.yml`.

### B.1 Build and push the image from your local machine

```bash
cd cms-web

# Build the image
docker build -t ghcr.io/YOUR_GITHUB_USERNAME/gbs-cms-web:latest .

# Push to GHCR
docker login ghcr.io -u YOUR_GITHUB_USERNAME
docker push ghcr.io/YOUR_GITHUB_USERNAME/gbs-cms-web:latest
```

Make the `gbs-cms-web` package public on GitHub (or log in to GHCR on the VPS).

### B.2 Start the frontend container on the VPS

The `docker-compose.prod.yml` already contains the `cms-web` service:

```yaml
  cms-web:
    image: ghcr.io/${GHCR_OWNER:-yourusername}/gbs-cms-web:latest
    container_name: gbs-cms-web
    restart: unless-stopped
    ports:
      - "127.0.0.1:3000:80"
    networks:
      - gbs
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:80/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

On the VPS:

```bash
cd /opt/gbs
docker compose -f docker-compose.prod.yml --env-file .env pull
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

### B.3 Add the public hostname to Cloudflare Tunnel

Edit `~/.cloudflared/config.yml`:

```yaml
ingress:
  - hostname: api-pos.armmada.id
    service: http://localhost:8082
  - hostname: api-cms.armmada.id
    service: http://localhost:8081
  - hostname: cms.armmada.id
    service: http://localhost:3000
  - service: http_status:404
```

Restart the tunnel:

```bash
sudo systemctl restart cloudflared
```

### B.4 Updating the site

Build and push the new image, then pull and restart on the VPS:

```bash
cd cms-web
docker build -t ghcr.io/YOUR_GITHUB_USERNAME/gbs-cms-web:latest .
docker push ghcr.io/YOUR_GITHUB_USERNAME/gbs-cms-web:latest

ssh ubuntu@YOUR_VPS_IP "cd /opt/gbs && docker compose -f docker-compose.prod.yml --env-file .env pull cms-web && docker compose -f docker-compose.prod.yml --env-file .env up -d cms-web"
```

---

## Step 3: Verify the Deployment

### DNS / Tunnel

```bash
# Should resolve through Cloudflare
curl -I https://cms.armmada.id
```

Expected: `HTTP/2 200` and HTML content.

### CORS / API

Open `https://cms.armmada.id` in a browser, log in, and check the DevTools **Console** and **Network** tabs. There should be no CORS errors when the frontend calls `api-cms.armmada.id` or `api-pos.armmada.id`.

### Login smoke test

```bash
# POS API
curl -X POST https://api-pos.armmada.id/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# CMS API
curl -X POST https://api-cms.armmada.id/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

---

## Files Created / Modified for Website Deployment

| File | Purpose |
|------|---------|
| `cms-web/Dockerfile` | Multi-stage Bun build → nginx image for Option B |
| `cms-web/nginx.conf` | SPA fallback + static asset caching for Option B |
| `docker-compose.prod.yml` | Added `cms-web` service (Option B) |
| `gbs-common/middleware/cors.go` | Added `https://cms.armmada.id` to allowed origins |
| `WEBSITE_DEPLOYMENT.md` | This guide |

---

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| Blank page after login | Static files missing / wrong path | Ensure `dist/` was copied or the image was rebuilt |
| `404` on refresh | SPA fallback not configured | Check Caddy `try_files` or nginx `try_files` |
| CORS errors | Origin not allowed | Confirm `https://cms.armmada.id` is in `gbs-common/middleware/cors.go` and APIs are redeployed |
| Cloudflare 502 | Tunnel pointing to wrong port | Option A → `localhost:80`; Option B → `localhost:3000` |
| Video preview fails | CMS download endpoint returns 401 | Already handled in `VideoPlayer.vue` via Axios with auth header |
| New build not reflected | Browser cache | Hard-refresh (`Ctrl+Shift+R`) or use incognito |

---

## Next Steps (Optional)

1. **Automate with GitHub Actions** — extend `.github/workflows/deploy.yml` to build and push `gbs-cms-web` and deploy it to the VPS.
2. **Add cache headers** for `/assets/*` (already included in the nginx Docker image).
3. **Set up log aggregation** for Caddy/nginx or the `cms-web` container.
