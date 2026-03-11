<script lang="ts">
  import { api } from '$lib/api.svelte';
  import { suggestName, humanToMinutes } from '$lib/utils';
  import { fade } from 'svelte/transition';

  let { show = $bindable() } = $props();

  let url = $state('');
  let name = $state('');
  let interval = $state('1h');
  let autoPatrol = $state(true);

  $effect(() => {
    if (url && !name) {
      name = suggestName(url);
    }
  });

  async function handleAdd() {
    if (!name || !url) return;
    try {
      const res = await api.apiFetch('/api/repositories', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          name, 
          url, 
          interval_minutes: humanToMinutes(interval), 
          auto_patrol: autoPatrol ? 1 : 0 
        })
      });
      if (res.ok) {
        url = ''; name = ''; autoPatrol = true; interval = '1h'; show = false;
        api.fetchRepos();
        api.showToast('Patrol deployed successfully!', 'success');
      } else {
        const err = await res.json();
        api.showToast(err.error || 'Failed to deploy patrol.', 'error');
      }
    } catch (e) {}
  }
</script>

<div class="modal-overlay" onclick={() => show = false} transition:fade>
  <div class="modal-content" style="max-width: 500px;" onclick={(e) => e.stopPropagation()}>
    <div class="modal-header">
      <h2 style="font-size: 1.5rem; margin: 0;">Deploy New Patrol</h2>
    </div>
    <div class="modal-body">
      <label>REPOSITORY URL</label>
      <input type="text" bind:value={url} placeholder="https://github.com/user/repo" />
      
      <label>ASSET NAME</label>
      <input type="text" bind:value={name} placeholder="Mission Name" />
      
      <div style="display: flex; gap: 20px; margin-bottom: 32px;">
        <div style="flex: 1;">
          <label>SCAN INTERVAL</label>
          <input type="text" bind:value={interval} placeholder="e.g. 1h, 30m, 1d" style="margin-bottom: 0;" />
        </div>
        <div style="display: flex; align-items: center; gap: 12px; margin-top: 24px;">
          <input type="checkbox" bind:checked={autoPatrol} style="width: 20px; height: 20px; margin: 0; cursor: pointer;" />
          <label style="margin: 0; cursor: pointer;">AUTO PATROL</label>
        </div>
      </div>

      <div style="display: flex; gap: 16px;">
        <button style="flex: 2;" onclick={handleAdd}>INITIATE DEPLOYMENT</button>
        <button class="secondary" style="flex: 1;" onclick={() => show = false}>CANCEL</button>
      </div>
    </div>
  </div>
</div>
