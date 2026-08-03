// types/api.ts - API Types

// Generic API Response
export interface ApiResponse<T> {
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
export interface User {
  id: number;
  username: string;
  name: string;
  role: 'ADMIN' | 'CASHIER';
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  token: string;
}

// Product Types
export interface Product {
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

export interface CreateProductRequest {
  name: string;
  price: number;
  category: string;
  storeType?: string;
  barcode?: string;
  stockQuantity?: number;
  lowStockThreshold?: number;
  imageUrl?: string;
}

// Order Types
export interface OrderItem {
  productId: number;
  productName: string;
  productPrice: number;
  qty: number;
  subtotal: number;
  variantId?: number;
  variantName?: string;
  sku?: string;
}

export interface CreateOrderRequest {
  id: string;
  items: OrderItem[];
  subtotal: number;
  tax: number;
  total: number;
  paymentMethod: 'CASH' | 'CARD' | 'QRIS';
  cashReceived?: number;
  changeAmount?: number;
  timestamp: number;
  storeType?: string;
  terminalId?: string;
  customerId?: number;
  customerPhone?: string;
  customerName?: string;
}

export interface Order {
  id: string;
  items: OrderItem[];
  subtotal: number;
  tax: number;
  total: number;
  paymentMethod: string;
  status: 'COMPLETED' | 'VOIDED';
  cashReceived?: number;
  changeAmount?: number;
  createdAt: string;
  createdBy?: string;
}

export interface VoidOrderRequest {
  reason: string;
}

// Dashboard Types
export interface DashboardSummary {
  totalOrders: number;
  totalRevenue: number;
  avgOrderValue: number;
  cashTotal: number;
  cardTotal: number;
  qrisTotal: number;
  voidedCount: number;
}

export interface RevenuePoint {
  date: string;
  revenue: number;
  orders: number;
}

export interface TopProduct {
  productId: number;
  productName: string;
  totalSold: number;
  revenue: number;
}

// Fuel Types
export interface FuelPrice {
  code: string;
  name: string;
  pricePerLiter: number;
  updatedAt: number;
}

export interface Pump {
  id: string;
  name: string;
  isActive: boolean;
}

export interface Nozzle {
  id: string;
  pumpId: string;
  name: string;
  fuelCode: string;
  isActive: boolean;
}

export interface FuelSale {
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

// CMS Types
export interface Ad {
  id: number;
  name: string;
  fileUrl: string;
  fileType: string;
  duration: number;
  storeType: string;
  status: 'ACTIVE' | 'INACTIVE';
  playCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface Setting {
  key: string;
  value: string;
  type: string;
}

// Customer Types
export interface Customer {
  id: number;
  name: string;
  phone: string;
  email?: string;
  totalOrders: number;
  totalSpent: number;
  createdAt: string;
}
