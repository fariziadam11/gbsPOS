# GBS CMS Website

A modern Content Management System for GBS POS. Built with Svelte 5, Vite, and TypeScript.

## Tech Stack

- **Framework**: Svelte 5 with Runes
- **Build Tool**: Vite 8
- **Runtime**: Bun
- **Styling**: Tailwind CSS v4
- **Icons**: Lucide Svelte
- **Language**: TypeScript

## Prerequisites

- [Bun](https://bun.sh/) installed
- GBS POS API running on port 8080
- GBS CMS API running on port 8081

## Getting Started

### Install Dependencies

```bash
bun install
```

### Development

```bash
bun run dev
```

The app will be available at `http://localhost:5173`

### Production Build

```bash
bun run build
```

Preview the production build:

```bash
bun run preview
```

## Configuration

Create a `.env` file in the root directory:

```env
# API Configuration
VITE_API_BASE_URL=http://localhost:8080/v1
VITE_CMS_API_BASE_URL=http://localhost:8081/v1

# Keycloak (optional - for production)
VITE_KEYCLOAK_BASE_URL=https://auth.armmada.id
VITE_KEYCLOAK_REALM=gbs-pos
VITE_KEYCLOAK_CLIENT_ID=gbs-cms-web
```

## Project Structure

```
src/
├── lib/
│   ├── api/
│   │   └── client.ts          # API client with auth handling
│   ├── stores/
│   │   └── auth.ts            # Auth store (Svelte 5 runes)
│   ├── types/
│   │   └── api.ts             # TypeScript types
│   ├── components/
│   │   ├── Layout.svelte     # Main layout with sidebar
│   │   ├── Toast.svelte      # Toast notifications
│   │   ├── Loading.svelte    # Loading spinner
│   │   └── EmptyState.svelte # Empty state component
│   └── router.ts             # Simple SPA router
├── routes/
│   ├── login/                # Login page
│   ├── dashboard/            # Dashboard page
│   ├── products/             # Products management
│   ├── orders/               # Orders management
│   ├── ads/                  # Ads management (ADMIN)
│   └── settings/             # Settings page (ADMIN)
├── App.svelte               # Main app component
├── main.ts                  # Entry point
└── app.css                  # Global styles
```

## Features

### Authentication
- JWT-based authentication
- Role-based access (ADMIN / CASHIER)
- Auto session restore
- Demo credentials: `admin/admin123` or `cashier/cashier123`

### Dashboard
- Sales summary cards
- Revenue trend
- Top selling products
- Payment breakdown (Cash/Card/QRIS)

### Products (ADMIN)
- List, create, edit, delete products
- Search and filter
- Low stock alerts
- Barcode support

### Orders
- View order history
- Order detail modal
- Void orders (ADMIN only)

### Ads Management (ADMIN)
- Upload video/image ads
- Toggle ad status
- Set duration and store type

### Settings (ADMIN)
- Store configuration
- Tax & pricing settings
- Display settings
- Notification preferences

## Deployment

### Build for Production

```bash
bun run build
```

The output will be in the `dist/` folder.

### Static Hosting

Upload the contents of `dist/` to any static hosting:

- **Vercel**: `vercel deploy`
- **Netlify**: Drag `dist/` to dashboard
- **Cloudflare Pages**: Connect to Git repo
- **Nginx**: Point to `dist/` folder

### Environment Variables for Production

Set these in your hosting provider:

```env
VITE_API_BASE_URL=https://api-pos.armmada.id/v1
VITE_CMS_API_BASE_URL=https://api-cms.armmada.id/v1
```

## API Endpoints

### POS API (port 8080)
- `/v1/login` - Authentication
- `/v1/dashboard/*` - Dashboard data
- `/v1/products/*` - Products CRUD
- `/v1/orders/*` - Orders management
- `/v1/fuel-prices/*` - Fuel pricing

### CMS API (port 8081)
- `/v1/ads/*` - Ads management
- `/v1/settings` - Settings management

## License

Private - All rights reserved
