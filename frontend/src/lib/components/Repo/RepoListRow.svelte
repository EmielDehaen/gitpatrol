<script lang="ts">
  import { type Repository } from '$lib/types';
  import { getAvatarUrl, getRemainingTime, getProgress, isSyncing, handleAvatarError } from '$lib/utils';
  import { API_URL } from '$lib/api.svelte';

  let { repo, selectedRepo = $bindable() } = $props<{ repo: Repository, selectedRepo: Repository | null }>();
</script>

<div class="list-item" onclick={() => selectedRepo = repo}>
  <div style="display: flex; align-items: center; gap: 24px;">
    <div class="radial-timer" class:is-syncing={isSyncing(repo)} data-tooltip={getRemainingTime(repo)} style="width: 32px; height: 32px;">
      <svg width="32" height="32">
        <circle cx="16" cy="16" r="12" />
        <circle 
          cx="16" cy="16" r="12" 
          class="progress" 
          class:active-pulse={repo.status !== 'syncing' && repo.auto_patrol === 1} 
          style="stroke-dasharray: 75; stroke-dashoffset: {isSyncing(repo) ? 0 : 75 - (repo.progress || 0) * 0.75}" 
        />
      </svg>
    </div>
    <img 
      src={getAvatarUrl(repo.url)} 
      alt="" 
      style="width: 32px; height: 32px; border-radius: 8px; border: 1px solid var(--glass-border);"
      onerror={handleAvatarError}
    />
    <div>
      <div style="display: flex; align-items: center; gap: 12px;">
        <div style="font-weight: 700; font-size: 1.1rem;">{repo.name}</div>
        {#if repo.status === 'synced'}
          <span class="badge" style="color: var(--status-green); background: rgba(0, 255, 136, 0.05); font-size: 0.6rem; padding: 2px 8px;">SYNCED</span>
        {:else if repo.status === 'error'}
          <span class="badge" style="color: var(--status-red); background: rgba(255, 77, 77, 0.05); font-size: 0.6rem; padding: 2px 8px;">ERROR</span>
        {/if}
      </div>
      <div style="font-size: 0.8rem; color: var(--efinity-text-muted);">{repo.url.replace('https://github.com/', '')}</div>
    </div>
  </div>

  <div style="display: flex; gap: 40px; align-items: center;">
    <div class="stat-item">
      <svg width="12" height="12" viewBox="0 0 24 24" fill="var(--status-yellow)">
        <path d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"/>
      </svg>
      <b>{repo.stars}</b>
    </div>
    <div class="stat-item"><b>{repo.health_score}</b> Health</div>
    <div class="badge" style="color: {repo.status === 'synced' ? 'var(--status-green)' : '#fff'}">{repo.status.toUpperCase()}</div>
  </div>
</div>
