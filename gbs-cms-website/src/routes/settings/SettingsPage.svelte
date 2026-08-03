<script lang="ts">
  import { onMount } from 'svelte';
  import { Settings, Save, RefreshCw, ChevronDown, ChevronUp, Store, DollarSign, Monitor, Bell, ShoppingBag, Check } from 'lucide-svelte';
  import { cmsApi } from '../../lib/api/client';

  interface SettingsResponse {
    settings: Record<string, string>;
  }

  let settings = $state<Record<string, string>>({});
  let loading = $state(true);
  let saving = $state(false);
  let error = $state('');
  let success = $state('');
  let expandedGroups = $state<Set<string>>(new Set(['General']));

  async function loadSettings() {
    loading = true;
    error = '';

    try {
      const response = await cmsApi.get<SettingsResponse>('/settings');

      if (response.success && response.data) {
        settings = response.data.settings || {};
      } else {
        error = response.error?.message || 'Failed to load settings';
      }
    } catch (e) {
      error = 'Connection error';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadSettings();
  });

  function getSettingValue(key: string): string {
    return settings[key] || '';
  }

  function updateSetting(key: string, value: string) {
    settings[key] = value;
  }

  function toggleGroup(title: string) {
    if (expandedGroups.has(title)) {
      expandedGroups.delete(title);
    } else {
      expandedGroups.add(title);
    }
    expandedGroups = new Set(expandedGroups);
  }

  async function handleSave() {
    saving = true;
    error = '';
    success = '';

    try {
      const payload = { settings: settings };
      const response = await cmsApi.put('/settings', payload);

      if (response.success) {
        success = 'Settings saved successfully';
        setTimeout(() => success = '', 3000);
      } else {
        error = response.error?.message || 'Save failed';
      }
    } catch (e) {
      error = 'Save failed';
    } finally {
      saving = false;
    }
  }

  // Settings grouped by category with icons
  const settingGroups = [
    {
      title: 'General',
      icon: Store,
      color: 'blue',
      settings: [
        { key: 'store_name', label: 'Store Name', type: 'text', placeholder: 'e.g. GBS Supermarket' },
        { key: 'store_address', label: 'Store Address', type: 'text', placeholder: 'Full address' },
        { key: 'store_phone', label: 'Phone Number', type: 'text', placeholder: '+62 xxx xxxx xxxx' },
        { key: 'store_email', label: 'Email', type: 'text', placeholder: 'store@example.com' },
      ],
    },
    {
      title: 'Tax & Pricing',
      icon: DollarSign,
      color: 'emerald',
      settings: [
        { key: 'tax_rate', label: 'Tax Rate (%)', type: 'number', placeholder: '11' },
        { key: 'currency', label: 'Currency Code', type: 'text', placeholder: 'IDR' },
        { key: 'currency_symbol', label: 'Currency Symbol', type: 'text', placeholder: 'Rp' },
      ],
    },
    {
      title: 'Display',
      icon: Monitor,
      color: 'purple',
      settings: [
        { key: 'slideshow_interval', label: 'Slideshow Interval (sec)', type: 'number', placeholder: '10' },
        { key: 'auto_play_ads', label: 'Auto-play Ads', type: 'checkbox', placeholder: '' },
        { key: 'show_prices', label: 'Show Prices on Display', type: 'checkbox', placeholder: '' },
      ],
    },
    {
      title: 'Notifications',
      icon: Bell,
      color: 'amber',
      settings: [
        { key: 'low_stock_threshold', label: 'Low Stock Alert Threshold', type: 'number', placeholder: '10' },
        { key: 'email_notifications', label: 'Email Notifications', type: 'checkbox', placeholder: '' },
      ],
    },
    {
      title: 'POS',
      icon: ShoppingBag,
      color: 'red',
      settings: [
        { key: 'default_payment_method', label: 'Default Payment Method', type: 'text', placeholder: 'CASH' },
        { key: 'cash_drawer_enabled', label: 'Enable Cash Drawer', type: 'checkbox', placeholder: '' },
      ],
    },
  ];

  const colorMap: Record<string, { icon: string; header: string; badge: string }> = {
    blue:   { icon: '#eff6ff', header: '#2563eb', badge: '#bfdbfe' },
    emerald:{ icon: '#f0fdf4', header: '#059669', badge: '#a7f3d0' },
    purple: { icon: '#faf5ff', header: '#7c3aed', badge: '#ddd6fe' },
    amber:  { icon: '#fffbeb', header: '#d97706', badge: '#fde68a' },
    red:    { icon: '#fff1f2', header: '#dc2626', badge: '#fecaca' },
  };
</script>

<div class="space-y-4 md:space-y-5 page-enter">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
    <div>
      <h1 class="text-xl sm:text-2xl font-bold text-slate-900">Settings</h1>
      <p class="text-sm text-slate-500 mt-0.5">Configure your store settings</p>
    </div>
    <div class="flex gap-2">
      <button
        onclick={loadSettings}
        class="btn btn-secondary"
        disabled={loading}
        aria-label="Refresh settings"
      >
        <RefreshCw size={15} class="{loading ? 'animate-spin' : ''}" />
        <span class="hidden sm:inline">Refresh</span>
      </button>
      <button
        onclick={handleSave}
        disabled={saving || loading}
        class="btn btn-primary"
      >
        {#if saving}
          <svg class="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
          </svg>
          Saving...
        {:else}
          <Save size={15} />
          Save Settings
        {/if}
      </button>
    </div>
  </div>

  <!-- Error -->
  {#if error}
    <div class="p-4 rounded-xl text-red-700 flex items-center gap-3 text-sm"
         style="background: #fef2f2; border: 1px solid #fecaca;">
      {error}
    </div>
  {/if}

  <!-- Success toast inline -->
  {#if success}
    <div class="p-4 rounded-xl text-emerald-700 flex items-center gap-3 text-sm"
         style="background: #f0fdf4; border: 1px solid #bbf7d0;">
      <Check size={16} class="flex-shrink-0" />
      {success}
    </div>
  {/if}

  {#if loading}
    <!-- Skeleton -->
    <div class="space-y-3">
      {#each [1,2,3] as _}
        <div class="skeleton-card">
          <div class="flex items-center justify-between p-5 border-b border-slate-100">
            <div class="flex items-center gap-3">
              <div class="skeleton w-8 h-8 rounded-lg"></div>
              <div class="skeleton-text w-28"></div>
            </div>
          </div>
          <div class="p-5 space-y-4">
            {#each [1,2,3] as __}
              <div class="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4">
                <div class="skeleton-text w-36 sm:w-1/3"></div>
                <div class="skeleton h-10 rounded-lg sm:flex-1"></div>
              </div>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <div class="space-y-3">
      {#each settingGroups as group}
        {@const colors = colorMap[group.color]}
        <div class="card overflow-hidden">
          <!-- Collapsible Header -->
          <button
            onclick={() => toggleGroup(group.title)}
            class="w-full flex items-center justify-between p-4 md:p-5 text-left hover:bg-slate-50 transition-colors"
            style="border-bottom: {expandedGroups.has(group.title) ? '1px solid #f1f5f9' : 'none'};"
          >
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0"
                   style="background: {colors.icon};">
                {#if group.icon === Store}
                  <Store size={15} style="color: {colors.header};" />
                {:else if group.icon === DollarSign}
                  <DollarSign size={15} style="color: {colors.header};" />
                {:else if group.icon === Monitor}
                  <Monitor size={15} style="color: {colors.header};" />
                {:else if group.icon === Bell}
                  <Bell size={15} style="color: {colors.header};" />
                {:else if group.icon === ShoppingBag}
                  <ShoppingBag size={15} style="color: {colors.header};" />
                {/if}
              </div>
              <h2 class="font-bold text-slate-800 text-sm md:text-base">{group.title}</h2>
              <span class="text-xs px-2 py-0.5 rounded-full font-medium hidden sm:inline"
                    style="background: {colors.badge}; color: {colors.header};">
                {group.settings.length} settings
              </span>
            </div>
            <div class="text-slate-400 flex-shrink-0">
              {#if expandedGroups.has(group.title)}
                <ChevronUp size={18} />
              {:else}
                <ChevronDown size={18} />
              {/if}
            </div>
          </button>

          <!-- Collapsible Content -->
          {#if expandedGroups.has(group.title)}
            <div class="p-4 md:p-5 space-y-5">
              {#each group.settings as setting}
                <div class="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-6">
                  <label for={setting.key} class="sm:w-2/5 font-medium text-slate-700 text-sm flex-shrink-0">
                    {setting.label}
                  </label>
                  <div class="sm:flex-1">
                    {#if setting.type === 'checkbox'}
                      <!-- Toggle Switch -->
                      <button
                        type="button"
                        role="switch"
                        id={setting.key}
                        aria-checked={getSettingValue(setting.key) === 'true'}
                        onclick={() => updateSetting(setting.key, String(getSettingValue(setting.key) !== 'true'))}
                        class="toggle-switch {getSettingValue(setting.key) === 'true' ? 'bg-blue-600' : 'bg-slate-200'}"
                      >
                        <span
                          class="toggle-thumb {getSettingValue(setting.key) === 'true' ? 'translate-x-5' : 'translate-x-0'}"
                        ></span>
                      </button>
                      <span class="ml-2 text-sm text-slate-500 align-middle">
                        {getSettingValue(setting.key) === 'true' ? 'Enabled' : 'Disabled'}
                      </span>
                    {:else if setting.type === 'number'}
                      <input
                        id={setting.key}
                        type="number"
                        value={getSettingValue(setting.key)}
                        oninput={(e) => updateSetting(setting.key, (e.target as HTMLInputElement).value)}
                        class="input text-sm"
                        placeholder={setting.placeholder}
                      />
                    {:else}
                      <input
                        id={setting.key}
                        type="text"
                        value={getSettingValue(setting.key)}
                        oninput={(e) => updateSetting(setting.key, (e.target as HTMLInputElement).value)}
                        class="input text-sm"
                        placeholder={setting.placeholder}
                      />
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/each}
    </div>

    <!-- Save button at bottom for convenience -->
    <div class="flex justify-end pt-2 pb-6">
      <button
        onclick={handleSave}
        disabled={saving}
        class="btn btn-primary"
      >
        {#if saving}
          <svg class="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
          </svg>
          Saving...
        {:else}
          <Save size={15} />
          Save All Settings
        {/if}
      </button>
    </div>
  {/if}
</div>
