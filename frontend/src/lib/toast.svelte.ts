import type { Toast } from './types';

class ToastHandler {
  toasts = $state<Toast[]>([]);
  private toastId = 0

  showToast(message: string, type: 'success' | 'error' | 'info' = 'info') {
    // Deduplicate: don't show the same message+type if it's already visible
    if (this.toasts.some(t => t.message === message && t.type === type)) return;
    const id = this.toastId++;
    this.toasts = [...this.toasts, { id, message, type }];
    setTimeout(() => {
      this.toasts = this.toasts.filter(t => t.id !== id);
    }, 5000);
  }
}

export const toastHandler = new ToastHandler();
