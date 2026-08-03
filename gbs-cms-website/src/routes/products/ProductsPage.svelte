<script lang="ts">
  import { onMount } from 'svelte';
  import { Package, Search, Plus, Edit2, Trash2, AlertTriangle, X, ChevronLeft, ChevronRight, Tag } from 'lucide-svelte';
  import { posApi } from '../../lib/api/client';
  import { authStore } from '../../lib/stores/auth';
  import type { Product } from '../../types/api';

  let products = $state<Product[]>([]);
  let loading = $state(true);
  let error = $state('');
  let searchQuery = $state('');
  let showModal = $state(false);
  let editingProduct = $state<Product | null>(null);

  // Pagination
  let currentPage = $state(1);
  let pageSize = $state(10);
  let totalProducts = $state(0);

  // Form state
  let formData = $state({
    name: '',
    price: 0,
    category: '',
    storeType: 'ALL',
    barcode: '',
    stockQuantity: 0,
    lowStockThreshold: 10,
  });

  // Check if admin
  let isAdmin = $state(false);
  onMount(() => {
    loadProducts();
    const unsub = authStore.isAdmin.subscribe(v => isAdmin = v);
    return unsub;
  });

  async function loadProducts() {
    loading = true;
    error = '';

    try {
      const response = await posApi.get<Product[]>('/products');

      if (response.success && response.data) {
        products = Array.isArray(response.data) ? response.data : [];
        totalProducts = products.length;
      } else {
        error = response.error?.message || 'Failed to load products';
      }
    } catch (e) {
      error = 'Connection error';
    } finally {
      loading = false;
    }
  }

  let filteredProducts = $derived(
    products.filter(p =>
      p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      p.barcode.includes(searchQuery) ||
      p.category.toLowerCase().includes(searchQuery.toLowerCase())
    )
  );

  $effect(() => {
    // Reset to page 1 when search changes
    currentPage = 1;
    searchQuery; // track
  });

  let totalPages = $derived(Math.ceil(filteredProducts.length / pageSize));
  let paginatedProducts = $derived(
    filteredProducts.slice((currentPage - 1) * pageSize, currentPage * pageSize)
  );

  function formatCurrency(amount: number): string {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);
  }

  function openModal(product?: Product) {
    if (product) {
      editingProduct = product;
      formData = {
        name: product.name,
        price: product.price,
        category: product.category,
        storeType: product.storeType,
        barcode: product.barcode,
        stockQuantity: product.stockQuantity,
        lowStockThreshold: product.lowStockThreshold,
      };
    } else {
      editingProduct = null;
      formData = {
        name: '',
        price: 0,
        category: '',
        storeType: 'ALL',
        barcode: '',
        stockQuantity: 0,
        lowStockThreshold: 10,
      };
    }
    showModal = true;
  }

  function closeModal() {
    showModal = false;
    editingProduct = null;
  }

  async function handleSubmit() {
    try {
      const payload = {
        ...formData,
        price: Number(formData.price),
        stockQuantity: Number(formData.stockQuantity),
        lowStockThreshold: Number(formData.lowStockThreshold),
      };

      let response;
      if (editingProduct) {
        response = await posApi.put(`/products/${editingProduct.id}`, payload);
      } else {
        response = await posApi.post('/products', payload);
      }

      if (response.success) {
        closeModal();
        loadProducts();
      } else {
        error = response.error?.message || 'Operation failed';
      }
    } catch (e) {
      error = 'Operation failed';
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Are you sure you want to delete this product?')) return;

    try {
      const response = await posApi.delete(`/products/${id}`);
      if (response.success) {
        loadProducts();
      } else {
        error = response.error?.message || 'Delete failed';
      }
    } catch (e) {
      error = 'Delete failed';
    }
  }

  function isLowStock(product: Product): boolean {
    return product.stockQuantity <= product.lowStockThreshold;
  }

  function getPageNumbers(): number[] {
    const pages: number[] = [];
    const delta = 2;
    for (let i = Math.max(1, currentPage - delta); i <= Math.min(totalPages, currentPage + delta); i++) {
      pages.push(i);
    }
    return pages;
  }
</script>

<div class="space-y-4 md:space-y-5 page-enter">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
    <div>
      <h1 class="text-xl sm:text-2xl font-bold text-slate-900">Products</h1>
      <p class="text-sm text-slate-500 mt-0.5">{totalProducts} total products</p>
    </div>
    {#if isAdmin}
      <button onclick={() => openModal()} class="btn btn-primary w-full sm:w-auto">
        <Plus size={17} />
        Add Product
      </button>
    {/if}
  </div>

  <!-- Search -->
  <div class="relative">
    <Search class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" size={18} />
    <input
      type="text"
      bind:value={searchQuery}
      placeholder="Search by name, barcode, or category..."
      class="input pl-10"
    />
    {#if searchQuery}
      <button
        class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 transition-colors"
        onclick={() => searchQuery = ''}
        aria-label="Clear search"
      >
        <X size={16} />
      </button>
    {/if}
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
          <div class="skeleton w-10 h-10 rounded-lg"></div>
          <div class="flex-1 space-y-2">
            <div class="skeleton-text w-40"></div>
            <div class="skeleton-text w-24 h-3"></div>
          </div>
          <div class="skeleton-text w-20 h-5"></div>
          <div class="skeleton-text w-16 h-5 hidden md:block"></div>
        </div>
      {/each}
    </div>
  {:else if paginatedProducts.length === 0}
    <div class="card p-12 text-center">
      <div class="w-16 h-16 rounded-2xl flex items-center justify-center mx-auto mb-4"
           style="background: #f8fafc; border: 2px dashed #cbd5e1;">
        <Package size={28} class="text-slate-300" />
      </div>
      <p class="text-slate-500 font-medium">No products found</p>
      <p class="text-slate-400 text-sm mt-1">
        {searchQuery ? 'Try a different search term' : 'Add your first product to get started'}
      </p>
    </div>
  {:else}
    <!-- Desktop Table -->
    <div class="hidden md:block card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="table">
          <thead>
            <tr>
              <th>Product</th>
              <th>Category</th>
              <th>Price</th>
              <th>Stock</th>
              <th>Barcode</th>
              {#if isAdmin}
                <th class="text-right">Actions</th>
              {/if}
            </tr>
          </thead>
          <tbody>
            {#each paginatedProducts as product}
              <tr>
                <td>
                  <div class="flex items-center gap-3">
                    {#if product.imageUrl}
                      <img src={product.imageUrl} alt={product.name} class="w-10 h-10 rounded-lg object-cover flex-shrink-0" />
                    {:else}
                      <div class="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0"
                           style="background: #f1f5f9;">
                        <Package class="text-slate-400" size={18} />
                      </div>
                    {/if}
                    <div>
                      <p class="font-semibold text-slate-800">{product.name}</p>
                      {#if product.discount}
                        <span class="badge badge-red mt-0.5">
                          <Tag size={10} />
                          -{product.discount.value}{product.discount.type === 'PERCENT' ? '%' : ''}
                        </span>
                      {/if}
                    </div>
                  </div>
                </td>
                <td>
                  <span class="badge badge-slate">{product.category}</span>
                </td>
                <td>
                  <div>
                    <span class="font-bold text-emerald-600">{formatCurrency(product.finalPrice)}</span>
                    {#if product.discount}
                      <p class="text-xs text-slate-400 line-through">{formatCurrency(product.price)}</p>
                    {/if}
                  </div>
                </td>
                <td>
                  {#if isLowStock(product)}
                    <span class="badge badge-amber">
                      <AlertTriangle size={11} />
                      {product.stockQuantity}
                    </span>
                  {:else}
                    <span class="badge badge-green">{product.stockQuantity}</span>
                  {/if}
                </td>
                <td>
                  <span class="font-mono text-xs text-slate-500">{product.barcode || '—'}</span>
                </td>
                {#if isAdmin}
                  <td>
                    <div class="flex items-center justify-end gap-1">
                      <button
                        onclick={() => openModal(product)}
                        class="btn-icon text-slate-500 hover:text-blue-600 hover:bg-blue-50"
                        aria-label="Edit"
                      >
                        <Edit2 size={15} />
                      </button>
                      <button
                        onclick={() => handleDelete(product.id)}
                        class="btn-icon text-slate-400 hover:text-red-600 hover:bg-red-50"
                        aria-label="Delete"
                      >
                        <Trash2 size={15} />
                      </button>
                    </div>
                  </td>
                {/if}
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>

    <!-- Mobile Card View -->
    <div class="md:hidden space-y-3">
      {#each paginatedProducts as product}
        <div class="card p-4">
          <div class="flex items-start gap-3">
            {#if product.imageUrl}
              <img src={product.imageUrl} alt={product.name} class="w-12 h-12 rounded-xl object-cover flex-shrink-0" />
            {:else}
              <div class="w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0"
                   style="background: #f1f5f9;">
                <Package class="text-slate-400" size={22} />
              </div>
            {/if}
            <div class="flex-1 min-w-0">
              <div class="flex items-start justify-between gap-2">
                <p class="font-semibold text-slate-800 leading-tight">{product.name}</p>
                {#if isAdmin}
                  <div class="flex gap-1 flex-shrink-0">
                    <button onclick={() => openModal(product)} class="btn-icon text-slate-400 hover:text-blue-600 hover:bg-blue-50">
                      <Edit2 size={14} />
                    </button>
                    <button onclick={() => handleDelete(product.id)} class="btn-icon text-slate-400 hover:text-red-600 hover:bg-red-50">
                      <Trash2 size={14} />
                    </button>
                  </div>
                {/if}
              </div>
              <div class="flex items-center gap-2 mt-1 flex-wrap">
                <span class="badge badge-slate text-xs">{product.category}</span>
                {#if product.discount}
                  <span class="badge badge-red text-xs">
                    <Tag size={9} />
                    -{product.discount.value}{product.discount.type === 'PERCENT' ? '%' : ''}
                  </span>
                {/if}
              </div>
            </div>
          </div>

          <div class="mt-3 pt-3 border-t border-slate-100 flex items-center justify-between gap-4">
            <div>
              <span class="font-bold text-emerald-600">{formatCurrency(product.finalPrice)}</span>
              {#if product.discount}
                <span class="text-xs text-slate-400 line-through ml-1">{formatCurrency(product.price)}</span>
              {/if}
            </div>
            <div class="flex items-center gap-3">
              {#if product.barcode}
                <span class="font-mono text-xs text-slate-400">{product.barcode}</span>
              {/if}
              {#if isLowStock(product)}
                <span class="badge badge-amber">
                  <AlertTriangle size={11} />
                  {product.stockQuantity}
                </span>
              {:else}
                <span class="badge badge-green">Stock: {product.stockQuantity}</span>
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
          Showing {(currentPage - 1) * pageSize + 1}–{Math.min(currentPage * pageSize, filteredProducts.length)} of {filteredProducts.length}
        </p>
        <p class="text-sm text-slate-500 sm:hidden">
          {currentPage}/{totalPages}
        </p>
        <div class="flex items-center gap-1">
          <button
            onclick={() => currentPage = Math.max(1, currentPage - 1)}
            disabled={currentPage === 1}
            class="btn-icon text-slate-500 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-slate-100"
          >
            <ChevronLeft size={18} />
          </button>
          {#each getPageNumbers() as pageNum}
            <button
              onclick={() => currentPage = pageNum}
              class="min-w-[32px] h-8 px-2 rounded-lg text-sm font-medium transition-all
                     {currentPage === pageNum
                       ? 'bg-blue-600 text-white shadow-sm'
                       : 'text-slate-600 hover:bg-slate-100'}"
            >
              {pageNum}
            </button>
          {/each}
          <button
            onclick={() => currentPage = Math.min(totalPages, currentPage + 1)}
            disabled={currentPage === totalPages}
            class="btn-icon text-slate-500 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-slate-100"
          >
            <ChevronRight size={18} />
          </button>
        </div>
      </div>
    {/if}
  {/if}
</div>

<!-- Modal -->
{#if showModal}
  <div class="fixed inset-0 z-50 flex items-end sm:items-center justify-center p-0 sm:p-4">
    <button class="absolute inset-0 bg-black/60" style="backdrop-filter: blur(3px);" onclick={closeModal} aria-label="Close modal"></button>
    <div class="bg-white w-full sm:max-w-md sm:rounded-2xl rounded-t-2xl shadow-2xl max-h-[95vh] overflow-y-auto relative modal-content">
      <div class="flex items-center justify-between p-5 border-b border-slate-100">
        <h3 class="text-lg font-bold text-slate-900">{editingProduct ? 'Edit Product' : 'Add Product'}</h3>
        <button onclick={closeModal} class="btn-icon text-slate-400 hover:text-slate-600 hover:bg-slate-100">
          <X size={20} />
        </button>
      </div>

      <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="p-5 space-y-4">
        <div>
          <label class="label" for="name">Product Name</label>
          <input id="name" type="text" bind:value={formData.name} class="input" required placeholder="Enter product name" />
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="label" for="price">Price (IDR)</label>
            <input id="price" type="number" bind:value={formData.price} class="input" required min="0" />
          </div>
          <div>
            <label class="label" for="category">Category</label>
            <input id="category" type="text" bind:value={formData.category} class="input" required placeholder="e.g. Beverage" />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="label" for="stock">Stock Qty</label>
            <input id="stock" type="number" bind:value={formData.stockQuantity} class="input" min="0" />
          </div>
          <div>
            <label class="label" for="threshold">Low Stock Alert</label>
            <input id="threshold" type="number" bind:value={formData.lowStockThreshold} class="input" min="0" />
          </div>
        </div>

        <div>
          <label class="label" for="barcode">Barcode</label>
          <input id="barcode" type="text" bind:value={formData.barcode} class="input" placeholder="Optional" />
        </div>

        <div>
          <label class="label" for="storeType">Store Type</label>
          <select id="storeType" bind:value={formData.storeType} class="input">
            <option value="ALL">All Stores</option>
            <option value="RETAIL">Retail</option>
            <option value="FB">Food & Beverage</option>
            <option value="FUEL">Fuel Station</option>
          </select>
        </div>

        <div class="flex gap-3 pt-2">
          <button type="button" onclick={closeModal} class="btn btn-secondary flex-1">Cancel</button>
          <button type="submit" class="btn btn-primary flex-1">
            {editingProduct ? 'Update Product' : 'Add Product'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
