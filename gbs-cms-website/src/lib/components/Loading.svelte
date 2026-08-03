<script lang="ts">
  interface Props {
    size?: 'sm' | 'md' | 'lg';
    text?: string;
    fullPage?: boolean;
  }

  let { size = 'md', text = '', fullPage = false }: Props = $props();

  const sizes = {
    sm: 'w-5 h-5 border-2',
    md: 'w-8 h-8 border-2',
    lg: 'w-12 h-12 border-3',
  };
</script>

{#if fullPage}
  <div class="min-h-[400px] flex flex-col items-center justify-center gap-4">
    <div class="animate-spin rounded-full border-blue-600 border-t-transparent {sizes[size]}"></div>
    {#if text}
      <p class="text-slate-500">{text}</p>
    {/if}
  </div>
{:else}
  <div class="flex items-center justify-center gap-3 py-8">
    <div class="animate-spin rounded-full border-blue-600 border-t-transparent {sizes[size]}"></div>
    {#if text}
      <p class="text-slate-500">{text}</p>
    {/if}
  </div>
{/if}
