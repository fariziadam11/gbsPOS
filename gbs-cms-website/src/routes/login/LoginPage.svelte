<script lang="ts">
  import { Lock, User, AlertCircle, Eye, EyeOff, ArrowRight } from 'lucide-svelte';
  import { authStore } from '../../lib/stores/auth';
  import { navigate } from '../../lib/router';

  let username = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);
  let showPassword = $state(false);

  async function handleLogin(e: Event) {
    e.preventDefault();
    loading = true;
    error = '';

    try {
      const result = await authStore.login(username, password);

      if (result.success) {
        navigate('/dashboard');
      } else {
        error = result.error?.message || 'Login failed. Please check your credentials.';
      }
    } catch (err) {
      error = 'Connection error. Please try again.';
    } finally {
      loading = false;
    }
  }

  function fillCredentials(type: 'admin' | 'cashier') {
    if (type === 'admin') {
      username = 'admin';
      password = 'admin123';
    } else {
      username = 'cashier';
      password = 'cashier123';
    }
  }
</script>

<svelte:head>
  <title>GBS CMS — Sign In</title>
</svelte:head>

<div class="min-h-screen flex login-bg relative overflow-hidden">
  <!-- Decorative orbs -->
  <div class="absolute top-[-10%] right-[-5%] w-72 h-72 rounded-full opacity-20 pointer-events-none"
       style="background: radial-gradient(circle, #60a5fa, transparent); filter: blur(40px);"></div>
  <div class="absolute bottom-[-10%] left-[-5%] w-96 h-96 rounded-full opacity-15 pointer-events-none"
       style="background: radial-gradient(circle, #818cf8, transparent); filter: blur(60px);"></div>

  <!-- Left branding panel (hidden on mobile) -->
  <div class="hidden lg:flex flex-1 flex-col items-center justify-center p-12 relative">
    <div class="max-w-sm text-center">
      <div class="w-20 h-20 rounded-2xl flex items-center justify-center mx-auto mb-6"
           style="background: rgba(255,255,255,0.15); backdrop-filter: blur(10px); border: 1px solid rgba(255,255,255,0.2);">
        <span class="text-white font-black text-4xl">G</span>
      </div>
      <h1 class="text-4xl font-black text-white mb-4 leading-tight">GBS CMS</h1>
      <p class="text-blue-200 text-lg leading-relaxed">
        Centralized management system for your Point-of-Sale operations.
      </p>

      <!-- Feature dots -->
      <div class="mt-10 flex flex-col gap-3 text-left">
        {#each ['Real-time sales dashboard', 'Product & inventory management', 'Digital ads management', 'Multi-store support'] as feature}
          <div class="flex items-center gap-3">
            <div class="w-5 h-5 rounded-full flex items-center justify-center flex-shrink-0"
                 style="background: rgba(255,255,255,0.2);">
              <div class="w-2 h-2 rounded-full bg-white"></div>
            </div>
            <span class="text-blue-100 text-sm">{feature}</span>
          </div>
        {/each}
      </div>
    </div>
  </div>

  <!-- Right login form -->
  <div class="w-full lg:w-[460px] flex items-center justify-center p-4 sm:p-8 lg:p-12 relative">
    <!-- Card -->
    <div class="w-full max-w-sm rounded-2xl p-7 sm:p-8 page-enter"
         style="background: rgba(255,255,255,0.97); box-shadow: 0 25px 50px rgba(0,0,0,0.35); border: 1px solid rgba(255,255,255,0.5);">

      <!-- Mobile logo -->
      <div class="lg:hidden flex items-center gap-3 mb-7">
        <div class="w-10 h-10 rounded-xl flex items-center justify-center"
             style="background: linear-gradient(135deg, #3b82f6, #6366f1);">
          <span class="text-white font-black text-lg">G</span>
        </div>
        <div>
          <p class="text-slate-800 font-bold text-lg leading-tight">GBS CMS</p>
          <p class="text-slate-500 text-xs">Point of Sale</p>
        </div>
      </div>

      <!-- Heading -->
      <div class="mb-7">
        <h2 class="text-2xl font-bold text-slate-900 mb-1">Welcome back</h2>
        <p class="text-slate-500 text-sm">Sign in to access your dashboard</p>
      </div>

      <!-- Error Message -->
      {#if error}
        <div class="mb-5 p-3.5 rounded-xl flex items-start gap-3 text-red-700"
             style="background: #fef2f2; border: 1px solid #fecaca;">
          <AlertCircle size={18} class="flex-shrink-0 mt-0.5" />
          <span class="text-sm leading-snug">{error}</span>
        </div>
      {/if}

      <!-- Login Form -->
      <form onsubmit={handleLogin} class="space-y-4">
        <!-- Username -->
        <div>
          <label for="username" class="label">Username</label>
          <div class="relative">
            <User class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" size={17} />
            <input
              id="username"
              type="text"
              bind:value={username}
              class="input pl-10"
              placeholder="Enter your username"
              required
              disabled={loading}
              autocomplete="username"
            />
          </div>
        </div>

        <!-- Password -->
        <div>
          <label for="password" class="label">Password</label>
          <div class="relative">
            <Lock class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" size={17} />
            <input
              id="password"
              type={showPassword ? 'text' : 'password'}
              bind:value={password}
              class="input pl-10 pr-10"
              placeholder="Enter your password"
              required
              disabled={loading}
              autocomplete="current-password"
            />
            <button
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 transition-colors"
              onclick={() => showPassword = !showPassword}
              aria-label="Toggle password visibility"
            >
              {#if showPassword}
                <EyeOff size={17} />
              {:else}
                <Eye size={17} />
              {/if}
            </button>
          </div>
        </div>

        <!-- Submit -->
        <button
          type="submit"
          disabled={loading}
          class="btn btn-primary w-full py-3 text-sm mt-2"
        >
          {#if loading}
            <svg class="animate-spin h-4 w-4 text-white" viewBox="0 0 24 24" fill="none">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
            </svg>
            Signing in...
          {:else}
            Sign In
            <ArrowRight size={16} />
          {/if}
        </button>
      </form>

      <!-- Demo Credentials -->
      <div class="mt-7 pt-5 border-t border-slate-100">
        <p class="text-center text-xs text-slate-400 mb-3 font-medium uppercase tracking-wider">Quick Login</p>
        <div class="grid grid-cols-2 gap-3">
          <button
            type="button"
            onclick={() => fillCredentials('admin')}
            class="p-3 rounded-xl text-left transition-all hover:scale-[1.02] active:scale-95 cursor-pointer"
            style="background: linear-gradient(135deg, #eff6ff, #dbeafe); border: 1px solid #bfdbfe;"
          >
            <p class="font-semibold text-blue-800 text-sm">Admin</p>
            <p class="text-blue-500 text-xs mt-0.5">admin / admin123</p>
          </button>
          <button
            type="button"
            onclick={() => fillCredentials('cashier')}
            class="p-3 rounded-xl text-left transition-all hover:scale-[1.02] active:scale-95 cursor-pointer"
            style="background: linear-gradient(135deg, #f0fdf4, #dcfce7); border: 1px solid #bbf7d0;"
          >
            <p class="font-semibold text-emerald-800 text-sm">Cashier</p>
            <p class="text-emerald-500 text-xs mt-0.5">cashier / cashier123</p>
          </button>
        </div>
      </div>
    </div>
  </div>
</div>
