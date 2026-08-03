<script lang="ts">
  import { onMount } from 'svelte';
  import { Image, Upload, Search, Edit2, Trash2, ToggleLeft, ToggleRight, X, Play, ChevronLeft, ChevronRight, Clock, Eye } from 'lucide-svelte';
  import { cmsApi } from '../../lib/api/client';

  // Backend Ad model
  interface BackendAd {
    id: number;
    name: string;
    filename: string;
    mimeType: string;
    fileSize: number;
    storagePath: string;
    isActive: boolean;
    playCount: number;
    playlistOrder: number;
    storeTypes: string[];
    durationSeconds: number | null;
    startDate: string | null;
    endDate: string | null;
    createdAt: string;
    updatedAt: string;
  }

  interface AdListResponse {
    ads: BackendAd[];
    pagination: {
      page: number;
      limit: number;
      total: number;
      totalPages: number;
    };
  }

  let ads = $state<BackendAd[]>([]);
  let loading = $state(true);
  let error = $state('');
  let searchQuery = $state('');
  let showModal = $state(false);
  let editingAd = $state<BackendAd | null>(null);
  let uploadFile = $state<File | null>(null);
  let uploadProgress = $state(0);
  let isDragOver = $state(false);

  // Pagination
  let currentPage = $state(1);
  let pageSize = $state(6);

  // Form state
  let formData = $state({
    name: '',
    storeTypes: ['RETAIL'] as string[],
    playlistOrder: 0,
    durationSeconds: 30,
    startDate: '',
    endDate: '',
  });

  async function loadAds() {
    loading = true;
    error = '';

    try {
      const response = await cmsApi.get<AdListResponse>('/ads');

      if (response.success && response.data) {
        ads = response.data.ads || [];
      } else {
        error = response.error?.message || 'Failed to load ads';
      }
    } catch (e) {
      error = 'Connection error';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadAds();
  });

  let filteredAds = $derived(
    ads.filter(a => a.name.toLowerCase().includes(searchQuery.toLowerCase()))
  );

  $effect(() => {
    currentPage = 1;
    searchQuery;
  });

  let totalPages = $derived(Math.ceil(filteredAds.length / pageSize));
  let paginatedAds = $derived(
    filteredAds.slice((currentPage - 1) * pageSize, currentPage * pageSize)
  );

  function openModal(ad?: BackendAd) {
    if (ad) {
      editingAd = ad;
      formData = {
        name: ad.name,
        storeTypes: ad.storeTypes || [],
        playlistOrder: ad.playlistOrder,
        durationSeconds: ad.durationSeconds || 30,
        startDate: ad.startDate || '',
        endDate: ad.endDate || '',
      };
    } else {
      editingAd = null;
      formData = {
        name: '',
        storeTypes: ['RETAIL'],
        playlistOrder: 0,
        durationSeconds: 30,
        startDate: '',
        endDate: '',
      };
      uploadFile = null;
    }
    showModal = true;
  }

  function closeModal() {
    showModal = false;
    editingAd = null;
    uploadFile = null;
    uploadProgress = 0;
    isDragOver = false;
  }

  function handleFileChange(e: Event) {
    const input = e.target as HTMLInputElement;
    if (input.files && input.files[0]) {
      uploadFile = input.files[0];
    }
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    isDragOver = false;
    const file = e.dataTransfer?.files[0];
    if (file && (file.type.startsWith('video/') || file.type.startsWith('image/'))) {
      uploadFile = file;
    }
  }

  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    isDragOver = true;
  }

  function handleDragLeave() {
    isDragOver = false;
  }

  async function handleSubmit() {
    try {
      if (editingAd) {
        const updateData = {
          name: formData.name,
          playlistOrder: formData.playlistOrder,
          storeTypes: formData.storeTypes,
          durationSeconds: formData.durationSeconds,
          startDate: formData.startDate || null,
          endDate: formData.endDate || null,
        };
        const response = await cmsApi.put(`/ads/${editingAd.id}`, updateData);
        if (response.success) {
          closeModal();
          loadAds();
        } else {
          error = response.error?.message || 'Update failed';
        }
      } else {
        if (!uploadFile) {
          error = 'Please select a file';
          return;
        }

        const form = new FormData();
        form.append('file', uploadFile);
        form.append('name', formData.name);
        form.append('storeTypes', formData.storeTypes.join(','));
        form.append('playlistOrder', String(formData.playlistOrder));
        form.append('durationSeconds', String(formData.durationSeconds));
        if (formData.startDate) form.append('startDate', formData.startDate);
        if (formData.endDate) form.append('endDate', formData.endDate);

        const xhr = new XMLHttpRequest();

        xhr.upload.addEventListener('progress', (e) => {
          if (e.lengthComputable) {
            uploadProgress = (e.loaded / e.total) * 100;
          }
        });

        xhr.addEventListener('load', () => {
          if (xhr.status >= 200 && xhr.status < 300) {
            closeModal();
            loadAds();
          } else {
            try {
              const errData = JSON.parse(xhr.responseText);
              error = errData.error?.message || 'Upload failed';
            } catch {
              error = 'Upload failed';
            }
          }
        });

        xhr.addEventListener('error', () => {
          error = 'Upload failed';
        });

        const token = localStorage.getItem('token');
        xhr.open('POST', `${import.meta.env.VITE_CMS_API_BASE}/ads/upload`);
        xhr.setRequestHeader('Authorization', `Bearer ${token}`);
        xhr.send(form);
      }
    } catch (e) {
      error = 'Operation failed';
    }
  }

  async function handleToggle(ad: BackendAd) {
    try {
      const response = await cmsApi.post(`/ads/${ad.id}/toggle`);
      if (response.success) {
        loadAds();
      } else {
        error = response.error?.message || 'Toggle failed';
      }
    } catch (e) {
      error = 'Toggle failed';
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Are you sure you want to delete this ad?')) return;
    try {
      const response = await cmsApi.delete(`/ads/${id}`);
      if (response.success) {
        loadAds();
      } else {
        error = response.error?.message || 'Delete failed';
      }
    } catch (e) {
      error = 'Delete failed';
    }
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('id-ID', {
      day: '2-digit',
      month: 'short',
      year: '2-digit',
    });
  }

  function formatFileSize(bytes: number): string {
    if (bytes >= 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
    return `${(bytes / 1024).toFixed(0)} KB`;
  }

  function getDownloadUrl(id: number): string {
    return `${import.meta.env.VITE_CMS_API_BASE}/ads/download/${id}`;
  }

  let activeCount = $derived(ads.filter(a => a.isActive).length);
</script>

<div class="space-y-4 md:space-y-5 page-enter">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
    <div>
      <h1 class="text-xl sm:text-2xl font-bold text-slate-900">Ads Management</h1>
      <p class="text-sm text-slate-500 mt-0.5">{ads.length} ads · {activeCount} active</p>
    </div>
    <button onclick={() => openModal()} class="btn btn-primary w-full sm:w-auto">
      <Upload size={17} />
      Upload Ad
    </button>
  </div>

  <!-- Search -->
  <div class="relative">
    <Search class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" size={18} />
    <input
      type="text"
      bind:value={searchQuery}
      placeholder="Search ads by name..."
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

  {#if error}
    <div class="p-4 rounded-xl text-red-700 flex items-center gap-3 text-sm"
         style="background: #fef2f2; border: 1px solid #fecaca;">
      {error}
    </div>
  {/if}

  {#if loading}
    <!-- Skeleton grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
      {#each [1,2,3,4,5,6] as _}
        <div class="skeleton-card overflow-hidden">
          <div class="aspect-video skeleton"></div>
          <div class="p-4 space-y-2">
            <div class="skeleton-text w-3/4"></div>
            <div class="skeleton-text w-1/2 h-3"></div>
            <div class="flex gap-2 mt-3">
              <div class="skeleton h-8 flex-1 rounded-lg"></div>
              <div class="skeleton h-8 w-8 rounded-lg"></div>
              <div class="skeleton h-8 w-8 rounded-lg"></div>
            </div>
          </div>
        </div>
      {/each}
    </div>
  {:else if paginatedAds.length === 0}
    <div class="card p-12 text-center">
      <div class="w-16 h-16 rounded-2xl flex items-center justify-center mx-auto mb-4"
           style="background: #f8fafc; border: 2px dashed #cbd5e1;">
        <Image size={28} class="text-slate-300" />
      </div>
      <p class="text-slate-500 font-medium">No ads found</p>
      <p class="text-slate-400 text-sm mt-1">Upload your first ad to get started</p>
      <button onclick={() => openModal()} class="btn btn-primary mt-5">
        <Upload size={16} />
        Upload Ad
      </button>
    </div>
  {:else}
    <!-- Grid — 1 col on xs, 2 on sm, 3 on md+ -->
    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
      {#each paginatedAds as ad}
        <div class="card overflow-hidden group card-hover">
          <!-- Preview -->
          <div class="aspect-video bg-slate-100 relative overflow-hidden">
            {#if ad.mimeType?.startsWith('video/')}
              <video
                src={getDownloadUrl(ad.id)}
                class="w-full h-full object-cover"
                muted
                preload="metadata"
              ></video>
            {:else if ad.mimeType?.startsWith('image/')}
              <img src={getDownloadUrl(ad.id)} alt={ad.name} class="w-full h-full object-cover" loading="lazy" />
            {:else}
              <div class="w-full h-full flex flex-col items-center justify-center gap-2">
                <Image class="text-slate-300" size={36} />
                <p class="text-xs text-slate-400">{ad.mimeType}</p>
              </div>
            {/if}

            <!-- Hover overlay with quick actions -->
            <div class="absolute inset-0 bg-black/0 group-hover:bg-black/40 transition-all duration-200 flex items-center justify-center gap-3 opacity-0 group-hover:opacity-100">
              <button
                onclick={() => openModal(ad)}
                class="w-9 h-9 bg-white/90 rounded-lg flex items-center justify-center text-slate-700 hover:bg-white transition-colors"
                aria-label="Edit"
              >
                <Edit2 size={14} />
              </button>
              <button
                onclick={() => handleToggle(ad)}
                class="w-9 h-9 rounded-lg flex items-center justify-center transition-colors
                       {ad.isActive ? 'bg-red-500/90 text-white hover:bg-red-500' : 'bg-emerald-500/90 text-white hover:bg-emerald-500'}"
                aria-label="Toggle"
              >
                {#if ad.isActive}
                  <ToggleRight size={14} />
                {:else}
                  <ToggleLeft size={14} />
                {/if}
              </button>
              <button
                onclick={() => handleDelete(ad.id)}
                class="w-9 h-9 bg-red-500/90 rounded-lg flex items-center justify-center text-white hover:bg-red-500 transition-colors"
                aria-label="Delete"
              >
                <Trash2 size={14} />
              </button>
            </div>

            <!-- Status badge -->
            <div class="absolute top-2 left-2">
              <span class="badge text-xs font-bold {ad.isActive ? 'bg-emerald-500 text-white' : 'bg-slate-600 text-white'}">
                {ad.isActive ? '● Active' : '○ Inactive'}
              </span>
            </div>

            <!-- Duration badge -->
            <div class="absolute bottom-2 right-2">
              <span class="flex items-center gap-1 px-2 py-0.5 rounded-lg text-xs font-medium text-white"
                    style="background: rgba(0,0,0,0.65); backdrop-filter: blur(4px);">
                <Clock size={11} />
                {ad.durationSeconds || 0}s
              </span>
            </div>

            <!-- Play count badge -->
            <div class="absolute bottom-2 left-2">
              <span class="flex items-center gap-1 px-2 py-0.5 rounded-lg text-xs font-medium text-white"
                    style="background: rgba(0,0,0,0.65); backdrop-filter: blur(4px);">
                <Eye size={11} />
                {ad.playCount}
              </span>
            </div>
          </div>

          <!-- Info + Actions -->
          <div class="p-3.5">
            <div class="mb-2">
              <h3 class="font-semibold text-slate-800 text-sm truncate">{ad.name}</h3>
              <p class="text-xs text-slate-400 mt-0.5">
                {ad.storeTypes?.join(', ') || 'N/A'} · {formatDate(ad.updatedAt)}
              </p>
            </div>

            <!-- Actions row — always show text -->
            <div class="flex items-center gap-2 mt-3">
              <button
                onclick={() => handleToggle(ad)}
                class="flex-1 flex items-center justify-center gap-1.5 py-2 rounded-lg text-xs font-semibold transition-colors
                  {ad.isActive
                    ? 'bg-red-50 text-red-600 hover:bg-red-100'
                    : 'bg-emerald-50 text-emerald-600 hover:bg-emerald-100'}"
              >
                {#if ad.isActive}
                  <ToggleRight size={13} />
                  Deactivate
                {:else}
                  <ToggleLeft size={13} />
                  Activate
                {/if}
              </button>
              <button
                onclick={() => openModal(ad)}
                class="btn-icon text-slate-500 hover:text-blue-600 hover:bg-blue-50 border border-slate-200"
                aria-label="Edit"
              >
                <Edit2 size={14} />
              </button>
              <button
                onclick={() => handleDelete(ad.id)}
                class="btn-icon text-slate-400 hover:text-red-600 hover:bg-red-50 border border-slate-200"
                aria-label="Delete"
              >
                <Trash2 size={14} />
              </button>
            </div>
          </div>
        </div>
      {/each}
    </div>

    <!-- Pagination -->
    {#if totalPages > 1}
      <div class="flex items-center justify-between">
        <p class="text-sm text-slate-500 hidden sm:block">
          Showing {(currentPage - 1) * pageSize + 1}–{Math.min(currentPage * pageSize, filteredAds.length)} of {filteredAds.length}
        </p>
        <div class="flex items-center gap-1">
          <button
            onclick={() => currentPage = Math.max(1, currentPage - 1)}
            disabled={currentPage === 1}
            class="btn-icon text-slate-500 disabled:opacity-40 hover:bg-slate-100"
          >
            <ChevronLeft size={18} />
          </button>
          {#each Array.from({length: totalPages}, (_, i) => i + 1) as pageNum}
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

<!-- Upload/Edit Modal -->
{#if showModal}
  <div class="fixed inset-0 z-50 flex items-end sm:items-center justify-center p-0 sm:p-4">
    <button class="absolute inset-0 bg-black/60" style="backdrop-filter: blur(3px);" onclick={closeModal} aria-label="Close"></button>
    <div class="bg-white w-full sm:max-w-md sm:rounded-2xl rounded-t-2xl shadow-2xl max-h-[95vh] overflow-y-auto relative modal-content">
      <div class="flex items-center justify-between p-5 border-b border-slate-100">
        <h3 class="text-lg font-bold text-slate-900">{editingAd ? 'Edit Ad' : 'Upload Ad'}</h3>
        <button onclick={closeModal} class="btn-icon text-slate-400 hover:text-slate-600 hover:bg-slate-100">
          <X size={20} />
        </button>
      </div>

      <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="p-5 space-y-4">
        {#if !editingAd}
          <div>
            <label class="label" id="file-upload-label">Media File</label>
            <!-- Drag & Drop zone -->
            <div
              class="drop-zone {isDragOver ? 'drag-over' : ''}"
              ondrop={handleDrop}
              ondragover={handleDragOver}
              ondragleave={handleDragLeave}
              role="region"
              aria-labelledby="file-upload-label"
            >
              <input
                type="file"
                accept="video/mp4,video/webm,video/quicktime,image/jpeg,image/png,image/webp"
                onchange={handleFileChange}
                class="hidden"
                id="file-upload"
                aria-labelledby="file-upload-label"
              />
              <label for="file-upload" class="cursor-pointer flex flex-col items-center gap-2">
                {#if uploadFile}
                  <div class="w-12 h-12 rounded-xl flex items-center justify-center"
                       style="background: linear-gradient(135deg, #eff6ff, #dbeafe);">
                    {#if uploadFile.type.startsWith('video/')}
                      <Play class="text-blue-600" size={22} />
                    {:else}
                      <Image class="text-blue-600" size={22} />
                    {/if}
                  </div>
                  <p class="text-sm font-semibold text-slate-700 text-center break-all">{uploadFile.name}</p>
                  <p class="text-xs text-slate-400">{formatFileSize(uploadFile.size)}</p>
                  <p class="text-xs text-blue-500 underline">Change file</p>
                {:else}
                  <div class="w-12 h-12 rounded-xl flex items-center justify-center"
                       style="background: #f8fafc; border: 2px dashed #cbd5e1;">
                    <Upload class="text-slate-400" size={22} />
                  </div>
                  <p class="text-sm font-medium text-slate-600">Drop file or click to browse</p>
                  <p class="text-xs text-slate-400">mp4, webm, mov, jpg, png — max 100MB</p>
                {/if}
              </label>
            </div>

            {#if uploadProgress > 0 && uploadProgress < 100}
              <div class="mt-3">
                <div class="flex justify-between text-xs text-slate-500 mb-1">
                  <span>Uploading...</span>
                  <span>{Math.round(uploadProgress)}%</span>
                </div>
                <div class="h-2 bg-slate-200 rounded-full overflow-hidden">
                  <div
                    class="h-full rounded-full transition-all duration-300"
                    style="width: {uploadProgress}%; background: linear-gradient(90deg, #2563eb, #6366f1);"
                  ></div>
                </div>
              </div>
            {/if}
          </div>
        {/if}

        <div>
          <label class="label" for="name">Ad Name</label>
          <input id="name" type="text" bind:value={formData.name} class="input" required placeholder="e.g. Summer Promo 2026" />
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="label" for="duration">Duration (sec)</label>
            <input id="duration" type="number" bind:value={formData.durationSeconds} class="input" min="5" max="300" />
          </div>
          <div>
            <label class="label" for="playlistOrder">Playlist Order</label>
            <input id="playlistOrder" type="number" bind:value={formData.playlistOrder} class="input" min="0" />
          </div>
        </div>

        <div>
          <label class="label" for="storeTypes">Store Type</label>
          <select id="storeTypes" class="input">
            <option value="RETAIL" selected={formData.storeTypes.includes('RETAIL')}>Retail</option>
            <option value="FOOD" selected={formData.storeTypes.includes('FOOD')}>Food & Beverage</option>
            <option value="FUEL" selected={formData.storeTypes.includes('FUEL')}>Fuel Station</option>
          </select>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="label" for="startDate">Start Date</label>
            <input id="startDate" type="date" bind:value={formData.startDate} class="input" />
          </div>
          <div>
            <label class="label" for="endDate">End Date</label>
            <input id="endDate" type="date" bind:value={formData.endDate} class="input" />
          </div>
        </div>

        <div class="flex gap-3 pt-2">
          <button type="button" onclick={closeModal} class="btn btn-secondary flex-1">Cancel</button>
          <button type="submit" class="btn btn-primary flex-1">
            {editingAd ? 'Save Changes' : 'Upload'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
