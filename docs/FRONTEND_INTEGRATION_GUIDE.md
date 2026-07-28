# Panduan Integrasi Frontend Website ke GBS POS & CMS API

Panduan ini membantu Anda mengintegrasikan website frontend dengan GBS POS API dan GBS CMS API.

---

## Daftar Isi

1. [Arsitektur & Endpoints](#1-arsitektur--endpoints)
2. [Autentikasi](#2-autentikasi)
3. [Struktur Response API](#3-struktur-response-api)
4. [Konfigurasi CORS](#4-konfigurasi-cors)
5. [Contoh Kode Frontend](#5-contoh-kode-frontend)
6. [Referensi API Lengkap](#6-referensi-api-lengkap)

---

## 1. Arsitektur & Endpoints

Sistem GBS POS terdiri dari dua API:

| API | Port Default | Base URL (Development) | Base URL (Production) |
|-----|--------------|------------------------|-----------------------|
| **POS API** | 8080 | `http://localhost:8080` | `https://api-pos.armmada.id` |
| **CMS API** | 8081 | `http://localhost:8081` | `https://api-cms.armmada.id` |

### Prefiks Versioning
Semua endpoints menggunakan prefix `/v1`:
```
http://localhost:8080/v1/...
http://localhost:8081/v1/...
```

### Health Check
```http
GET /health
```
Response: `ok`

---

## 2. Autentikasi

### 2.1 Mode Autentikasi

Sistem mendukung dual authentication:

| Mode | Algoritma | Keterangan |
|------|-----------|------------|
| **JWT (Demo)** | HS256 | Untuk development, aktifkan `ENABLE_DEMO_AUTH=true` |
| **Keycloak** | RS256 | Untuk production dengan SSO |

### 2.2 Login (JWT Mode)

**Endpoint:** `POST /v1/login`

```http
POST /v1/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

**Response Sukses:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "username": "admin",
      "name": "Admin User",
      "role": "ADMIN"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Response Error:**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid username or password"
  }
}
```

### 2.3 Menggunakan Token JWT

Tambahkan header `Authorization` pada setiap request yang memerlukan auth:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 2.4 Role-Based Access

| Role | Akses |
|------|-------|
| `ADMIN` | Full access (create, update, delete, void orders) |
| `CASHIER` | Read-only + create orders |

---

## 3. Struktur Response API

### 3.1 Success Response
```typescript
interface ApiResponse<T> {
  success: true;
  data: T;
  idempotent?: boolean;  // untuk operasi idempotent
}
```

### 3.2 Error Response
```typescript
interface ErrorResponse {
  success: false;
  error: {
    code: string;        // contoh: "INVALID_CREDENTIALS", "VALIDATION_ERROR"
    message: string;
    details?: {          // hanya untuk VALIDATION_ERROR
      field: string;
      message: string;
    }[];
  };
}
```

### 3.3 Error Codes Umum

| Code | HTTP Status | Keterangan |
|------|-------------|------------|
| `INVALID_CREDENTIALS` | 401 | Username/password salah |
| `UNAUTHORIZED` | 401 | Token tidak valid/expired |
| `FORBIDDEN` | 403 | Tidak punya akses |
| `VALIDATION_ERROR` | 422 | Input tidak valid |
| `NOT_FOUND` | 404 | Resource tidak ditemukan |
| `INTERNAL_ERROR` | 500 | Server error |

---

## 4. Konfigurasi CORS

API sudah dikonfigurasi untuk mengizinkan origin berikut:

```go
AllowOrigins: []string{
    "https://cms.armmada.id",
    "http://localhost:5173",    // Vite dev server
    "http://localhost:3000",    // Next.js dev server
}
```

**Allowed Methods:** `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`

**Allowed Headers:**
```go
"Origin", "Content-Type", "Accept", "Authorization", "X-Client-Type"
```

---

## 5. Contoh Kode Frontend

### 5.1 Setup Konfigurasi

**Environment Variables (.env):**
```env
VITE_API_BASE_URL=https://api-pos.armmada.id/v1
VITE_CMS_API_BASE_URL=https://api-cms.armmada.id/v1
```

### 5.2 TypeScript Types

```typescript
// types/api.ts

// Generic API Response
interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: { field: string; message: string }[];
  };
  idempotent?: boolean;
}

// Auth Types
interface User {
  id: number;
  username: string;
  name: string;
  role: 'ADMIN' | 'CASHIER';
}

interface LoginRequest {
  username: string;
  password: string;
}

interface LoginResponse {
  user: User;
  token: string;
}

// Product Types
interface Product {
  id: number;
  name: string;
  price: number;
  category: string;
  imageUrl: string;
  storeType: string;
  stockQuantity: number;
  lowStockThreshold: number;
  barcode: string;
  discount?: {
    id: number;
    name: string;
    type: string;
    value: number;
    status: string;
  };
  finalPrice: number;
  createdAt: string;
  updatedAt: string;
}

// Order Types
interface OrderItem {
  productId: number;
  productName: string;
  productPrice: number;
  qty: number;
  subtotal: number;
  variantId?: number;
  variantName?: string;
  sku?: string;
}

interface CreateOrderRequest {
  id: string;              // UUID
  items: OrderItem[];
  subtotal: number;
  tax: number;
  total: number;
  paymentMethod: 'CASH' | 'CARD' | 'QRIS';
  cashReceived?: number;
  changeAmount?: number;
  timestamp: number;       // Unix timestamp
  storeType?: string;
  terminalId?: string;
  customerId?: number;
  customerPhone?: string;
  customerName?: string;
}

interface Order {
  id: string;
  items: OrderItem[];
  subtotal: number;
  tax: number;
  total: number;
  paymentMethod: string;
  status: 'COMPLETED' | 'VOIDED';
  createdAt: string;
  // ... other fields
}

// Dashboard Types
interface DashboardSummary {
  totalOrders: number;
  totalRevenue: number;
  avgOrderValue: number;
  cashTotal: number;
  cardTotal: number;
  qrisTotal: number;
  voidedCount: number;
}

interface RevenuePoint {
  date: string;
  revenue: number;
  orders: number;
}

interface TopProduct {
  productId: number;
  productName: string;
  totalSold: number;
  revenue: number;
}

// Fuel Types
interface FuelPrice {
  code: string;
  name: string;
  pricePerLiter: number;
  updatedAt: number;
}

interface Pump {
  id: string;
  name: string;
  isActive: boolean;
}

interface Nozzle {
  id: string;
  pumpId: string;
  name: string;
  fuelCode: string;
  isActive: boolean;
}

interface FuelSale {
  id: string;
  pumpId: string;
  nozzleId: string;
  fuelCode: string;
  pricePerLiter: number;
  liters: number;
  totalAmount: number;
  paymentMethod: string;
  timestamp: string;
}
```

### 5.3 API Client Base Class

```typescript
// lib/api-client.ts

const API_BASE = import.meta.env.VITE_API_BASE_URL;
const CMS_API_BASE = import.meta.env.VITE_CMS_API_BASE_URL;

class ApiClient {
  private token: string | null = null;

  constructor(private baseUrl: string) {}

  setToken(token: string) {
    this.token = token;
  }

  clearToken() {
    this.token = null;
  }

  private async request<T>(
    method: string,
    endpoint: string,
    body?: unknown
  ): Promise<ApiResponse<T>> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
      });

      const data = await response.json();

      if (!response.ok) {
        throw data;
      }

      return data;
    } catch (error) {
      throw error;
    }
  }

  get<T>(endpoint: string) {
    return this.request<T>('GET', endpoint);
  }

  post<T>(endpoint: string, body: unknown) {
    return this.request<T>('POST', endpoint, body);
  }

  put<T>(endpoint: string, body: unknown) {
    return this.request<T>('PUT', endpoint, body);
  }

  patch<T>(endpoint: string, body?: unknown) {
    return this.request<T>('PATCH', endpoint, body);
  }

  delete<T>(endpoint: string) {
    return this.request<T>('DELETE', endpoint);
  }
}

// Instances
export const posApi = new ApiClient(API_BASE);
export const cmsApi = new ApiClient(CMS_API_BASE);

// Auth helper
export const auth = {
  async login(username: string, password: string) {
    const response = await posApi.post<LoginResponse>('/login', {
      username,
      password,
    });
    if (response.success && response.data) {
      posApi.setToken(response.data.token);
      localStorage.setItem('token', response.data.token);
      localStorage.setItem('user', JSON.stringify(response.data.user));
    }
    return response;
  },

  logout() {
    posApi.clearToken();
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  },

  restoreSession() {
    const token = localStorage.getItem('token');
    if (token) {
      posApi.setToken(token);
      return JSON.parse(localStorage.getItem('user') || 'null');
    }
    return null;
  },

  getUser(): User | null {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
  },

  isAuthenticated() {
    return !!localStorage.getItem('token');
  },

  hasRole(role: 'ADMIN' | 'CASHIER') {
    const user = this.getUser();
    return user?.role === role;
  },
};
```

### 5.4 Contoh Halaman Login (Svelte)

```svelte
<!-- routes/login/+page.svelte -->
<script lang="ts">
  import { auth } from '$lib/api-client';
  import { goto } from '$app/navigation';

  let username = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleLogin() {
    loading = true;
    error = '';

    try {
      const response = await auth.login(username, password);

      if (response.success) {
        goto('/dashboard');
      } else {
        error = response.error?.message || 'Login failed';
      }
    } catch (e) {
      error = 'Connection error. Please try again.';
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-100">
  <div class="bg-white p-8 rounded-lg shadow-md w-96">
    <h1 class="text-2xl font-bold mb-6 text-center">GBS POS Login</h1>

    {#if error}
      <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
        {error}
      </div>
    {/if}

    <form onsubmit={(e) => { e.preventDefault(); handleLogin(); }}>
      <div class="mb-4">
        <label class="block text-gray-700 text-sm font-bold mb-2" for="username">
          Username
        </label>
        <input
          id="username"
          type="text"
          bind:value={username}
          class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          required
        />
      </div>

      <div class="mb-6">
        <label class="block text-gray-700 text-sm font-bold mb-2" for="password">
          Password
        </label>
        <input
          id="password"
          type="password"
          bind:value={password}
          class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          required
        />
      </div>

      <button
        type="submit"
        disabled={loading}
        class="w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 disabled:opacity-50"
      >
        {loading ? 'Loading...' : 'Login'}
      </button>
    </form>

    <p class="mt-4 text-center text-sm text-gray-500">
      Demo: admin / admin123
    </p>
  </div>
</div>
```

### 5.5 Contoh Halaman Products

```svelte
<!-- routes/products/+page.svelte -->
<script lang="ts">
  import { posApi } from '$lib/api-client';
  import type { Product } from '$lib/types';

  let products = $state<Product[]>([]);
  let loading = $state(true);
  let error = $state('');

  async function loadProducts() {
    loading = true;
    error = '';

    try {
      const response = await posApi.get<{ products: Product[] }>('/products');

      if (response.success && response.data) {
        products = response.data.products || [];
      } else {
        error = response.error?.message || 'Failed to load products';
      }
    } catch (e) {
      error = 'Connection error';
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    loadProducts();
  });
</script>

<div class="p-6">
  <h1 class="text-2xl font-bold mb-6">Products</h1>

  {#if error}
    <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
      {error}
    </div>
  {/if}

  {#if loading}
    <p>Loading...</p>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {#each products as product}
        <div class="bg-white p-4 rounded-lg shadow">
          {#if product.imageUrl}
            <img src={product.imageUrl} alt={product.name} class="w-full h-32 object-cover mb-2" />
          {/if}
          <h3 class="font-bold">{product.name}</h3>
          <p class="text-gray-600">{product.category}</p>
          <p class="text-lg font-bold text-blue-600">
            Rp {product.finalPrice.toLocaleString('id-ID')}
          </p>
          {#if product.discount}
            <span class="text-sm text-red-500">
              -{product.discount.value}{product.discount.type === 'PERCENT' ? '%' : ''}
            </span>
          {/if}
          <p class="text-sm text-gray-500">
            Stock: {product.stockQuantity}
          </p>
        </div>
      {/each}
    </div>
  {/if}
</div>
```

### 5.6 Contoh Halaman Dashboard

```svelte
<!-- routes/dashboard/+page.svelte -->
<script lang="ts">
  import { posApi } from '$lib/api-client';
  import type { DashboardSummary, TopProduct, RevenuePoint } from '$lib/types';

  let summary = $state<DashboardSummary | null>(null);
  let topProducts = $state<TopProduct[]>([]);
  let revenueTrend = $state<RevenuePoint[]>([]);
  let loading = $state(true);

  async function loadDashboard() {
    loading = true;

    try {
      // Load summary
      const summaryRes = await posApi.get<DashboardSummary>('/dashboard/summary?range=today');
      if (summaryRes.success) summary = summaryRes.data;

      // Load top products
      const productsRes = await posApi.get<{ products: TopProduct[] }>('/dashboard/top-products?limit=5');
      if (productsRes.success) topProducts = productsRes.data?.products || [];

      // Load revenue trend
      const revenueRes = await posApi.get<{ trend: RevenuePoint[] }>('/dashboard/revenue?range=week');
      if (revenueRes.success) revenueTrend = revenueRes.data?.trend || [];

    } catch (e) {
      console.error('Dashboard load error:', e);
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    loadDashboard();
  });
</script>

<div class="p-6">
  <h1 class="text-2xl font-bold mb-6">Dashboard</h1>

  {#if loading}
    <p>Loading...</p>
  {:else if summary}
    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      <div class="bg-white p-4 rounded-lg shadow">
        <p class="text-gray-500">Total Orders</p>
        <p class="text-2xl font-bold">{summary.totalOrders}</p>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <p class="text-gray-500">Total Revenue</p>
        <p class="text-2xl font-bold text-green-600">
          Rp {summary.totalRevenue.toLocaleString('id-ID')}
        </p>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <p class="text-gray-500">Avg Order Value</p>
        <p class="text-2xl font-bold">
          Rp {summary.avgOrderValue.toLocaleString('id-ID')}
        </p>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <p class="text-gray-500">Voided Orders</p>
        <p class="text-2xl font-bold text-red-600">{summary.voidedCount}</p>
      </div>
    </div>

    <!-- Payment Breakdown -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
      <div class="bg-white p-4 rounded-lg shadow">
        <p class="text-gray-500">Cash</p>
        <p class="text-xl font-bold">Rp {summary.cashTotal.toLocaleString('id-ID')}</p>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <p class="text-gray-500">Card</p>
        <p class="text-xl font-bold">Rp {summary.cardTotal.toLocaleString('id-ID')}</p>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <p class="text-gray-500">QRIS</p>
        <p class="text-xl font-bold">Rp {summary.qrisTotal.toLocaleString('id-ID')}</p>
      </div>
    </div>

    <!-- Top Products -->
    <div class="bg-white p-4 rounded-lg shadow mb-8">
      <h2 class="text-xl font-bold mb-4">Top Products</h2>
      <table class="w-full">
        <thead>
          <tr class="border-b">
            <th class="text-left py-2">Product</th>
            <th class="text-right py-2">Sold</th>
            <th class="text-right py-2">Revenue</th>
          </tr>
        </thead>
        <tbody>
          {#each topProducts as product}
            <tr class="border-b">
              <td class="py-2">{product.productName}</td>
              <td class="text-right">{product.totalSold}</td>
              <td class="text-right">Rp {product.revenue.toLocaleString('id-ID')}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>
```

### 5.7 Contoh Membuat Order

```typescript
// lib/services/order-service.ts
import { posApi } from '$lib/api-client';
import type { CreateOrderRequest, Order } from '$lib/types';

export async function createOrder(orderData: Omit<CreateOrderRequest, 'id' | 'timestamp'>): Promise<{ success: boolean; data?: Order; error?: any }> {
  const order: CreateOrderRequest = {
    ...orderData,
    id: crypto.randomUUID(),  // Generate UUID
    timestamp: Date.now(),
  };

  const response = await posApi.post<Order>('/orders', order);
  return response;
}

export async function getOrder(orderId: string) {
  return posApi.get<Order>(`/orders/${orderId}`);
}

export async function voidOrder(orderId: string, reason: string) {
  // Requires ADMIN role
  return posApi.patch<Order>(`/orders/${orderId}/void`, { reason });
}

export async function getUnsettledSummary() {
  return posApi.get('/orders/unsettled/summary');
}
```

### 5.8 Contoh Fuel Operations

```typescript
// lib/services/fuel-service.ts
import { posApi } from '$lib/api-client';
import type { FuelPrice, Pump, Nozzle, FuelSale } from '$lib/types';

// Fuel Prices
export async function getFuelPrices() {
  return posApi.get<FuelPrice[]>('/fuel-prices');
}

export async function updateFuelPrice(code: string, pricePerLiter: number) {
  // Requires ADMIN role
  return posApi.put<FuelPrice>(`/fuel-prices/${code}`, { pricePerLiter });
}

// Pumps
export async function getPumps() {
  return posApi.get<Pump[]>('/pumps');
}

export async function createPump(id: string, name: string) {
  return posApi.post<Pump>('/pumps', { id, name });
}

// Nozzles
export async function getNozzles() {
  return posApi.get<Nozzle[]>('/nozzles');
}

export async function createNozzle(nozzle: { id: string; pumpId: string; name: string; fuelCode: string }) {
  return posApi.post<Nozzle>('/nozzles', nozzle);
}

// Fuel Sales
export async function createFuelSale(sale: {
  id: string;
  pumpId: string;
  nozzleId: string;
  fuelCode: string;
  pricePerLiter: number;
  liters: number;
  totalAmount: number;
  paymentMethod: string;
  transactionId?: string;
  timestamp?: number;
}) {
  return posApi.post<FuelSale>('/fuel-sales', {
    ...sale,
    timestamp: sale.timestamp || Date.now(),
  });
}

export async function getFuelReport(startDate?: string, endDate?: string) {
  const params = new URLSearchParams();
  if (startDate) params.set('startDate', startDate);
  if (endDate) params.set('endDate', endDate);
  return posApi.get(`/fuel-sales/report?${params}`);
}
```

---

## 6. Referensi API Lengkap

### 6.1 Authentication

| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| POST | `/v1/login` | No | Login user |

### 6.2 Products (POS API)

| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| GET | `/v1/products` | Yes | List all products |
| GET | `/v1/products/:id` | Yes | Get product by ID |
| GET | `/v1/products/barcode/:code` | Yes | Get product by barcode |
| GET | `/v1/products/low-stock` | Yes | Get low stock products |
| POST | `/v1/products` | ADMIN | Create product |
| PUT | `/v1/products/:id` | ADMIN | Update product |
| DELETE | `/v1/products/:id` | ADMIN | Delete product |
| POST | `/v1/products/import` | ADMIN | Import products (CSV) |
| GET | `/v1/products/export` | Yes | Export products (CSV) |
| GET | `/v1/products/:id/stock-history` | Yes | Get stock history |
| POST | `/v1/products/:id/stock` | ADMIN | Adjust stock |

### 6.3 Orders (POS API)

| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| GET | `/v1/orders` | Yes | List orders (filterable) |
| GET | `/v1/orders/:id` | Yes | Get order by ID |
| POST | `/v1/orders` | Yes | Create new order |
| POST | `/v1/sync/orders` | Yes | Bulk sync orders |
| PATCH | `/v1/orders/:id/void` | ADMIN | Void order |
| GET | `/v1/orders/unsettled/summary` | Yes | Get unsettled summary |
| POST | `/v1/orders/settle` | ADMIN | Settle orders |

### 6.4 Dashboard (POS API)

| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| GET | `/v1/dashboard/summary` | Yes | Dashboard summary |
| GET | `/v1/dashboard/revenue` | Yes | Revenue trend |
| GET | `/v1/dashboard/top-products` | Yes | Top selling products |

**Query Parameters for Dashboard:**
- `range`: `today` | `week` | `month` | `custom`
- `startDate`: ISO date (for custom range)
- `endDate`: ISO date (for custom range)
- `limit`: number (for top-products)

### 6.5 Customers (POS API)

| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| GET | `/v1/customers` | Yes | List customers |
| GET | `/v1/customers/:id` | Yes | Get customer details |
| GET | `/v1/customers/:id/orders` | Yes | Get customer orders |
| POST | `/v1/customers` | Yes | Create customer |
| PUT | `/v1/customers/:id` | Yes | Update customer |

### 6.6 Fuel Operations (POS API)

| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| GET | `/v1/fuel-prices` | Yes | List fuel prices |
| PUT | `/v1/fuel-prices/:code` | ADMIN | Update fuel price |
| GET | `/v1/pumps` | Yes | List pumps |
| POST | `/v1/pumps` | ADMIN | Create pump |
| PUT | `/v1/pumps/:id` | ADMIN | Update pump |
| DELETE | `/v1/pumps/:id` | ADMIN | Delete pump |
| GET | `/v1/nozzles` | Yes | List nozzles |
| POST | `/v1/nozzles` | ADMIN | Create nozzle |
| PUT | `/v1/nozzles/:id` | ADMIN | Update nozzle |
| DELETE | `/v1/nozzles/:id` | ADMIN | Delete nozzle |
| POST | `/v1/fuel-sales` | Yes | Record fuel sale |
| GET | `/v1/fuel-sales/report` | Yes | Get fuel sales report |

### 6.7 CMS API (Ads & Settings)

| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| POST | `/v1/ads/upload` | ADMIN | Upload ad |
| GET | `/v1/ads` | ADMIN | List ads |
| GET | `/v1/ads/:id` | ADMIN | Get ad |
| PUT | `/v1/ads/:id` | ADMIN | Update ad |
| DELETE | `/v1/ads/:id` | ADMIN | Delete ad |
| POST | `/v1/ads/:id/toggle` | ADMIN | Toggle ad status |
| GET | `/v1/ads/active` | No | Get active playlist |
| GET | `/v1/ads/download/:id` | No | Download ad file |
| POST | `/v1/ads/:id/play` | No | Log ad play |
| GET | `/v1/settings` | ADMIN | Get all settings |
| PUT | `/v1/settings` | ADMIN | Update settings |

### 6.8 Cart Display (CMS API)

| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| GET | `/v1/display/cart` | No | Get cart display (public) |
| POST | `/v1/display/cart` | Yes | Save cart display |
| GET | `/v1/display/terminals` | ADMIN | List terminals |

### 6.9 Settlements (POS API)

| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| GET | `/v1/settlements` | Yes | List settlements |
| GET | `/v1/settlements/:id` | Yes | Get settlement details |

---

## Quick Reference: Demo Credentials

| Username | Password | Role |
|----------|----------|------|
| `admin` | `admin123` | ADMIN |
| `cashier` | `cashier123` | CASHIER |

---

## Troubleshooting

### CORS Error
Pastikan frontend origin sudah ditambahkan di `gbs-common/middleware/cors.go`

### 401 Unauthorized
- Token expired → login ulang
- Token tidak diset → cek `auth.restoreSession()` di app startup

### 403 Forbidden
- User tidak punya role yang diperlukan
- ADMIN endpoints tidak bisa diakses dengan role CASHIER

### 422 Validation Error
- Cek `error.details` untuk info field yang bermasalah
- Pastikan semua field `required` sudah diisi

---

Dokumen ini dibuat berdasarkan analysis codebase GBS POS & CMS API.
