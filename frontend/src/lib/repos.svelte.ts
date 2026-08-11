import { apiFetch } from './client';
import { toastHandler } from './toast.svelte';
import type { Repository } from './types';

class ReposStore {
  repositories = $state<Repository[]>([]);

  constructor() {
    if (typeof window !== 'undefined') {
      window.addEventListener('gitpatrol:logged_out', () => {
        this.repositories = [];
      });
    }
  }

  async fetchRepos() {
    try {
      const res = await apiFetch('/api/repositories');
      if (res.ok) {
        this.repositories = await res.json();
      }
    } catch (e) {
      toastHandler.showToast('Network error: could not load repositories.', 'error');
    }
  }
}

export const reposStore = new ReposStore();
