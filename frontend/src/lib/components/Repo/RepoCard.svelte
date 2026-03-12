<script lang="ts">
  import { type Repository } from '$lib/types';
  import { getAvatarUrl, getRemainingTime, getProgress, isSyncing } from '$lib/utils';
  import { API_URL, api } from '$lib/api.svelte';

  let { repo, selectedRepo = $bindable() } = $props<{ repo: Repository, selectedRepo: Repository | null }>();

  function handleAvatarError(e: Event) {
    const img = e.target as HTMLImageElement;
    img.src = "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png";
  }
</script>

<div class="card" onclick={() => selectedRepo = repo}>
  <div style="position: absolute; top: 32px; right: 32px; display: flex; align-items: center; gap: 16px;">
    <div 
      class="radial-timer" 
      class:is-syncing={isSyncing(repo)}
      data-tooltip={getRemainingTime(repo)}
    >
      <svg width="40" height="40">
        <circle cx="20" cy="20" r="16" />
        <circle 
          class="progress" 
          class:active-pulse={repo.status === 'syncing'}
          cx="20" cy="20" r="16" 
          stroke-dasharray="100.5" 
          stroke-dashoffset={100.5 - (repo.progress || 0)} 
        />
      </svg>
    </div>
    <div class="health-score" data-tooltip="Tactical Health Score" style="color: {repo.health_score > 70 ? 'var(--status-green)' : 'var(--status-yellow)'}; border-color: {repo.health_score > 70 ? 'var(--status-green)' : repo.health_score > 40 ? 'var(--status-yellow)' : 'var(--status-red)'}44">
      {repo.health_score}
    </div>
  </div>

  <div style="display: flex; align-items: center; gap: 24px;">
    <img 
      src={getAvatarUrl(repo.url, API_URL)} 
      alt="" 
      style="width: 44px; height: 44px; border-radius: 12px; background: var(--glass); border: 1px solid var(--glass-border);"
      onerror={handleAvatarError}
    />
    <div>
      <h3 style="font-size: 1.5rem; margin: 0; letter-spacing: -0.02em;">{repo.name}</h3>
      <div style="display: flex; align-items: center; gap: 8px; margin-top: 8px;">
        {#if repo.status === 'syncing'}
          <div class="badge" style="color: var(--status-green); background: rgba(0, 255, 136, 0.05); font-size: 0.6rem; padding: 2px 8px;">
            <span class="pulse-dot" style="background: var(--status-green)"></span>
            SYNCING ASSET
          </div>
        {:else if repo.status === 'error'}
          <div class="badge" style="color: var(--status-red); background: rgba(255, 77, 77, 0.05); font-size: 0.6rem; padding: 2px 8px;">
            <span style="width: 6px; height: 6px; background: var(--status-red); border-radius: 50%;"></span>
            LINK FAILURE
          </div>
        {:else}
          <div class="badge" style="color: var(--efinity-blue); background: rgba(0, 112, 243, 0.05); font-size: 0.6rem; padding: 2px 8px;">
            <span style="width: 6px; height: 6px; background: var(--efinity-blue); border-radius: 50%;"></span>
            SECURED
          </div>
        {/if}
        <span class="branch-badge">{repo.default_branch}</span>
      </div>
    </div>
  </div>

  <div class="stats-row">
    <div class="stat-item" data-tooltip="GitHub Stars"><svg width="14" height="14" viewBox="0 0 24 24" fill="var(--status-yellow)" style="opacity: 0.8;"><path d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"/></svg><b>{repo.stars}</b></div>
    <div class="stat-item" data-tooltip="Open Issues"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg><b>{repo.open_issues}</b></div>
  </div>

  <div class="mini-chart">
    {#each JSON.parse(repo.commit_history || '[]') as count}
      <div class="chart-bar" style="height: {count === 0 ? '4px' : Math.min(100, (count / 10) * 100)}%; background: {count === 0 ? 'rgba(255,255,255,0.05)' : `rgba(0, 112, 243, ${0.3 + (Math.min(count, 10) / 10) * 0.7})`}; box-shadow: {count > 5 ? `0 0 12px rgba(0, 112, 243, ${(Math.min(count, 10) / 10) * 0.4})` : 'none'};" data-tooltip="{count} commits"></div>
    {/each}
  </div>

  <div class="info-row">
    <div class="info-item">
      <label>Last Patrol</label>
      <span>{repo.last_sync ? new Date(repo.last_sync).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : 'NEVER'}</span>
    </div>
    <div class="info-item" style="text-align: right;">
      <label>Interval</label>
      <span>{minutesToHuman(repo.interval_minutes)}</span>
    </div>
  </div>
</div>
