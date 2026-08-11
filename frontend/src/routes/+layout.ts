export const ssr = false;

import { api } from '$lib/api.svelte';

export async function load() {
  if (typeof window !== 'undefined') {
    // We only need to check once on initial load, since it's an SPA
    if (api.authLoading) {
      await api.checkAuth();
    }
  }
  return {};
}
