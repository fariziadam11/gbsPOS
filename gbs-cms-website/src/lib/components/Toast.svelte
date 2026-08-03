<script lang="ts">
  import { X, CheckCircle, AlertCircle, Info, AlertTriangle } from 'lucide-svelte';

  type ToastType = 'success' | 'error' | 'info' | 'warning';

  interface Toast {
    id: number;
    type: ToastType;
    message: string;
  }

  let toasts = $state<Toast[]>([]);
  let nextId = 0;

  export function showToast(message: string, type: ToastType = 'info') {
    const id = nextId++;
    toasts = [...toasts, { id, type, message }];

    setTimeout(() => {
      toasts = toasts.filter(t => t.id !== id);
    }, 4000);
  }

  function dismiss(id: number) {
    toasts = toasts.filter(t => t.id !== id);
  }

  const icons = {
    success: CheckCircle,
    error: AlertCircle,
    info: Info,
    warning: AlertTriangle,
  };

  const styles = {
    success: 'bg-green-50 border-green-200 text-green-800',
    error: 'bg-red-50 border-red-200 text-red-800',
    info: 'bg-blue-50 border-blue-200 text-blue-800',
    warning: 'bg-amber-50 border-amber-200 text-amber-800',
  };
</script>

<div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
  {#each toasts as toast (toast.id)}
    {@const Icon = icons[toast.type]}
    <div
      class="flex items-center gap-3 px-4 py-3 rounded-lg border shadow-lg min-w-[300px] max-w-[400px] animate-slide-in {styles[toast.type]}"
      role="alert"
    >
      <Icon size={20} />
      <span class="flex-1 text-sm">{toast.message}</span>
      <button
        onclick={() => dismiss(toast.id)}
        class="p-1 hover:bg-black/10 rounded"
        aria-label="Dismiss"
      >
        <X size={16} />
      </button>
    </div>
  {/each}
</div>

<style>
  @keyframes slide-in {
    from {
      transform: translateX(100%);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }

  .animate-slide-in {
    animation: slide-in 0.3s ease-out;
  }
</style>
