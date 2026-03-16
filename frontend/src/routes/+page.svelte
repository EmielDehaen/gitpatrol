<script lang="ts">
  import { onMount } from 'svelte';
  import { fade } from 'svelte/transition';
  import { api } from '$lib/api.svelte';
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
    const ws = new WebSocket('ws://localhost:8080/ws');
    ws.onmessage = () => {
      if (api.isAuthenticated) {
        api.fetchRepos();
        api.fetchIncidents();
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
  <div class="fixed inset-0 bg-black/85 backdrop-blur-3xl z-[1000] flex items-center justify-center p-10">
    <div class="text-center">
      <h1 class="text-4xl font-black mb-4 tracking-tighter bg-gradient-to-b from-white to-[#555] bg-clip-text text-transparent">Authenticating...</h1>
      <p class="text-[var(--efinity-text-muted)]">Verifying tactical clearance.</p>
    </div>
  </div>
{:else if !api.isAuthenticated}
  <AuthModal />
{:else}
  <div class="max-w-[1400px] mx-auto px-10 py-20" transition:fade>
    <Header bind:viewMode bind:showIncidentModal bind:showUserModal />

    {#if viewMode === 'grid'}
      <div class="grid grid-cols-[repeat(auto-fill,minmax(400px,1fr))] gap-8">
        {#each api.repositories as repo (repo.id)}
          <RepoCard {repo} bind:selectedRepo />
        {/each}
      </div>
    {:else}
      <div class="flex flex-col gap-3">
        {#each api.repositories as repo (repo.id)}
          <RepoListRow {repo} bind:selectedRepo />
        {/each}
      </div>
    {/if}

    {#if api.repositories.length === 0}
      <div class="text-center p-[120px_40px] glass-panel rounded-[32px]">
        <h2 class="text-[2rem] font-bold mb-4">No Assets Under Patrol</h2>
        <p class="text-[var(--efinity-text-muted)] mb-10">Deploy your first patrol to start monitoring repositories.</p>
        <button class="premium-button rounded-xl px-8 py-4 font-bold text-sm" onclick={() => showAddModal = true}>DEPLOY FIRST PATROL</button>
      </div>
    {/if}

    <div class="fixed bottom-12 right-12 fab" data-tooltip="Deploy New Patrol">
      <button class="premium-button p-6 rounded-3xl z-[100] transition-all duration-400 hover:scale-120 hover:rotate-90" onclick={() => showAddModal = true} aria-label="Deploy New Patrol">
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
