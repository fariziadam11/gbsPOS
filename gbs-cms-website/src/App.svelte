<script lang="ts">
  import { onMount } from 'svelte';
  import { router } from './lib/router';
  import { authStore } from './lib/stores/auth';
  import Layout from './lib/components/Layout.svelte';
  import Toast from './lib/components/Toast.svelte';
  import LoginPage from './routes/login/LoginPage.svelte';
  import DashboardPage from './routes/dashboard/DashboardPage.svelte';
  import ProductsPage from './routes/products/ProductsPage.svelte';
  import OrdersPage from './routes/orders/OrdersPage.svelte';
  import AdsPage from './routes/ads/AdsPage.svelte';
  import SettingsPage from './routes/settings/SettingsPage.svelte';

  let ready = $state(false);

  // Subscribe to router path changes
  let currentPath = $state('/dashboard');

  onMount(() => {
    // Initialize auth
    authStore.initialize();
    ready = true;

    // Subscribe to router changes
    const unsubPath = router.currentPath.subscribe(path => {
      currentPath = path;
      checkAuth();
    });

    return () => {
      unsubPath();
    };
  });

  function checkAuth() {
    const route = router.currentRoute;
    if (!route) return;

    const requiresAuth = route.meta?.requiresAuth;
    const requiresAdmin = route.meta?.requiresAdmin;

    // Use store values
    let isAuth = false;
    let isAdmin = false;
    authStore.isAuthenticated.subscribe(v => isAuth = v)();
    authStore.isAdmin.subscribe(v => isAdmin = v)();

    // Redirect to login if auth required but not authenticated
    if (requiresAuth && !isAuth) {
      router.navigate('/login', true);
      return;
    }

    // Redirect to dashboard if admin required but not admin
    if (requiresAdmin && !isAdmin) {
      router.navigate('/dashboard', true);
      return;
    }

    // Redirect to dashboard if authenticated and trying to access login
    if (currentPath === '/login' && isAuth) {
      router.navigate('/dashboard', true);
    }
  }

  // Determine current page component
  let currentPage = $derived.by(() => {
    if (!authStore.isAuthenticated) {
      return LoginPage;
    }

    if (currentPath.startsWith('/login')) {
      return LoginPage;
    }

    if (currentPath.startsWith('/dashboard')) {
      return DashboardPage;
    }

    if (currentPath.startsWith('/products')) {
      return ProductsPage;
    }

    if (currentPath.startsWith('/orders')) {
      return OrdersPage;
    }

    if (currentPath.startsWith('/ads')) {
      return AdsPage;
    }

    if (currentPath.startsWith('/settings')) {
      return SettingsPage;
    }

    return DashboardPage;
  });

  let showLayout = $derived(
    authStore.isAuthenticated && currentPath !== '/login'
  );

  // Get page title
  let pageTitle = $derived.by(() => {
    let route = null;
    router.currentRoute.subscribe(r => route = r)();
    return route?.meta?.title || 'Dashboard';
  });
</script>

<Toast />

{#if !ready}
  <div class="min-h-screen flex items-center justify-center bg-slate-50">
    <div class="animate-spin w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
  </div>
{:else if showLayout}
  <Layout>
    {#if currentPath.startsWith('/dashboard')}
      <DashboardPage />
    {:else if currentPath.startsWith('/products')}
      <ProductsPage />
    {:else if currentPath.startsWith('/orders')}
      <OrdersPage />
    {:else if currentPath.startsWith('/ads')}
      <AdsPage />
    {:else if currentPath.startsWith('/settings')}
      <SettingsPage />
    {:else}
      <DashboardPage />
    {/if}
  </Layout>
{:else}
  <LoginPage />
{/if}
