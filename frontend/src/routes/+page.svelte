<script lang="ts">
  import { onMount } from 'svelte';
  import { fade } from 'svelte/transition';
  import { API_URL } from '$lib/client';
  import { toastHandler } from '$lib/toast.svelte';
  import { authStore } from '$lib/auth.svelte';
  import { reposStore } from '$lib/repos.svelte';
  import { incidentsStore } from '$lib/incidents.svelte';
  import { healthStore } from '$lib/health.svelte';
  import { getProgress } from '$lib/utils';
  import { type Repository } from '$lib/types';

  // Components
  import Header from '$lib/components/Layout/Header.svelte';
  import Sidebar from '$lib/components/Layout/Sidebar.svelte';
  import ToastContainer from '$lib/components/Layout/ToastContainer.svelte';
  import RepoCard from '$lib/components/Repo/RepoCard.svelte';
  import RepoListRow from '$lib/components/Repo/RepoListRow.svelte';
  import RepoDetail from '$lib/components/Repo/RepoDetail.svelte';
  import AuthModal from '$lib/components/Modals/AuthModal.svelte';
  import AddRepoModal from '$lib/components/Modals/AddRepoModal.svelte';
  import ConfigModal from '$lib/components/Modals/ConfigModal.svelte';
  import IncidentModal from '$lib/components/Modals/IncidentModal.svelte';
  import UserModal from '$lib/components/Modals/UserModal.svelte';
  import Icon from '$lib/components/Icon.svelte';

  let viewMode = $state<'grid' | 'list'>('grid');
  let selectedRepo = $state<Repository | null>(null);
  let showAddModal = $state(false);
  let showConfigModal = $state(false);
  let showIncidentModal = $state(false);
  let showUserModal = $state(false);

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
            healthStore.setupHealthPolling(); // Reset the 1-minute fallback timer
          }
        } catch (e) {
            reposStore.fetchRepos();
            incidentsStore.fetchIncidents();
        }
      }
    };
    const timer = setInterval(() => {
      reposStore.repositories = reposStore.repositories.map(r => ({ ...r, progress: getProgress(r) }));
    }, 1000);
    return () => clearInterval(timer);
  });
</script>

<ToastContainer />

{#if authStore.authLoading}
  <div class="modal-overlay">
    <div style="text-align: center;">
      <h1 style="font-size: 2rem;">Authenticating...</h1>
      <p style="color: var(--efinity-text-muted); margin-top: 16px;">Verifying tactical clearance.</p>
    </div>
  </div>
{:else if !authStore.isAuthenticated}
  <AuthModal />
{:else}
  <div class="layout-wrapper" transition:fade>
    <Sidebar />
    <div class="main-content">
      <Header bind:viewMode bind:showIncidentModal />
      
      <div class="hero">
        <div class="hero-stats">
          <div class="stat-block">
            <span class="stat-label">GLOBAL UPTIME</span>
            <span class="stat-value">99.998%</span>
          </div>
          <div class="stat-block">
            <span class="stat-label">TOTAL PROTECTED</span>
            <span class="stat-value">1.4TB</span>
          </div>
          <div class="stat-block">
            <span class="stat-label">ACTIVE MIRRORS</span>
            <span class="stat-value">12/12</span>
          </div>
        </div>
        
        <div class="hero-filters">
          <div class="tab-container" style="width: auto; margin-bottom: 0;">
            <button class="tab-btn active">ALL</button>
            <button class="tab-btn">SYNCING</button>
            <button class="tab-btn">SECURED</button>
            <button class="tab-btn">FAILED</button>
          </div>
        </div>
      </div>

      <div class="content-area">
        {#if viewMode === 'grid'}
          <div class="repo-grid">
            {#each reposStore.repositories as repo (repo.id)}
              <RepoCard {repo} bind:selectedRepo />
            {/each}
          </div>
        {:else}
          <div class="repo-list">
            {#each reposStore.repositories as repo (repo.id)}
              <RepoListRow {repo} bind:selectedRepo />
            {/each}
          </div>
        {/if}

        {#if reposStore.repositories.length === 0}
          <div style="text-align: center; padding: 120px 40px; background: var(--surface-container-low); border-radius: 32px;">
            <h2 style="font-size: 2rem; margin-bottom: 16px;">No Repositories</h2>
            <p style="color: var(--on-surface-variant); margin-bottom: 32px;">Deploy your first sentinel to begin mirroring.</p>
            <button onclick={() => showAddModal = true}>ADD REPOSITORY</button>
          </div>
        {/if}
      </div>

      <div class="fab" data-tooltip="Deploy New Patrol">
        <button onclick={() => showAddModal = true} aria-label="Deploy New Patrol">
          <Icon name="plus" size={28} stroke="white" strokeWidth={3} />
        </button>
      </div>
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
  }
  .hero {
    padding: 0 40px 40px 40px;
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
  }
  .hero-stats {
    display: flex;
    gap: 40px;
  }
  .stat-block {
    display: flex;
    flex-direction: column;
  }
  .stat-label {
    font-size: 0.65rem;
    font-weight: 800;
    letter-spacing: 0.1em;
    color: var(--on-surface-variant);
    margin-bottom: 4px;
  }
  .stat-value {
    font-size: 2.5rem;
    font-weight: 700;
    font-family: 'Space Grotesk', sans-serif;
    color: var(--on-surface);
    line-height: 1;
  }
  .content-area {
    padding: 0 40px 80px 40px;
  }
</style>
