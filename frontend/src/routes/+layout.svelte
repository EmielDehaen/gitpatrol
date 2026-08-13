<script lang="ts">
  import '../app.scss';
  import { onMount } from 'svelte';
  import { API_URL } from '$lib/client';
  import { authStore } from '$lib/auth.svelte';
  import { reposStore } from '$lib/repos.svelte';
  import { incidentsStore } from '$lib/incidents.svelte';
  import { healthStore } from '$lib/health.svelte';
  import { toastHandler } from '$lib/toast.svelte';

  import AuthModal from '$lib/components/Modals/AuthModal.svelte';
  import Sidebar from '$lib/components/Layout/Sidebar.svelte';
  import ToastContainer from '$lib/components/Layout/ToastContainer.svelte';

  let { children } = $props();

  onMount(() => {
    const wsBaseUrl = API_URL ? API_URL.replace('http', 'ws') : (window.location.protocol === 'https:' ? 'wss:' : 'ws:') + '//' + window.location.host;
    const wsUrl = wsBaseUrl + '/ws';
    const ws = new WebSocket(wsUrl);
    ws.onmessage = (event) => {
      if (authStore.isAuthenticated) {
        try {
          const data = JSON.parse(event.data);
          if (data.type === 'status_update') {
            const repos = $state.snapshot(reposStore.repositories);
            const repo = repos.find(r => r.id === Number(data.id));
            if (repo) {
              if (data.status === 'synced') {
                toastHandler.showToast(`Sync completed for ${repo.name}`, 'success');
              } else if (data.status === 'error') {
                toastHandler.showToast(`Sync failed for ${repo.name}: ${data.error}`, 'error');
              }
            }
            reposStore.fetchRepos();
            incidentsStore.fetchIncidents();
          } else if (data.type === 'health_update') {
            healthStore.healthStatus = data.health;
            healthStore.setupHealthPolling();
          }
        } catch (e) {
            reposStore.fetchRepos();
            incidentsStore.fetchIncidents();
        }
      }
    };
    return () => ws.close();
  });
</script>

<ToastContainer />

{#if authStore.authLoading}
  <div class="modal-overlay">
    <div style="text-align: center;">
      <h1 style="font-size: 2rem;">Authenticating...</h1>
      <p style="color: var(--on-surface-variant); margin-top: 16px;">Verifying tactical clearance.</p>
    </div>
  </div>
{:else if !authStore.isAuthenticated}
  <AuthModal />
{:else}
  <div class="layout-wrapper">
    <Sidebar />
    <div class="main-content">
      {@render children()}
    </div>
  </div>
{/if}

<style>
  .layout-wrapper {
    display: flex;
    min-height: 100vh;
  }
  .main-content {
    flex: 1;
    margin-left: 280px;
    display: flex;
    flex-direction: column;
    min-width: 0;

    @media (max-width: 1439px) { margin-left: 250px; }
    @media (max-width: 1023px) { margin-left: 80px; }
    @media (max-width: 767px) { margin-left: 0; padding-bottom: 70px; }
  }
</style>
