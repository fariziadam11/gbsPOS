export interface ApiResponse<T> {
  success: boolean;
  data: T;
  idempotent?: boolean;
}

export interface ApiErrorDetail {
  field: string;
  message: string;
}

export interface ApiErrorResponse {
  success: false;
  error: {
    code: string;
    message: string;
    details?: ApiErrorDetail[];
  };
}

export interface User {
  id: number;
  username: string;
  name: string;
  role: string;
  createdAt: string;
  updatedAt: string;
}

export interface Ad {
  id: number;
  name: string;
  filename: string;
  storagePath: string;
  fileSize: number;
  mimeType: string;
  durationSeconds: number | null;
  storeTypes: string[];
  playlistOrder: number;
  isActive: boolean;
  startDate: string | null;
  endDate: string | null;
  startTime: string | null;
  endTime: string | null;
  createdBy: number;
  createdAt: string;
  updatedAt: string;
}

export interface Pagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface AdListResponse {
  ads: Ad[];
  pagination: Pagination;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  token: string;
}

export interface CreateAdRequest {
  file: File;
  name: string;
  storeTypes: string[];
  playlistOrder?: number;
  startDate?: string;
  endDate?: string;
  startTime?: string;
  endTime?: string;
}

export interface UpdateAdRequest {
  name?: string;
  storeTypes?: string[];
  playlistOrder?: number;
  isActive?: boolean;
  startDate?: string | null;
  endDate?: string | null;
  startTime?: string | null;
  endTime?: string | null;
}

export interface PlaylistResponse {
  playlist: Ad[];
  updatedAt: string;
}

export interface ToggleAdResponse {
  id: number;
  isActive: boolean;
  updatedAt: string;
}

// Product types
export interface Product {
  id: number;
  name: string;
  price: number;
  category: string;
  imageUrl: string;
  storeType: string;
  stockQuantity: number;
  lowStockThreshold: number;
  createdAt: string;
  updatedAt: string;
}

export interface CreateProductRequest {
  name: string;
  price: number;
  category: string;
  imageUrl?: string;
  storeType: string;
  stockQuantity?: number;
  lowStockThreshold?: number;
}

export interface UpdateProductRequest {
  name?: string;
  price?: number;
  category?: string;
  imageUrl?: string;
  storeType?: string;
  stockQuantity?: number;
  lowStockThreshold?: number;
}

export interface ImportResult {
  success: number;
  failed: number;
  errors: string[];
}

// Order types
export interface OrderItem {
  id: number;
  productId: number;
  productName: string;
  productPrice: number;
  qty: number;
  subtotal: number;
}

export interface Order {
  id: string;
  items: OrderItem[];
  subtotal: number;
  tax: number;
  total: number;
  paymentMethod: string;
  cashReceived: number | null;
  changeAmount: number | null;
  discountType: string | null;
  discountValue: number | null;
  discountAmount: number | null;
  timestamp: number;
  isVoided: boolean;
  isSettled: boolean;
  transactionId: string | null;
  approvalCode: string | null;
  entryMode: string | null;
  maskedAccount: string | null;
  acqMid: string | null;
  acqTid: string | null;
  posMessageId: string | null;
  bankName: string | null;
  storeType: string;
  terminalId: string | null;
  voidReason: string | null;
  voidedBy: string | null;
  voidedAt: string | null;
  customerId: number | null;
  customerPhone: string;
  customerName: string;
  loyaltyPointsEarned: number;
  createdAt: string;
  updatedAt: string;
}

// Customer types
export interface Customer {
  id: number;
  name: string;
  phone: string;
  email: string | null;
  address: string | null;
  loyaltyPoints: number;
  createdAt: string;
  updatedAt: string;
}

export interface CustomerDetail {
  customer: Customer;
  orderHistory: Order[];
  totalSpent: number;
  totalOrders: number;
}

export interface CreateCustomerRequest {
  name: string;
  phone: string;
  email?: string;
  address?: string;
}

export interface UpdateCustomerRequest {
  name?: string;
  phone?: string;
  email?: string;
  address?: string;
}

// Dashboard types
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

// Settings types
export interface SettingsResponse {
  settings: Record<string, string>;
}

export interface UpdateSettingsRequest {
  settings: Record<string, string>;
}

// User management types
export interface CreateUserRequest {
  username: string;
  password: string;
  name: string;
  role: string;
  gender?: string;
}

export interface UpdateUserRequest {
  name?: string;
  role?: string;
  password?: string;
  gender?: string;
}

// Fuel types
export interface FuelPrice {
  code: string;
  name: string;
  pricePerLiter: number;
  updatedAt: number;
}

export interface UpdateFuelPriceRequest {
  pricePerLiter: number;
}

export interface Pump {
  id: string;
  name: string;
  isActive: boolean;
}

export interface CreatePumpRequest {
  id: string;
  name: string;
}

export interface UpdatePumpRequest {
  name?: string;
  isActive?: boolean;
}

export interface Nozzle {
  id: string;
  pumpId: string;
  name: string;
  fuelCode: string;
  isActive: boolean;
}

export interface CreateNozzleRequest {
  id: string;
  pumpId: string;
  name: string;
  fuelCode: string;
}

export interface UpdateNozzleRequest {
  name?: string;
  fuelCode?: string;
  isActive?: boolean;
}

export interface FuelReportItem {
  fuelCode: string;
  liters: number;
  totalAmount: number;
}

export interface PumpReportItem {
  pumpId: string;
  liters: number;
  totalAmount: number;
}

export interface FuelSalesReport {
  summary: FuelReportItem[];
  pumpTotals: PumpReportItem[];
}

export interface UserListItem {
  id: number;
  username: string;
  name: string;
  role: string;
  gender: string;
  createdAt: string;
  updatedAt: string;
}
