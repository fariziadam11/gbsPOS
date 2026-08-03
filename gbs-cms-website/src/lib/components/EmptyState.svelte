<script lang="ts">
  import type { Component } from 'svelte';

  interface Props {
    icon?: Component;
    title?: string;
    description?: string;
  }

  let { icon: Icon, title = 'No data', description = '' }: Props = $props();
</script>

<div class="flex flex-col items-center justify-center py-12 text-center">
  {#if Icon}
    <div class="w-16 h-16 bg-slate-100 rounded-full flex items-center justify-center mb-4">
      <svelte:component this={Icon} size={32} class="text-slate-400" />
    </div>
  {/if}
  <h3 class="text-lg font-medium text-slate-600 mb-1">{title}</h3>
  {#if description}
    <p class="text-sm text-slate-400 max-w-sm">{description}</p>
  {/if}
</div>
