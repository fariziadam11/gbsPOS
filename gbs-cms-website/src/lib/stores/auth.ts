// lib/stores/auth.ts - Auth Store
import { writable, derived } from 'svelte/store';
import { posApi, cmsApi, auth as authClient, setSharedToken } from '../api/client';
import type { User } from '../types/api';

// Create a reactive auth state using Svelte stores
function createAuthStore() {
  const user = writable<User | null>(null);
  const isAuthenticated = writable(false);
  const isLoading = writable(true);

  function initialize() {
    const token = localStorage.getItem('token');
    const storedUser = localStorage.getItem('user');

    if (token && storedUser) {
      setSharedToken(token);
      try {
        user.set(JSON.parse(storedUser));
        isAuthenticated.set(true);
      } catch {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
      }
    }
    isLoading.set(false);
  }

  async function login(username: string, password: string) {
    isLoading.set(true);
    try {
      const response = await authClient.login(username, password);

      if (response.success && response.data) {
        user.set(response.data.user);
        isAuthenticated.set(true);
        return { success: true, data: response.data };
      }

      return { success: false, error: response.error };
    } finally {
      isLoading.set(false);
    }
  }

  function logout() {
    authClient.logout();
    user.set(null);
    isAuthenticated.set(false);
  }

  return {
    user,
    isAuthenticated,
    isLoading,
    isAdmin: derived(user, ($user) => $user?.role === 'ADMIN'),
    initialize,
    login,
    logout,
  };
}

export const authStore = createAuthStore();
