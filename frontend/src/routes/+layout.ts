export const ssr = false;

import { authStore } from '$lib/auth.svelte';
import { reposStore } from '$lib/repos.svelte';
import { incidentsStore } from '$lib/incidents.svelte';
import { healthStore } from '$lib/health.svelte';

export async function load() {
  if (typeof window !== 'undefined') {
    // We only need to check once on initial load, since it's an SPA
    if (authStore.authLoading) {
      await authStore.checkAuth();
      if (authStore.isAuthenticated) {
        reposStore.fetchRepos();
        incidentsStore.fetchIncidents();
        healthStore.fetchHealth();
        healthStore.setupHealthPolling();
      }
    }
  }
  return {};
}
