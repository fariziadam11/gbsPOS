// lib/router.ts - Simple Client-Side Router
import { writable, derived } from 'svelte/store';

export type Route = {
  path: string;
  component: any;
  meta?: {
    requiresAuth?: boolean;
    requiresAdmin?: boolean;
    title?: string;
  };
};

export const routes: Route[] = [
  { path: '/', redirect: '/dashboard' },
  { path: '/login', component: null, meta: { title: 'Login' } },
  { path: '/dashboard', component: null, meta: { requiresAuth: true, title: 'Dashboard' } },
  { path: '/products', component: null, meta: { requiresAuth: true, title: 'Products' } },
  { path: '/orders', component: null, meta: { requiresAuth: true, title: 'Orders' } },
  { path: '/ads', component: null, meta: { requiresAuth: true, requiresAdmin: true, title: 'Ads' } },
  { path: '/settings', component: null, meta: { requiresAuth: true, requiresAdmin: true, title: 'Settings' } },
];

// Router store
function createRouter() {
  const currentPath = writable(typeof window !== 'undefined' ? window.location.pathname : '/');
  const currentRoute = writable<Route | null>(null);

  function matchRoute(path: string) {
    // Exact match first
    let route = routes.find(r => r.path === path && !r.path.includes(':'));
    if (route) {
      currentRoute.set(route);
      return;
    }

    // Dynamic route matching (e.g., /products/:id)
    for (const route of routes) {
      if (route.path.includes(':')) {
        const pattern = route.path.replace(/:[^/]+/g, '[^/]+');
        const regex = new RegExp(`^${pattern}$`);
        if (regex.test(path)) {
          currentRoute.set(route);
          return;
        }
      }
    }

    // 404 - redirect to dashboard
    const dashboardRoute = routes.find(r => r.path === '/dashboard') || null;
    currentRoute.set(dashboardRoute);
  }

  function navigate(path: string, replace = false) {
    if (typeof window === 'undefined') return;

    if (replace) {
      history.replaceState(null, '', path);
    } else {
      history.pushState(null, '', path);
    }
    currentPath.set(path);
    matchRoute(path);
  }

  function getParams(): Record<string, string> {
    let route: Route | null = null;
    let path = '';

    currentRoute.subscribe(r => route = r)();
    currentPath.subscribe(p => path = p)();

    if (!route || !route.path.includes(':')) return {};

    const params: Record<string, string> = {};
    const routeParts = route.path.split('/');
    const pathParts = path.split('/');

    routeParts.forEach((part, i) => {
      if (part.startsWith(':')) {
        params[part.slice(1)] = pathParts[i] || '';
      }
    });

    return params;
  }

  function handlePopState() {
    if (typeof window === 'undefined') return;
    const path = window.location.pathname;
    currentPath.set(path);
    matchRoute(path);
  }

  // Initialize
  if (typeof window !== 'undefined') {
    window.addEventListener('popstate', handlePopState);
    let initialPath = window.location.pathname;
    currentPath.set(initialPath);
    matchRoute(initialPath);
  }

  return {
    currentPath,
    currentRoute,
    navigate,
    getParams,
  };
}

export const router = createRouter();

// Navigate helper - export for convenience
export function navigate(path: string, replace = false) {
  router.navigate(path, replace);
}

// Redirect helper
export function redirect(path: string) {
  router.navigate(path, true);
}
