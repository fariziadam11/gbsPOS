// lib/components barrel export
export { default as Layout } from './Layout.svelte';
export { default as Toast } from './Toast.svelte';
export { default as Loading } from './Loading.svelte';
export { default as EmptyState } from './EmptyState.svelte';

// Toast store for showing notifications
import { writable } from 'svelte/store';
type ToastType = 'success' | 'error' | 'info' | 'warning';

interface ToastItem {
  id: number;
  type: ToastType;
  message: string;
}

function createToastStore() {
  const { subscribe, update } = writable<ToastItem[]>([]);
  let nextId = 0;

  return {
    subscribe,
    show(message: string, type: ToastType = 'info') {
      const id = nextId++;
      update(toasts => [...toasts, { id, type, message }]);
      setTimeout(() => {
        update(toasts => toasts.filter(t => t.id !== id));
      }, 4000);
    },
    success(message: string) { this.show(message, 'success'); },
    error(message: string) { this.show(message, 'error'); },
    warning(message: string) { this.show(message, 'warning'); },
    info(message: string) { this.show(message, 'info'); },
  };
}

export const toast = createToastStore();
