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
