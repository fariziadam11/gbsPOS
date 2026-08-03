<script lang="ts">
  import { onMount } from 'svelte';
  import { LayoutDashboard, Package, ShoppingCart, Image, Settings, LogOut, Menu, X, ChevronLeft } from 'lucide-svelte';
  import { router, navigate } from '../router';
  import { authStore } from '../stores/auth';

  interface Props {
    children: any;
  }

  let { children }: Props = $props();

  let sidebarOpen = $state(true);
  let mobileMenuOpen = $state(false);
  let currentPath = $state('/dashboard');
  let currentUser = $state<{ name: string; role: string } | null>(null);
  let isAdmin = $state(false);
  let pageTitle = $state('Dashboard');

  const navItems = [
    { path: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
    { path: '/products', label: 'Products', icon: Package },
    { path: '/orders', label: 'Orders', icon: ShoppingCart },
    { path: '/ads', label: 'Ads', icon: Image, adminOnly: true },
    { path: '/settings', label: 'Settings', icon: Settings, adminOnly: true },
  ];

  onMount(() => {
    const unsubPath = router.currentPath.subscribe(path => {
      currentPath = path;
      // Close mobile menu on navigation
      mobileMenuOpen = false;
    });

    const unsubRoute = router.currentRoute.subscribe(route => {
      if (route?.meta?.title) {
        pageTitle = route.meta.title;
      }
    });

    const unsubUser = authStore.user.subscribe(user => {
      currentUser = user;
    });

    const unsubAdmin = authStore.isAdmin.subscribe(admin => {
      isAdmin = admin;
    });

    // Auto-close sidebar on small screens
    const handleResize = () => {
      if (window.innerWidth < 768) {
        sidebarOpen = true; // keep full when opening mobile overlay
      }
    };
    window.addEventListener('resize', handleResize);

    return () => {
      unsubPath();
      unsubRoute();
      unsubUser();
      unsubAdmin();
      window.removeEventListener('resize', handleResize);
    };
  });

  function handleLogout() {
    authStore.logout();
    navigate('/login', true);
  }

  function isActive(path: string): boolean {
    return currentPath.startsWith(path);
  }

  function shouldShowNav(item: { adminOnly?: boolean }): boolean {
    if (item.adminOnly) return isAdmin;
    return true;
  }

  function getUserInitials(name: string): string {
    return name
      .split(' ')
      .map(n => n.charAt(0))
      .slice(0, 2)
      .join('')
      .toUpperCase();
  }
</script>

<div class="min-h-screen flex bg-slate-50">
  <!-- =====================
       Sidebar
       ===================== -->
  <aside
    class="flex flex-col flex-shrink-0 transition-all duration-300 ease-in-out z-50
           {sidebarOpen ? 'w-64' : 'w-[70px]'}
           fixed md:sticky top-0 h-screen
           {mobileMenuOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'}"
    style="
      background: linear-gradient(180deg, #0f172a 0%, #1e293b 60%, #1e293b 100%);
      box-shadow: 4px 0 24px rgba(0,0,0,0.15);
    "
  >
    <!-- Logo / Brand -->
    <div class="flex items-center justify-between px-4 py-4 border-b border-white/10 flex-shrink-0"
         style="min-height: 64px;">
      <div class="flex items-center gap-3 overflow-hidden">
        <div class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0"
             style="background: linear-gradient(135deg, #3b82f6, #6366f1);">
          <span class="text-white font-bold text-sm">G</span>
        </div>
        {#if sidebarOpen}
          <div class="overflow-hidden whitespace-nowrap">
            <p class="text-white font-bold text-base leading-tight">GBS CMS</p>
            <p class="text-blue-400 text-xs font-medium">Point of Sale</p>
          </div>
        {/if}
      </div>
      <!-- Close button on mobile -->
      <button
        class="md:hidden p-1.5 hover:bg-white/10 rounded-lg text-slate-400 hover:text-white transition-colors flex-shrink-0"
        onclick={() => mobileMenuOpen = false}
        aria-label="Close menu"
      >
        <X size={18} />
      </button>
    </div>

    <!-- Navigation -->
    <nav class="flex-1 py-4 px-3 space-y-1 overflow-y-auto overflow-x-hidden">
      {#each navItems as item}
        {#if shouldShowNav(item)}
          <a
            href={item.path}
            onclick={(e) => { e.preventDefault(); navigate(item.path); }}
            title={!sidebarOpen ? item.label : ''}
            class="relative flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium
                   transition-all duration-150 group no-underline
                   {isActive(item.path)
                     ? 'text-white'
                     : 'text-slate-400 hover:text-white hover:bg-white/10'}"
            style={isActive(item.path)
              ? 'background: linear-gradient(135deg, rgba(59,130,246,0.9), rgba(99,102,241,0.8)); box-shadow: 0 2px 10px rgba(59,130,246,0.35);'
              : ''}
          >
            {#if item.icon === LayoutDashboard}
              <LayoutDashboard size={18} class="flex-shrink-0" />
            {:else if item.icon === Package}
              <Package size={18} class="flex-shrink-0" />
            {:else if item.icon === ShoppingCart}
              <ShoppingCart size={18} class="flex-shrink-0" />
            {:else if item.icon === Image}
              <Image size={18} class="flex-shrink-0" />
            {:else if item.icon === Settings}
              <Settings size={18} class="flex-shrink-0" />
            {/if}
            {#if sidebarOpen}
              <span class="whitespace-nowrap">{item.label}</span>
            {/if}
            <!-- Tooltip when collapsed -->
            {#if !sidebarOpen}
              <div class="absolute left-full ml-2 px-2 py-1 bg-slate-900 text-white text-xs rounded-md
                          opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity duration-150
                          whitespace-nowrap z-50 border border-white/10">
                {item.label}
              </div>
            {/if}
          </a>
        {/if}
      {/each}
    </nav>

    <!-- User & Logout -->
    <div class="px-3 py-4 border-t border-white/10 flex-shrink-0">
      {#if sidebarOpen && currentUser}
        <div class="flex items-center gap-3 px-2 py-2 mb-2">
          <div class="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold flex-shrink-0"
               style="background: linear-gradient(135deg, #3b82f6, #6366f1);">
            {getUserInitials(currentUser.name)}
          </div>
          <div class="overflow-hidden min-w-0">
            <p class="text-white text-sm font-medium leading-tight truncate">{currentUser.name}</p>
            <p class="text-slate-400 text-xs capitalize">{currentUser.role}</p>
          </div>
        </div>
      {:else if !sidebarOpen && currentUser}
        <div class="flex justify-center mb-2">
          <div class="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold"
               title={currentUser.name}
               style="background: linear-gradient(135deg, #3b82f6, #6366f1);">
            {getUserInitials(currentUser.name)}
          </div>
        </div>
      {/if}
      <button
        onclick={handleLogout}
        class="flex items-center gap-3 w-full px-3 py-2.5 rounded-xl text-slate-400
               hover:bg-red-500/20 hover:text-red-400 transition-all duration-150 text-sm font-medium"
        title={!sidebarOpen ? 'Logout' : ''}
      >
        <LogOut size={18} class="flex-shrink-0" />
        {#if sidebarOpen}
          <span>Logout</span>
        {/if}
      </button>
    </div>
  </aside>

  <!-- Mobile Overlay (backdrop) -->
  {#if mobileMenuOpen}
    <button
      class="fixed inset-0 z-40 md:hidden"
      style="background: rgba(0,0,0,0.6); backdrop-filter: blur(3px);"
      onclick={() => mobileMenuOpen = false}
      aria-label="Close menu"
    ></button>
  {/if}

  <!-- =====================
       Main Content
       ===================== -->
  <div class="flex-1 flex flex-col min-w-0 transition-all duration-300"
       style="margin-left: 0;">
    <!-- Top Bar -->
    <header class="sticky top-0 z-30 flex items-center gap-3 px-4 md:px-6 bg-white border-b border-slate-200"
            style="min-height: 64px; box-shadow: 0 1px 3px rgba(0,0,0,0.06);">
      <!-- Mobile hamburger -->
      <button
        class="md:hidden p-2 hover:bg-slate-100 rounded-lg transition-colors text-slate-600"
        onclick={() => mobileMenuOpen = true}
        aria-label="Open menu"
      >
        <Menu size={20} />
      </button>

      <!-- Desktop collapse toggle -->
      <button
        class="hidden md:flex p-2 hover:bg-slate-100 rounded-lg transition-colors text-slate-500 hover:text-slate-700"
        onclick={() => sidebarOpen = !sidebarOpen}
        aria-label="Toggle sidebar"
      >
        {#if sidebarOpen}
          <ChevronLeft size={20} />
        {:else}
          <Menu size={20} />
        {/if}
      </button>

      <!-- Page Title -->
      <div class="flex items-center gap-2">
        <h2 class="text-base md:text-lg font-semibold text-slate-800">{pageTitle}</h2>
      </div>

      <div class="flex-1"></div>

      <!-- User Info (topbar) -->
      {#if currentUser}
        <div class="flex items-center gap-3">
          <div class="text-right hidden sm:block">
            <p class="text-sm font-semibold text-slate-800 leading-tight">{currentUser.name}</p>
            <p class="text-xs text-slate-400 capitalize">{currentUser.role}</p>
          </div>
          <div class="w-9 h-9 rounded-full flex items-center justify-center text-white text-sm font-bold flex-shrink-0"
               style="background: linear-gradient(135deg, #3b82f6, #6366f1);">
            {getUserInitials(currentUser.name)}
          </div>
        </div>
      {/if}
    </header>

    <!-- Page Content -->
    <main class="flex-1 p-4 md:p-6 overflow-auto bg-slate-50 page-enter">
      {@render children()}
    </main>
  </div>
</div>
