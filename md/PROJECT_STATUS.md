# GBS POS-CMS API - Project Status

## ✅ Project Successfully Running!

All services are up and running on your local machine.

---

## 🚀 Running Services

| Service | Status | Port | URL |
|---------|--------|------|-----|
| **PostgreSQL** | ✅ Running | 5433 | `localhost:5433` |
| **POS API** | ✅ Running | 8080 | http://localhost:8080 |
| **CMS API** | ✅ Running | 8081 | http://localhost:8081 |

---

## 🔐 Default Credentials

- **Username**: `admin`
- **Password**: `admin123`
- **Role**: `ADMIN`

---

## 📡 API Endpoints

### POS API (Port 8080)

**Base URL**: `http://localhost:8080/v1`

#### Authentication
- `POST /login` - Login and get JWT token

#### Products
- `GET /products` - List all products (filter by `storeType`)
- `POST /products` - Create product (ADMIN only)
- `PUT /products/:id` - Update product (ADMIN only)
- `DELETE /products/:id` - Delete product (ADMIN only)

#### Orders
- `GET /orders` - List orders (with filters)
- `GET /orders/:id` - Get order detail
- `POST /orders` - Create order (idempotent)
- `PATCH /orders/:id/void` - Void order (ADMIN only)
- `GET /orders/unsettled/summary` - Batch stats for settlement
- `POST /orders/settle` - Run settlement (ADMIN only)

#### Settlements
- `GET /settlements` - List recent settlements
- `GET /settlements/:id` - Get settlement detail

### CMS API (Port 8081)

**Base URL**: `http://localhost:8081/v1`

#### Authentication
- `POST /login` - Login and get JWT token

#### Ads Management
- `POST /ads/upload` - Upload video ad
- `GET /ads` - List all ads
- `GET /ads/:id` - Get ad details
- `DELETE /ads/:id` - Delete ad
- `GET /ads/:id/download` - Download video file

#### Playlists
- `POST /playlists` - Create playlist
- `GET /playlists` - List playlists
- `PUT /playlists/:id` - Update playlist
- `DELETE /playlists/:id` - Delete playlist

---

## 🧪 Test the APIs

### Login to POS API
```powershell
Invoke-WebRequest -Uri "http://localhost:8080/v1/login" `
  -Method POST `
  -Headers @{"Content-Type"="application/json"} `
  -Body '{"username":"admin","password":"admin123"}' `
  -UseBasicParsing | Select-Object -ExpandProperty Content
```

### Login to CMS API
```powershell
Invoke-WebRequest -Uri "http://localhost:8081/v1/login" `
  -Method POST `
  -Headers @{"Content-Type"="application/json"} `
  -Body '{"username":"admin","password":"admin123"}' `
  -UseBasicParsing | Select-Object -ExpandProperty Content
```

### Get Products (POS API)
```powershell
$token = "YOUR_JWT_TOKEN_HERE"
Invoke-WebRequest -Uri "http://localhost:8080/v1/products?storeType=RETAIL" `
  -Method GET `
  -Headers @{"Authorization"="Bearer $token"} `
  -UseBasicParsing | Select-Object -ExpandProperty Content
```

---

## 🛠️ Management Commands

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker logs gbs-pos-cms-api-pos-api-1 -f
docker logs gbs-pos-cms-api-cms-api-1 -f
docker logs gbs-pos-cms-api-postgres-1 -f
```

### Stop Services
```bash
cd gbs-pos-cms-api
docker-compose down
```

### Restart Services
```bash
cd gbs-pos-cms-api
docker-compose restart
```

### Rebuild and Restart
```bash
cd gbs-pos-cms-api
docker-compose down
docker-compose up --build -d
```

### Access PostgreSQL Database
```bash
docker exec -it gbs-pos-cms-api-postgres-1 psql -U postgres -d gbs_pos
```

---

## 📊 Database Information

- **Database Name**: `gbs_pos`
- **Username**: `postgres`
- **Password**: `123`
- **Host**: `localhost`
- **Port**: `5433`
- **Connection String**: `postgres://postgres:123@localhost:5433/gbs_pos?sslmode=disable`

---

## 🏗️ Architecture

This is a **monorepo** with three Go modules:

1. **gbs-pos-api/** - POS REST API
   - Products management
   - Orders and settlements
   - Multi-store support (RETAIL, F&B, OUTFIT)
   - Offline-first sync strategy

2. **gbs-cms-api/** - CMS REST API
   - Video ad upload and management
   - Playlist management
   - Video streaming with range support

3. **gbs-common/** - Shared utilities
   - JWT authentication middleware
   - CORS middleware
   - Logger middleware
   - Standard response format

---

## 🔧 Issue Fixed

**Problem**: POS API was failing with "Dirty database version 8" migration error.

**Solution**: Reset the migration state by running:
```sql
UPDATE schema_migrations SET dirty = false WHERE version = 8;
```

Then restarted the POS API container.

---

## 📝 Next Steps

1. **Test the APIs** using the examples above
2. **Explore the codebase** in the `gbs-pos-api/`, `gbs-cms-api/`, and `gbs-common/` directories
3. **Read the documentation**:
   - `AGENTS.md` - Development guidelines
   - `BACKEND_API_GUIDE.md` - Complete API specification
4. **Connect your Android app** by updating the base URL to `http://YOUR_IP:8080/v1`

---

## 🎯 Store Types

The system supports three store types:

- **RETAIL** - Snacks, Beverages, Household, Personal Care
- **FNB** - Food, Beverages, Desserts
- **OUTFIT** - Tops, Bottoms, Outerwear, Accessories

---

## 💳 Payment Methods

- **CASH** - Cash payments with change calculation
- **CARD** - Neurogine SoftPOS integration
- **QRIS** - QR code payments

---

## 📞 Support

For issues or questions, refer to:
- `AGENTS.md` for development guidelines
- `BACKEND_API_GUIDE.md` for API documentation
- Docker logs for troubleshooting

---

**Status**: ✅ All systems operational
**Last Updated**: May 21, 2026
