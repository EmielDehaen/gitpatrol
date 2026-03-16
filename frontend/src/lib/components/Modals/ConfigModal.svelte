<script lang="ts">
  import { api } from '$lib/api.svelte';
  import { humanToMinutes, minutesToHuman } from '$lib/utils';
  import { type Repository } from '$lib/types';
  import { fade } from 'svelte/transition';

  let { show = $bindable(), repo = $bindable(), selectedRepo = $bindable() } = $props<{ show: boolean, repo: Repository, selectedRepo: Repository | null }>();

  let intervalString = $state(minutesToHuman(repo.interval_minutes));

  async function updateConfig() {
    try {
      const res = await api.apiFetch(`/api/repositories/${repo.id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          interval_minutes: humanToMinutes(intervalString), 
          auto_patrol: repo.auto_patrol 
        })
      });
      if (res.ok) { 
        await api.fetchRepos();
        selectedRepo = api.repositories.find(r => r.id === repo.id) || null;
        show = false; 
        api.showToast('Configuration updated.', 'success');
      } else {
        api.showToast('Failed to update configuration.', 'error');
      }
    } catch (e) {}
  }

  async function deleteRepo() {
    if (!confirm('Are you sure you want to terminate this patrol?')) return;
    try {
      const res = await api.apiFetch(`/api/repositories/${repo.id}`, { method: 'DELETE' });
      if (res.ok) { 
        selectedRepo = null; show = false; api.fetchRepos(); 
        api.showToast('Patrol terminated.', 'info');
      }
    } catch (e) {}
  }
</script>

<div class="modal-overlay" onclick={() => show = false} onkeydown={(e) => e.key == 'ESCAPE' && (show = false)} role='presentation' transition:fade>
  <div class="modal-content" style="max-width: 500px;" onclick={(e) => e.stopPropagation()} role='none'>
    <div class="modal-header">
      <h2 style="font-size: 1.5rem; margin: 0;">Configure Patrol: {repo.name}</h2>
    </div>
    <div class="modal-body">
      <label for='interval'>SCAN INTERVAL</label>
      <input id='interval' type="text" bind:value={intervalString} placeholder="e.g. 1h, 30m, 1d" />
      
      <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 40px;">
        <input id='auto_patrol' type="checkbox" bind:checked={repo.auto_patrol} style="width: 20px; height: 20px; margin: 0; cursor: pointer;" />
        <label for='auto_patrol' style="margin: 0; cursor: pointer;">ACTIVE AUTO PATROL</label>
      </div>

      <div style="display: flex; gap: 16px; margin-bottom: 32px;">
        <button style="flex: 2;" onclick={updateConfig}>UPDATE CONFIG</button>
        <button class="secondary" style="flex: 1;" onclick={() => show = false}>CANCEL</button>
      </div>

      <div style="border-top: 1px solid var(--status-red); opacity: 0.3; margin-bottom: 32px;"></div>

      <button class="secondary" style="width: 100%; color: var(--status-red); border-color: rgba(255, 77, 77, 0.2);" onclick={deleteRepo}>
        TERMINATE PATROL (DELETE)
      </button>
    </div>
  </div>
</div>
