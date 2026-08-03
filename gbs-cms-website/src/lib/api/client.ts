// lib/api/client.ts - API Client

import type { ApiResponse } from '../types/api';

const POS_API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/v1';
const CMS_API_BASE = import.meta.env.VITE_CMS_API_BASE_URL || 'http://localhost:8081/v1';

// Shared token state
let sharedToken: string | null = null;

export function setSharedToken(token: string | null) {
  sharedToken = token;
}

export function getSharedToken(): string | null {
  if (sharedToken) return sharedToken;
  // Fallback to localStorage
  return localStorage.getItem('token');
}

export class ApiClient {
  constructor(private baseUrl: string) {}

  private async request<T>(
    method: string,
    endpoint: string,
    body?: unknown
  ): Promise<ApiResponse<T>> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    // Use shared token
    const token = getSharedToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
      });

      const data = await response.json();

      if (!response.ok) {
        return data;
      }

      return data;
    } catch (error) {
      console.error('API Error:', error);
      return {
        success: false,
        error: {
          code: 'NETWORK_ERROR',
          message: error instanceof Error ? error.message : 'Network error occurred',
        },
      };
    }
  }

  async get<T>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>('GET', endpoint);
  }

  async post<T>(endpoint: string, body: unknown): Promise<ApiResponse<T>> {
    return this.request<T>('POST', endpoint, body);
  }

  async put<T>(endpoint: string, body: unknown): Promise<ApiResponse<T>> {
    return this.request<T>('PUT', endpoint, body);
  }

  async patch<T>(endpoint: string, body?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>('PATCH', endpoint, body);
  }

  async delete<T>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>('DELETE', endpoint);
  }
}

// Instances
export const posApi = new ApiClient(POS_API_BASE);
export const cmsApi = new ApiClient(CMS_API_BASE);

// Auth helpers
export const auth = {
  async login(username: string, password: string) {
    const response = await posApi.post<{ user: any; token: string }>('/login', {
      username,
      password,
    });

    if (response.success && response.data) {
      const token = response.data.token;
      setSharedToken(token);
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(response.data.user));
    }

    return response;
  },

  logout() {
    setSharedToken(null);
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  },

  restoreSession(): boolean {
    const token = getSharedToken();
    return !!token;
  },

  getUser<T>(): T | null {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
  },

  isAuthenticated(): boolean {
    return !!getSharedToken();
  },

  hasRole(role: string): boolean {
    const user = this.getUser<{ role: string }>();
    return user?.role === role;
  },
};

// Initialize token from localStorage on load
setSharedToken(localStorage.getItem('token'));
