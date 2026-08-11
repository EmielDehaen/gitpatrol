import { apiFetch } from './client';
import { toastHandler } from './toast.svelte';
import type { Incident } from './types';

class IncidentsStore {
  incidents = $state<Incident[]>([]);

  constructor() {
    if (typeof window !== 'undefined') {
      window.addEventListener('gitpatrol:logged_out', () => {
        this.incidents = [];
      });
    }
  }

  async fetchIncidents() {
    try {
      const res = await apiFetch('/api/incidents');
      if (res.ok) {
        this.incidents = await res.json();
      }
    } catch (e) {
      toastHandler.showToast('Network error: could not load incidents.', 'error');
    }
  }

  async clearIncidents() {
    try {
      const res = await apiFetch('/api/incidents', { method: 'DELETE' });
      if (res.ok) {
        this.incidents = [];
        toastHandler.showToast('All incidents cleared.', 'info');
      }
    } catch (e) {
      toastHandler.showToast('Network error: could not clear incidents.', 'error');
    }
  }
}

export const incidentsStore = new IncidentsStore();
