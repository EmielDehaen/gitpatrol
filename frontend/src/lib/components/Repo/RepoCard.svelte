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
    <div class="health-score" style="border-color: {repo.health_score > 70 ? 'var(--status-green)' : repo.health_score > 40 ? 'var(--status-yellow)' : 'var(--status-red)'}">
      {repo.health_score}%
    </div>
  </div>

  <div style="display: flex; align-items: center; gap: 24px;">
    <img 
      src={getAvatarUrl(repo.url, API_URL)} 
      alt="" 
      style="width: 64px; height: 64px; border-radius: 16px; border: 1px solid var(--glass-border);"
      onerror={handleAvatarError}
    />
    <div>
      <h3 style="font-size: 1.5rem; margin: 0; letter-spacing: -0.02em;">{repo.name}</h3>
      <div style="display: flex; align-items: center; gap: 8px; margin-top: 8px;">
        {#if repo.status === 'syncing'}
          <div class="badge" style="color: var(--status-green); background: rgba(0, 255, 136, 0.05);">
            <span class="pulse-dot" style="background: var(--status-green)"></span>
            SYNCING ASSET
          </div>
        {:else if repo.status === 'error'}
          <div class="badge" style="color: var(--status-red); background: rgba(255, 77, 77, 0.05);">
            <span style="width: 6px; height: 6px; background: var(--status-red); border-radius: 50%;"></span>
            LINK FAILURE
          </div>
        {:else}
          <div class="badge" style="color: var(--efinity-blue); background: rgba(0, 112, 243, 0.05);">
            <span style="width: 6px; height: 6px; background: var(--efinity-blue); border-radius: 50%;"></span>
            SECURED
          </div>
        {/if}
        <span class="branch-badge">{repo.default_branch}</span>
      </div>
    </div>
  </div>

  <div class="stats-row">
    <div class="stat-item"><b>{repo.stars}</b> stars</div>
    <div class="stat-item"><b>{repo.forks}</b> forks</div>
    <div class="stat-item"><b>{repo.open_issues}</b> issues</div>
  </div>

  <div class="mini-chart" data-tooltip="Last 14 days activity">
    {#each JSON.parse(repo.commit_history || '[]') as count}
      <div class="chart-bar" style="height: {Math.min(100, (count / 10) * 100)}%; opacity: {count > 0 ? 1 : 0.2}"></div>
    {/each}
  </div>

  <div class="info-row">
    <div class="info-item">
      <label>Last Patrol</label>
      <span>{repo.last_sync ? new Date(repo.last_sync).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : 'NEVER'}</span>
    </div>
    <div class="info-item" style="text-align: right;">
      <label>Interval</label>
      <span>{repo.interval_minutes}m</span>
    </div>
  </div>
</div>
