<script lang="ts">
  import { onMount } from 'svelte';
  import { fade } from 'svelte/transition';
  import { api, API_URL, toastHandler } from '$lib/api.svelte';
  import { getProgress } from '$lib/utils';
  import { type Repository } from '$lib/types';

  // Components
  import Header from '$lib/components/Layout/Header.svelte';
  import ToastContainer from '$lib/components/Layout/ToastContainer.svelte';
  import RepoCard from '$lib/components/Repo/RepoCard.svelte';
  import RepoListRow from '$lib/components/Repo/RepoListRow.svelte';
  import RepoDetail from '$lib/components/Repo/RepoDetail.svelte';
  import AuthModal from '$lib/components/Modals/AuthModal.svelte';
  import AddRepoModal from '$lib/components/Modals/AddRepoModal.svelte';
  import ConfigModal from '$lib/components/Modals/ConfigModal.svelte';
  import IncidentModal from '$lib/components/Modals/IncidentModal.svelte';
  import UserModal from '$lib/components/Modals/UserModal.svelte';

  let viewMode = $state<'grid' | 'list'>('grid');
  let selectedRepo = $state<Repository | null>(null);
  let showAddModal = $state(false);
  let showConfigModal = $state(false);
  let showIncidentModal = $state(false);
  let showUserModal = $state(false);

  onMount(() => {
    const wsUrl = API_URL.replace('http', 'ws') + '/ws';
    const ws = new WebSocket(wsUrl);
    ws.onmessage = (event) => {
      if (api.isAuthenticated) {
        try {
          const data = JSON.parse(event.data);
          if (data.type === 'status_update') {
            const repos = $state.snapshot(api.repositories);
            const repo = repos.find(r => r.id === Number(data.id));
            if (repo) {
              if (data.status === 'synced') {
                toastHandler.showToast(`Sync completed for ${repo.name}`, 'success');
              } else if (data.status === 'error') {
                toastHandler.showToast(`Sync failed for ${repo.name}: ${data.error}`, 'error');
              }
            }
            api.fetchRepos();
            api.fetchIncidents();
          } else if (data.type === 'health_update') {
            api.healthStatus = data.health;
            api.setupHealthPolling(); // Reset the 1-minute fallback timer
          }
        } catch (e) {
          api.fetchRepos();
          api.fetchIncidents();
        }
      }
    };
    const timer = setInterval(() => {
      api.repositories = api.repositories.map(r => ({ ...r, progress: getProgress(r) }));
    }, 1000);
    return () => clearInterval(timer);
  });
</script>

<ToastContainer />

{#if api.authLoading}
  <div class="modal-overlay">
    <div style="text-align: center;">
      <h1 style="font-size: 2rem;">Authenticating...</h1>
      <p style="color: var(--efinity-text-muted); margin-top: 16px;">Verifying tactical clearance.</p>
    </div>
  </div>
{:else if !api.isAuthenticated}
  <AuthModal />
{:else}
  <div class="container" transition:fade>
    <Header bind:viewMode bind:showIncidentModal bind:showUserModal />

    {#if viewMode === 'grid'}
      <div class="repo-grid">
        {#each api.repositories as repo (repo.id)}
          <RepoCard {repo} bind:selectedRepo />
        {/each}
      </div>
    {:else}
      <div class="repo-list">
        {#each api.repositories as repo (repo.id)}
          <RepoListRow {repo} bind:selectedRepo />
        {/each}
      </div>
    {/if}

    {#if api.repositories.length === 0}
      <div style="text-align: center; padding: 120px 40px; background: var(--glass); border-radius: 32px; border: 1px solid var(--glass-border);">
        <h2 style="font-size: 2rem; margin-bottom: 16px;">No Assets Under Patrol</h2>
        <p style="color: var(--efinity-text-muted); margin-bottom: 40px;">Deploy your first patrol to start monitoring repositories.</p>
        <button onclick={() => showAddModal = true}>DEPLOY FIRST PATROL</button>
      </div>
    {/if}

    <div class="fab" data-tooltip="Deploy New Patrol">
      <button onclick={() => showAddModal = true} aria-label="Deploy New Patrol">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
      </button>
    </div>
  </div>

  {#if selectedRepo}
    <RepoDetail bind:repo={selectedRepo} bind:selectedRepo bind:showConfigModal />
  {/if}

  {#if showAddModal}
    <AddRepoModal bind:show={showAddModal} />
  {/if}

  {#if showConfigModal && selectedRepo}
    <ConfigModal bind:show={showConfigModal} repo={selectedRepo} bind:selectedRepo />
  {/if}

  {#if showIncidentModal}
    <IncidentModal bind:show={showIncidentModal} />
  {/if}

  {#if showUserModal}
    <UserModal bind:show={showUserModal} />
  {/if}
{/if}
