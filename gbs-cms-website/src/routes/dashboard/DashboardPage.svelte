<script lang="ts">
  import { onMount } from 'svelte';
  import {
    TrendingUp,
    ShoppingCart,
    DollarSign,
    XCircle,
    CreditCard,
    Smartphone,
    Banknote,
    Package,
    RefreshCw
  } from 'lucide-svelte';
  import { posApi } from '../../lib/api/client';
  import type { DashboardSummary, TopProduct, RevenuePoint } from '../../types/api';

  let summary = $state<DashboardSummary | null>(null);
  let topProducts = $state<TopProduct[]>([]);
  let revenueTrend = $state<RevenuePoint[]>([]);
  let loading = $state(true);
  let error = $state('');
  let range = $state<'today' | 'week' | 'month'>('today');

  async function loadDashboard() {
    loading = true;
    error = '';

    try {
      const [summaryRes, productsRes, revenueRes] = await Promise.all([
        posApi.get<DashboardSummary>(`/dashboard/summary`),
        posApi.get<TopProduct[]>('/dashboard/top-products?limit=5'),
        posApi.get<RevenuePoint[]>('/dashboard/revenue'),
      ]);

      if (summaryRes.success && summaryRes.data) {
        summary = summaryRes.data;
      }

      if (productsRes.success && productsRes.data) {
        topProducts = Array.isArray(productsRes.data) ? productsRes.data : [];
      }

      if (revenueRes.success && revenueRes.data) {
        revenueTrend = Array.isArray(revenueRes.data) ? revenueRes.data : [];
      }
    } catch (e) {
      error = 'Failed to load dashboard data';
      console.error(e);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadDashboard();
  });

  function formatCurrency(amount: number): string {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);
  }

  function formatCurrencyShort(amount: number): string {
    if (amount >= 1_000_000_000) return `Rp${(amount / 1_000_000_000).toFixed(1)}B`;
    if (amount >= 1_000_000) return `Rp${(amount / 1_000_000).toFixed(1)}M`;
    if (amount >= 1_000) return `Rp${(amount / 1_000).toFixed(0)}K`;
    return `Rp${amount}`;
  }

  function handleRangeChange(newRange: 'today' | 'week' | 'month') {
    range = newRange;
    loadDashboard();
  }

  // Max revenue for mini bar chart normalization
  let maxRevenue = $derived(
    revenueTrend.length > 0 ? Math.max(...revenueTrend.map(p => p.revenue)) : 1
  );

  // Payment totals for progress bars
  let totalPayments = $derived(
    summary ? (summary.cashTotal + summary.cardTotal + summary.qrisTotal) || 1 : 1
  );
</script>

<div class="space-y-5 md:space-y-6 page-enter">
  <!-- Header + Range Selector -->
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
    <div>
      <h1 class="text-xl sm:text-2xl font-bold text-slate-900">Dashboard</h1>
      <p class="text-sm text-slate-500 mt-0.5">Overview of your store performance</p>
    </div>
    <div class="flex items-center gap-2">
      <!-- Refresh -->
      <button
        onclick={loadDashboard}
        class="btn-icon text-slate-500 hover:text-slate-700 hover:bg-slate-100"
        disabled={loading}
        aria-label="Refresh"
      >
        <RefreshCw size={16} class="{loading ? 'animate-spin' : ''}" />
      </button>
      <!-- Range selector -->
      <div class="flex gap-1 p-1 rounded-xl" style="background: #f1f5f9; border: 1px solid #e2e8f0;">
        {#each [['today','Today'],['week','Week'],['month','Month']] as [val, label]}
          <button
            onclick={() => handleRangeChange(val as 'today'|'week'|'month')}
            class="px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all duration-150"
            class:bg-white={range === val}
            class:text-blue-700={range === val}
            class:shadow-sm={range === val}
            class:text-slate-500={range !== val}
            style={range === val ? 'border: 1px solid #e2e8f0;' : ''}
          >
            {label}
          </button>
        {/each}
      </div>
    </div>
  </div>

  {#if error}
    <div class="p-4 rounded-xl text-red-700 flex items-center gap-3"
         style="background: #fef2f2; border: 1px solid #fecaca;">
      <XCircle size={18} class="flex-shrink-0" />
      <span class="text-sm">{error}</span>
    </div>
  {/if}

  {#if loading}
    <!-- Skeleton Loading -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      {#each [1,2,3,4] as _}
        <div class="skeleton-card p-5 space-y-3">
          <div class="flex items-center justify-between">
            <div class="skeleton-text w-24 h-3"></div>
            <div class="skeleton w-12 h-12 rounded-xl"></div>
          </div>
          <div class="skeleton-text w-16 h-7"></div>
        </div>
      {/each}
    </div>
    <div class="skeleton-card h-48"></div>
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-5">
      <div class="skeleton-card h-64"></div>
      <div class="skeleton-card h-64"></div>
    </div>
  {:else if summary}
    <!-- =====================
         Summary Stat Cards
         ===================== -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 md:gap-4">
      <!-- Total Orders -->
      <div class="card p-4 md:p-5 card-hover">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">Orders</p>
            <p class="text-2xl md:text-3xl font-bold text-slate-900">{summary.totalOrders}</p>
          </div>
          <div class="stat-icon flex-shrink-0" style="background: linear-gradient(135deg, #eff6ff, #dbeafe);">
            <ShoppingCart class="text-blue-600" size={22} />
          </div>
        </div>
        <p class="text-xs text-slate-400 mt-2">Total transactions</p>
      </div>

      <!-- Total Revenue -->
      <div class="card p-4 md:p-5 card-hover">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">Revenue</p>
            <p class="text-lg md:text-xl font-bold text-emerald-600 leading-tight">
              {formatCurrencyShort(summary.totalRevenue)}
            </p>
          </div>
          <div class="stat-icon flex-shrink-0" style="background: linear-gradient(135deg, #f0fdf4, #dcfce7);">
            <DollarSign class="text-emerald-600" size={22} />
          </div>
        </div>
        <p class="text-xs text-slate-400 mt-2 truncate">{formatCurrency(summary.totalRevenue)}</p>
      </div>

      <!-- Avg Order Value -->
      <div class="card p-4 md:p-5 card-hover">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">Avg Order</p>
            <p class="text-lg md:text-xl font-bold text-purple-600 leading-tight">
              {formatCurrencyShort(summary.avgOrderValue)}
            </p>
          </div>
          <div class="stat-icon flex-shrink-0" style="background: linear-gradient(135deg, #faf5ff, #ede9fe);">
            <TrendingUp class="text-purple-600" size={22} />
          </div>
        </div>
        <p class="text-xs text-slate-400 mt-2 truncate">{formatCurrency(summary.avgOrderValue)}</p>
      </div>

      <!-- Voided Orders -->
      <div class="card p-4 md:p-5 card-hover">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">Voided</p>
            <p class="text-2xl md:text-3xl font-bold text-red-500">{summary.voidedCount}</p>
          </div>
          <div class="stat-icon flex-shrink-0" style="background: linear-gradient(135deg, #fff1f2, #ffe4e6);">
            <XCircle class="text-red-500" size={22} />
          </div>
        </div>
        <p class="text-xs text-slate-400 mt-2">Cancelled orders</p>
      </div>
    </div>

    <!-- =====================
         Payment Breakdown
         ===================== -->
    <div class="card p-5">
      <div class="flex items-center justify-between mb-5">
        <h2 class="text-base font-bold text-slate-900">Payment Breakdown</h2>
        <span class="text-xs text-slate-400 font-medium">
          Total: {formatCurrencyShort(summary.cashTotal + summary.cardTotal + summary.qrisTotal)}
        </span>
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <!-- Cash -->
        <div class="flex flex-col gap-3 p-4 rounded-xl" style="background: #f0fdf4; border: 1px solid #bbf7d0;">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <div class="w-9 h-9 rounded-xl flex items-center justify-center"
                   style="background: linear-gradient(135deg, #059669, #10b981);">
                <Banknote class="text-white" size={18} />
              </div>
              <span class="font-semibold text-emerald-900 text-sm">Cash</span>
            </div>
            <span class="text-xs font-bold text-emerald-700">
              {Math.round((summary.cashTotal / totalPayments) * 100)}%
            </span>
          </div>
          <div>
            <div class="h-1.5 rounded-full overflow-hidden mb-1.5" style="background: #bbf7d0;">
              <div class="h-full rounded-full" style="width: {Math.round((summary.cashTotal / totalPayments) * 100)}%; background: linear-gradient(90deg, #059669, #10b981); transition: width 0.8s ease;"></div>
            </div>
            <p class="font-bold text-emerald-800 text-sm">{formatCurrency(summary.cashTotal)}</p>
          </div>
        </div>

        <!-- Card -->
        <div class="flex flex-col gap-3 p-4 rounded-xl" style="background: #eff6ff; border: 1px solid #bfdbfe;">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <div class="w-9 h-9 rounded-xl flex items-center justify-center"
                   style="background: linear-gradient(135deg, #2563eb, #3b82f6);">
                <CreditCard class="text-white" size={18} />
              </div>
              <span class="font-semibold text-blue-900 text-sm">Card</span>
            </div>
            <span class="text-xs font-bold text-blue-700">
              {Math.round((summary.cardTotal / totalPayments) * 100)}%
            </span>
          </div>
          <div>
            <div class="h-1.5 rounded-full overflow-hidden mb-1.5" style="background: #bfdbfe;">
              <div class="h-full rounded-full" style="width: {Math.round((summary.cardTotal / totalPayments) * 100)}%; background: linear-gradient(90deg, #2563eb, #3b82f6); transition: width 0.8s ease;"></div>
            </div>
            <p class="font-bold text-blue-800 text-sm">{formatCurrency(summary.cardTotal)}</p>
          </div>
        </div>

        <!-- QRIS -->
        <div class="flex flex-col gap-3 p-4 rounded-xl" style="background: #faf5ff; border: 1px solid #ddd6fe;">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <div class="w-9 h-9 rounded-xl flex items-center justify-center"
                   style="background: linear-gradient(135deg, #7c3aed, #a855f7);">
                <Smartphone class="text-white" size={18} />
              </div>
              <span class="font-semibold text-purple-900 text-sm">QRIS</span>
            </div>
            <span class="text-xs font-bold text-purple-700">
              {Math.round((summary.qrisTotal / totalPayments) * 100)}%
            </span>
          </div>
          <div>
            <div class="h-1.5 rounded-full overflow-hidden mb-1.5" style="background: #ddd6fe;">
              <div class="h-full rounded-full" style="width: {Math.round((summary.qrisTotal / totalPayments) * 100)}%; background: linear-gradient(90deg, #7c3aed, #a855f7); transition: width 0.8s ease;"></div>
            </div>
            <p class="font-bold text-purple-800 text-sm">{formatCurrency(summary.qrisTotal)}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- =====================
         Top Products + Revenue Trend
         ===================== -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-5">
      <!-- Top Products -->
      <div class="card overflow-hidden">
        <div class="p-4 md:p-5 border-b border-slate-100 flex items-center gap-2">
          <div class="w-7 h-7 rounded-lg flex items-center justify-center"
               style="background: linear-gradient(135deg, #eff6ff, #dbeafe);">
            <Package class="text-blue-600" size={14} />
          </div>
          <h2 class="text-base font-bold text-slate-900">Top Products</h2>
        </div>
        <div class="p-4 md:p-5">
          {#if topProducts.length > 0}
            <div class="space-y-1">
              {#each topProducts as product, i}
                <div class="flex items-center gap-3 py-2.5 px-2 rounded-lg hover:bg-slate-50 transition-colors">
                  <span class="text-xs font-bold text-slate-400 w-5 text-center flex-shrink-0">
                    {i + 1}
                  </span>
                  <div class="flex-1 min-w-0">
                    <p class="font-medium text-slate-800 text-sm truncate">{product.productName}</p>
                    <p class="text-xs text-slate-400 mt-0.5">{product.totalSold} sold</p>
                  </div>
                  <span class="font-bold text-emerald-600 text-sm flex-shrink-0">
                    {formatCurrencyShort(product.revenue)}
                  </span>
                </div>
              {/each}
            </div>
          {:else}
            <div class="text-center py-10">
              <Package size={36} class="mx-auto mb-3 text-slate-300" />
              <p class="text-slate-400 text-sm">No sales data available</p>
            </div>
          {/if}
        </div>
      </div>

      <!-- Revenue Trend -->
      <div class="card overflow-hidden">
        <div class="p-4 md:p-5 border-b border-slate-100 flex items-center gap-2">
          <div class="w-7 h-7 rounded-lg flex items-center justify-center"
               style="background: linear-gradient(135deg, #f0fdf4, #dcfce7);">
            <TrendingUp class="text-emerald-600" size={14} />
          </div>
          <h2 class="text-base font-bold text-slate-900">Revenue Trend</h2>
        </div>
        <div class="p-4 md:p-5">
          {#if revenueTrend.length > 0}
            <div class="space-y-3">
              {#each revenueTrend as point}
                <div class="flex flex-col gap-1.5">
                  <div class="flex items-center justify-between text-xs">
                    <span class="font-medium text-slate-600">{point.date}</span>
                    <div class="flex items-center gap-3 flex-shrink-0">
                      <span class="text-slate-400">{point.orders} orders</span>
                      <span class="font-bold text-emerald-600">{formatCurrencyShort(point.revenue)}</span>
                    </div>
                  </div>
                  <!-- Mini bar -->
                  <div class="h-1.5 rounded-full overflow-hidden" style="background: #f1f5f9;">
                    <div class="h-full rounded-full revenue-bar"
                         style="width: {maxRevenue > 0 ? Math.round((point.revenue / maxRevenue) * 100) : 0}%;"></div>
                  </div>
                </div>
              {/each}
            </div>
          {:else}
            <div class="text-center py-10">
              <TrendingUp size={36} class="mx-auto mb-3 text-slate-300" />
              <p class="text-slate-400 text-sm">No revenue data available</p>
            </div>
          {/if}
        </div>
      </div>
    </div>
  {:else}
    <div class="text-center py-20 text-slate-400">
      <p>No data available</p>
    </div>
  {/if}
</div>
