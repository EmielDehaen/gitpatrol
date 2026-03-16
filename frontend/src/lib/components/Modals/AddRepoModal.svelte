<script lang="ts">
  import { api, API_URL } from '$lib/api.svelte';
  import { suggestName, humanToMinutes, getNormalizedUrl, getAvatarUrl } from '$lib/utils';
  import { fade } from 'svelte/transition';

  let { show = $bindable() } = $props();

  let url = $state('');
  let name = $state('');
  let intervalString = $state('1h');
  let autoPatrol = $state(true);

  function handleAvatarError(e: Event) {
    const img = e.target as HTMLImageElement;
    img.src = "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png";
  }

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
          interval_minutes: humanToMinutes(intervalString), 
          auto_patrol: autoPatrol ? 1 : 0 
        })
      });
      if (res.ok) {
        url = ''; name = ''; autoPatrol = true; intervalString = '1h'; show = false;
        api.fetchRepos();
        api.showToast('Patrol deployed successfully!', 'success');
      } else {
        const err = await res.json();
        api.showToast(err.error || 'Failed to deploy patrol.', 'error');
      }
    } catch (e) {}
  }
</script>

<div class="modal-overlay" onclick={() => show = false} onkeydown={(e) => e.key === 'Escape' && (show = false)} role="presentation" transition:fade>
  <div class="modal-content" style="max-width: 540px;" onclick={(e) => e.stopPropagation()} role='none'>
    <div class="modal-header">
      <h2 style="margin: 0; font-size: 1.8rem; font-weight: 800;">Deploy New Patrol</h2>
      <p style="color: var(--efinity-text-muted); margin: 8px 0 0 0; font-size: 0.9rem;">Configure a new asset for monitoring.</p>
    </div>
    <div class="modal-body">
      <label for="url">REPOSITORY URL</label>
      <input id="url" type="text" bind:value={url} placeholder="https://github.com/hoppscotch/hoppscotch" style="margin-bottom: 8px;" />
      
      {#if getNormalizedUrl(url)}
        <div style="margin-bottom: 24px;">
          <a href={getNormalizedUrl(url)} target="_blank" rel="noopener noreferrer" style="font-size: 0.75rem; color: var(--efinity-blue); text-decoration: none; font-family: monospace; opacity: 0.8; display: flex; align-items: center; gap: 6px;">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6M15 3h6v6M10 14L21 3"/></svg>
            {getNormalizedUrl(url)}
          </a>
        </div>
      {/if}

      {#if url.length > 3}
        <div transition:fade style="display: flex; gap: 20px; align-items: center; margin-bottom: 32px; background: rgba(255,255,255,0.02); padding: 16px; border-radius: 16px; border: 1px solid var(--glass-border);">
          <img src={getAvatarUrl(url)} onerror={handleAvatarError} alt="" style="width: 44px; height: 44px; border-radius: 12px; border: 1px solid var(--glass-border);" />
          <div style="flex: 1;">
            <label for="name" style="margin-bottom: 4px;">DISPLAY NAME</label>
            <input id="name" bind:value={name} placeholder="e.g. Hoppscotch" style="margin-bottom: 0;" />
          </div>
        </div>
      {/if}
      
      <div style="margin-bottom: 24px; display: flex; align-items: center; justify-content: space-between; background: rgba(255,255,255,0.02); padding: 20px; border-radius: 16px; border: 1px solid var(--glass-border);">
        <div>
          <div style="font-weight: 700; font-size: 0.9rem;">KEEP ACTIVE PATROL</div>
          <div style="font-size: 0.75rem; color: var(--efinity-text-muted);">Continuously monitor and sync this asset.</div>
        </div>
        <input type="checkbox" bind:checked={autoPatrol} style="width: 24px; height: 24px; margin: 0; cursor: pointer; accent-color: var(--efinity-blue);" />
      </div>

      {#if autoPatrol}
        <label for="interval">SYNC INTERVAL (e.g. 1h 30m, 1d 2h)</label>
        <input id="interval" type="text" bind:value={intervalString} placeholder="1d" />
      {/if}

      <div style="display: flex; gap: 16px; margin-top: 32px;">
        <button style="flex: 2;" onclick={handleAdd}>ACTIVATE</button>
        <button class="secondary" style="flex: 1;" onclick={() => { show = false; url = ''; name = ''; }}>CANCEL</button>
      </div>
    </div>
  </div>
</div>
