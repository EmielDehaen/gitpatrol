<script lang="ts">
  import { api } from '$lib/api.svelte';
  import { fade } from 'svelte/transition';

  let { show = $bindable() } = $props();

  async function clearAll() {
    await api.clearIncidents();
    show = false;
  }
</script>

<div class="modal-overlay" onclick={() => show = false} onkeydown={(e) => e.key == 'ESCAPE' && (show = false)} role='presentation' transition:fade>
  <div class="modal-content" onclick={(e) => e.stopPropagation()} role='none'>
    <div class="modal-header">
      <div style="display: flex; justify-content: space-between; align-items: center;">
        <h2 style="font-size: 1.5rem; margin: 0;">Tactical Incident Logs</h2>
        <button class="secondary" onclick={clearAll} style="color: var(--status-red); border-color: rgba(255, 77, 77, 0.2);">CLEAR LOGS</button>
      </div>
    </div>
    <div class="modal-body">
      {#each api.incidents as incident}
        <div class="incident-item">
          <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 12px;">
            <div style="display: flex; align-items: center; gap: 12px;">
              <div style="width: 8px; height: 8px; background: var(--status-red); border-radius: 50%; box-shadow: 0 0 10px var(--status-red);"></div>
              <span style="font-weight: 800; font-size: 0.9rem;">{incident.repo_name}</span>
            </div>
            <span style="font-size: 0.7rem; color: var(--efinity-text-muted); font-weight: 700; font-family: monospace;">{new Date(incident.created_at).toLocaleString()}</span>
          </div>
          <p style="margin: 0; color: #fff; font-size: 0.95rem; line-height: 1.6; font-family: monospace;">{incident.message}</p>
        </div>
      {:else}
        <div style="text-align: center; padding: 80px 40px;">
          <div style="color: var(--status-green); margin-bottom: 24px;">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>
          </div>
          <h3 style="font-size: 1.5rem; margin-bottom: 8px;">No Alerts Detected</h3>
          <p style="color: var(--efinity-text-muted);">All assets are operating within normal parameters.</p>
        </div>
      {/each}
    </div>
  </div>
</div>
