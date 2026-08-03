<script lang="ts">
  import { onMount } from 'svelte';
  import { ShoppingCart, Search, X, ChevronLeft, ChevronRight, Banknote, CreditCard, Smartphone, Eye, AlertTriangle, Filter } from 'lucide-svelte';
  import { posApi } from '../../lib/api/client';
  import { authStore } from '../../lib/stores/auth';
  import type { Order } from '../../types/api';

  let orders = $state<Order[]>([]);
  let loading = $state(true);
  let error = $state('');
  let searchQuery = $state('');
  let selectedOrder = $state<Order | null>(null);
  let showVoidModal = $state(false);
  let voidReason = $state('');
  let voidingOrderId = $state<string | null>(null);
  let statusFilter = $state<'ALL' | 'COMPLETED' | 'VOIDED'>('ALL');

  // Pagination
  let currentPage = $state(1);
  let pageSize = $state(10);

  // Check if admin
  let isAdmin = $state(false);
  onMount(() => {
    loadOrders();
    const unsub = authStore.isAdmin.subscribe(v => isAdmin = v);
    return unsub;
  });

  async function loadOrders() {
    loading = true;
    error = '';

    try {
      const response = await posApi.get<Order[]>('/orders');

      if (response.success && response.data) {
        orders = Array.isArray(response.data) ? response.data : [];
      } else {
        error = response.error?.message || 'Failed to load orders';
      }
    } catch (e) {
      error = 'Connection error';
    } finally {
      loading = false;
    }
  }

  let filteredOrders = $derived(
    orders.filter(o => {
      const matchSearch = o.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
        o.createdAt.includes(searchQuery);
      const matchStatus = statusFilter === 'ALL' || o.status === statusFilter;
      return matchSearch && matchStatus;
    })
  );

  $effect(() => {
    currentPage = 1;
    searchQuery; statusFilter; // track changes
  });

  let totalPages = $derived(Math.ceil(filteredOrders.length / pageSize));
  let paginatedOrders = $derived(
    filteredOrders.slice((currentPage - 1) * pageSize, currentPage * pageSize)
  );

  function formatCurrency(amount: number): string {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleString('id-ID', {
      day: '2-digit',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  function openOrderDetail(order: Order) {
    selectedOrder = order;
  }

  function closeOrderDetail() {
    selectedOrder = null;
  }

  function openVoidModal(order: Order) {
    selectedOrder = order;
    voidReason = '';
    showVoidModal = true;
  }

  function closeVoidModal() {
    showVoidModal = false;
    voidReason = '';
  }

  async function handleVoid() {
    if (!selectedOrder || !voidReason.trim()) return;

    voidingOrderId = selectedOrder.id;

    try {
      const response = await posApi.patch<Order>(`/orders/${selectedOrder.id}/void`, {
        reason: voidReason,
      });

      if (response.success) {
        closeVoidModal();
        closeOrderDetail();
        loadOrders();
      } else {
        error = response.error?.message || 'Void failed';
      }
    } catch (e) {
      error = 'Void failed';
    } finally {
      voidingOrderId = null;
    }
  }

  function getPaymentIcon(method: string) {
    if (method === 'CASH') return Banknote;
    if (method === 'CARD') return CreditCard;
    return Smartphone;
  }

  function getPaymentStyle(method: string): string {
    if (method === 'CASH') return 'bg-emerald-100 text-emerald-700';
    if (method === 'CARD') return 'bg-blue-100 text-blue-700';
    return 'bg-purple-100 text-purple-700';
  }

  function getStatusStyle(status: string): string {
    return status === 'COMPLETED'
      ? 'bg-emerald-100 text-emerald-700'
      : 'bg-red-100 text-red-700';
  }

  function getPageNumbers(): number[] {
    const pages: number[] = [];
    const delta = 2;
    for (let i = Math.max(1, currentPage - delta); i <= Math.min(totalPages, currentPage + delta); i++) {
      pages.push(i);
    }
    return pages;
  }

  let completedCount = $derived(orders.filter(o => o.status === 'COMPLETED').length);
  let voidedCount = $derived(orders.filter(o => o.status === 'VOIDED').length);
</script>

<div class="space-y-4 md:space-y-5 page-enter">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-xl sm:text-2xl font-bold text-slate-900">Orders</h1>
      <p class="text-sm text-slate-500 mt-0.5">{orders.length} total orders</p>
    </div>
  </div>

  <!-- Stats row -->
  {#if !loading && orders.length > 0}
    <div class="grid grid-cols-3 gap-3">
      <div class="card p-3 text-center">
        <p class="text-xl font-bold text-slate-800">{orders.length}</p>
        <p class="text-xs text-slate-500 mt-0.5">Total</p>
      </div>
      <div class="card p-3 text-center">
        <p class="text-xl font-bold text-emerald-600">{completedCount}</p>
        <p class="text-xs text-slate-500 mt-0.5">Completed</p>
      </div>
      <div class="card p-3 text-center">
        <p class="text-xl font-bold text-red-500">{voidedCount}</p>
        <p class="text-xs text-slate-500 mt-0.5">Voided</p>
      </div>
    </div>
  {/if}

  <!-- Search + Filter -->
  <div class="flex gap-2">
    <div class="relative flex-1">
      <Search class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" size={18} />
      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Search by order ID or date..."
        class="input pl-10"
      />
      {#if searchQuery}
        <button
          class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
          onclick={() => searchQuery = ''}
          aria-label="Clear"
        >
          <X size={15} />
        </button>
      {/if}
    </div>
    <!-- Status Filter -->
    <div class="relative">
      <select
        bind:value={statusFilter}
        class="input appearance-none pr-8 min-w-0 cursor-pointer"
        style="min-width: 130px;"
        aria-label="Filter by status"
      >
        <option value="ALL">All Status</option>
        <option value="COMPLETED">Completed</option>
        <option value="VOIDED">Voided</option>
      </select>
      <Filter size={14} class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" />
    </div>
  </div>

  {#if error}
    <div class="p-4 rounded-xl text-red-700 flex items-center gap-3 text-sm"
         style="background: #fef2f2; border: 1px solid #fecaca;">
      {error}
    </div>
  {/if}

  {#if loading}
    <!-- Skeleton -->
    <div class="card overflow-hidden">
      {#each [1,2,3,4,5] as _}
        <div class="flex items-center gap-4 px-5 py-4 border-b border-slate-100">
          <div class="skeleton-text w-24 h-4"></div>
          <div class="flex-1 space-y-1.5">
            <div class="skeleton-text w-32"></div>
            <div class="skeleton-text w-20 h-3"></div>
          </div>
          <div class="skeleton w-16 h-6 rounded-full"></div>
          <div class="skeleton w-12 h-7 rounded-lg hidden sm:block"></div>
        </div>
      {/each}
    </div>
  {:else if paginatedOrders.length === 0}
    <div class="card p-12 text-center">
      <div class="w-16 h-16 rounded-2xl flex items-center justify-center mx-auto mb-4"
           style="background: #f8fafc; border: 2px dashed #cbd5e1;">
        <ShoppingCart size={28} class="text-slate-300" />
      </div>
      <p class="text-slate-500 font-medium">No orders found</p>
      <p class="text-slate-400 text-sm mt-1">
        {searchQuery || statusFilter !== 'ALL' ? 'Try adjusting your filters' : 'Orders will appear here once transactions are made'}
      </p>
    </div>
  {:else}
    <!-- Desktop Table -->
    <div class="hidden md:block card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="table">
          <thead>
            <tr>
              <th>Order ID</th>
              <th>Date & Time</th>
              <th>Items</th>
              <th>Total</th>
              <th>Payment</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {#each paginatedOrders as order}
              <tr>
                <td>
                  <span class="font-mono text-xs font-medium text-slate-600 bg-slate-100 px-2 py-0.5 rounded">
                    {order.id.slice(0, 8)}...
                  </span>
                </td>
                <td class="text-slate-600">{formatDate(order.createdAt)}</td>
                <td>
                  <span class="text-sm text-slate-600">{order.items.length} items</span>
                </td>
                <td>
                  <span class="font-bold text-slate-800">{formatCurrency(order.total)}</span>
                </td>
                <td>
                  <span class="badge {getPaymentStyle(order.paymentMethod)}">
                    {order.paymentMethod}
                  </span>
                </td>
                <td>
                  <span class="badge {getStatusStyle(order.status)}">
                    {order.status}
                  </span>
                </td>
                <td>
                  <div class="flex items-center justify-end gap-2">
                    <button
                      onclick={() => openOrderDetail(order)}
                      class="btn-icon text-slate-500 hover:text-blue-600 hover:bg-blue-50"
                      aria-label="View order"
                    >
                      <Eye size={15} />
                    </button>
                    {#if order.status === 'COMPLETED' && isAdmin}
                      <button
                        onclick={() => openVoidModal(order)}
                        class="btn btn-danger py-1 px-2.5 text-xs"
                      >
                        Void
                      </button>
                    {/if}
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>

    <!-- Mobile Card View — FIX: payment method now visible -->
    <div class="md:hidden space-y-3">
      {#each paginatedOrders as order}
        <div class="card p-4">
          <!-- Row 1: ID + Status -->
          <div class="flex items-start justify-between mb-3 gap-2">
            <div>
              <span class="font-mono text-xs font-semibold text-slate-600 bg-slate-100 px-2 py-0.5 rounded">
                {order.id.slice(0, 8)}...
              </span>
              <p class="text-xs text-slate-400 mt-1">{formatDate(order.createdAt)}</p>
            </div>
            <span class="badge {getStatusStyle(order.status)} flex-shrink-0">
              {order.status}
            </span>
          </div>

          <!-- Row 2: Items + Total + Payment (FIX: payment visible on mobile) -->
          <div class="flex items-center justify-between gap-3">
            <div class="flex items-center gap-3">
              <div>
                <p class="text-xs text-slate-400">{order.items.length} items</p>
                <p class="font-bold text-slate-800">{formatCurrency(order.total)}</p>
              </div>
              <!-- Payment badge — was missing on mobile before -->
              <span class="badge {getPaymentStyle(order.paymentMethod)}">
                {order.paymentMethod}
              </span>
            </div>
            <!-- Actions -->
            <div class="flex items-center gap-2 flex-shrink-0">
              <button
                onclick={() => openOrderDetail(order)}
                class="btn btn-secondary py-1.5 px-3 text-xs"
              >
                <Eye size={13} />
                View
              </button>
              {#if order.status === 'COMPLETED' && isAdmin}
                <button
                  onclick={() => openVoidModal(order)}
                  class="btn btn-danger py-1.5 px-3 text-xs"
                >
                  Void
                </button>
              {/if}
            </div>
          </div>
        </div>
      {/each}
    </div>

    <!-- Pagination -->
    {#if totalPages > 1}
      <div class="flex items-center justify-between">
        <p class="text-sm text-slate-500 hidden sm:block">
          Showing {(currentPage - 1) * pageSize + 1}–{Math.min(currentPage * pageSize, filteredOrders.length)} of {filteredOrders.length}
        </p>
        <p class="text-sm text-slate-500 sm:hidden">{currentPage}/{totalPages}</p>
        <div class="flex items-center gap-1">
          <button
            onclick={() => currentPage = Math.max(1, currentPage - 1)}
            disabled={currentPage === 1}
            class="btn-icon text-slate-500 disabled:opacity-40 hover:bg-slate-100"
          >
            <ChevronLeft size={18} />
          </button>
          {#each getPageNumbers() as pageNum}
            <button
              onclick={() => currentPage = pageNum}
              class="min-w-[32px] h-8 px-2 rounded-lg text-sm font-medium transition-all
                     {currentPage === pageNum ? 'bg-blue-600 text-white shadow-sm' : 'text-slate-600 hover:bg-slate-100'}"
            >
              {pageNum}
            </button>
          {/each}
          <button
            onclick={() => currentPage = Math.min(totalPages, currentPage + 1)}
            disabled={currentPage === totalPages}
            class="btn-icon text-slate-500 disabled:opacity-40 hover:bg-slate-100"
          >
            <ChevronRight size={18} />
          </button>
        </div>
      </div>
    {/if}
  {/if}
</div>

<!-- Order Detail Modal -->
{#if selectedOrder && !showVoidModal}
  <div class="fixed inset-0 z-50 flex items-end sm:items-center justify-center p-0 sm:p-4">
    <button class="absolute inset-0 bg-black/60" style="backdrop-filter: blur(3px);" onclick={closeOrderDetail} aria-label="Close"></button>
    <div class="bg-white w-full sm:max-w-lg sm:rounded-2xl rounded-t-2xl shadow-2xl max-h-[92vh] overflow-y-auto relative modal-content">
      <div class="flex items-center justify-between p-5 border-b border-slate-100">
        <div>
          <h3 class="text-lg font-bold text-slate-900">Order Details</h3>
          <p class="text-xs text-slate-400 font-mono mt-0.5">{selectedOrder.id}</p>
        </div>
        <button onclick={closeOrderDetail} class="btn-icon text-slate-400 hover:text-slate-600 hover:bg-slate-100">
          <X size={20} />
        </button>
      </div>

      <div class="p-5 space-y-5">
        <!-- Meta info grid -->
        <div class="grid grid-cols-2 gap-3">
          <div class="p-3 rounded-xl" style="background: #f8fafc;">
            <p class="text-xs text-slate-400 mb-1">Status</p>
            <span class="badge {getStatusStyle(selectedOrder.status)}">
              {selectedOrder.status}
            </span>
          </div>
          <div class="p-3 rounded-xl" style="background: #f8fafc;">
            <p class="text-xs text-slate-400 mb-1">Payment</p>
            <span class="badge {getPaymentStyle(selectedOrder.paymentMethod)}">
              {selectedOrder.paymentMethod}
            </span>
          </div>
          <div class="col-span-2 p-3 rounded-xl" style="background: #f8fafc;">
            <p class="text-xs text-slate-400 mb-1">Date & Time</p>
            <p class="text-sm font-semibold text-slate-800">{formatDate(selectedOrder.createdAt)}</p>
          </div>
        </div>

        <!-- Items -->
        <div>
          <h4 class="text-sm font-bold text-slate-700 mb-3">Order Items</h4>
          <div class="space-y-2 max-h-48 overflow-y-auto pr-1">
            {#each selectedOrder.items as item}
              <div class="flex justify-between items-center py-2 px-3 rounded-lg" style="background: #f8fafc;">
                <div>
                  <p class="text-sm font-medium text-slate-800">{item.productName}</p>
                  <p class="text-xs text-slate-400">× {item.qty}</p>
                </div>
                <span class="font-bold text-sm text-slate-800">{formatCurrency(item.subtotal)}</span>
              </div>
            {/each}
          </div>
        </div>

        <!-- Totals -->
        <div class="rounded-xl overflow-hidden" style="border: 1px solid #e2e8f0;">
          <div class="flex justify-between text-sm p-3 border-b border-slate-100">
            <span class="text-slate-500">Subtotal</span>
            <span class="font-medium">{formatCurrency(selectedOrder.subtotal)}</span>
          </div>
          <div class="flex justify-between text-sm p-3 border-b border-slate-100">
            <span class="text-slate-500">Tax</span>
            <span class="font-medium">{formatCurrency(selectedOrder.tax)}</span>
          </div>
          <div class="flex justify-between p-3 font-bold" style="background: #f0fdf4;">
            <span class="text-slate-800">Total</span>
            <span class="text-emerald-600 text-lg">{formatCurrency(selectedOrder.total)}</span>
          </div>
          {#if selectedOrder.cashReceived}
            <div class="flex justify-between text-sm p-3 border-t border-slate-100">
              <span class="text-slate-500">Cash Received</span>
              <span>{formatCurrency(selectedOrder.cashReceived)}</span>
            </div>
            <div class="flex justify-between text-sm p-3 border-t border-slate-100">
              <span class="text-slate-500">Change</span>
              <span class="font-semibold">{formatCurrency(selectedOrder.changeAmount || 0)}</span>
            </div>
          {/if}
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Void Order Modal -->
{#if showVoidModal && selectedOrder}
  <div class="fixed inset-0 z-50 flex items-end sm:items-center justify-center p-0 sm:p-4">
    <button class="absolute inset-0 bg-black/60" style="backdrop-filter: blur(3px);" onclick={closeVoidModal} aria-label="Close"></button>
    <div class="bg-white w-full sm:max-w-sm sm:rounded-2xl rounded-t-2xl shadow-2xl relative modal-content">
      <div class="flex items-center justify-between p-5 border-b border-slate-100">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded-lg flex items-center justify-center" style="background: #fef2f2;">
            <AlertTriangle class="text-red-500" size={16} />
          </div>
          <h3 class="text-lg font-bold text-red-600">Void Order</h3>
        </div>
        <button onclick={closeVoidModal} class="btn-icon text-slate-400 hover:text-slate-600 hover:bg-slate-100">
          <X size={20} />
        </button>
      </div>

      <form onsubmit={(e) => { e.preventDefault(); handleVoid(); }} class="p-5 space-y-4">
        <p class="text-slate-600 text-sm">
          Are you sure you want to void order
          <code class="font-mono text-xs bg-slate-100 px-1.5 py-0.5 rounded">{selectedOrder.id.slice(0, 8)}...</code>?
          This action cannot be undone.
        </p>

        <div>
          <label class="label" for="voidReason">Reason for voiding <span class="text-red-500">*</span></label>
          <textarea
            id="voidReason"
            bind:value={voidReason}
            class="input min-h-[90px] resize-none text-sm"
            placeholder="Enter the reason for voiding this order..."
            required
          ></textarea>
        </div>

        <div class="flex gap-3">
          <button type="button" onclick={closeVoidModal} class="btn btn-secondary flex-1">Cancel</button>
          <button
            type="submit"
            disabled={voidingOrderId === selectedOrder.id}
            class="btn btn-danger flex-1"
          >
            {voidingOrderId === selectedOrder.id ? 'Voiding...' : 'Void Order'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
