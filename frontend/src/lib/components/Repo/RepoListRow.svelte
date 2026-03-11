<script lang="ts">
  import { type Repository } from '$lib/types';
  import { getAvatarUrl, getRemainingTime, getProgress, isSyncing } from '$lib/utils';
  import { API_URL } from '$lib/api.svelte';

  let { repo, selectedRepo = $bindable() } = $props<{ repo: Repository, selectedRepo: Repository | null }>();

  function handleAvatarError(e: Event) {
    const img = e.target as HTMLImageElement;
    img.src = "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png";
  }
</script>

<div class="list-item" onclick={() => selectedRepo = repo}>
  <div style="display: flex; align-items: center; gap: 20px;">
    <img 
      src={getAvatarUrl(repo.url, API_URL)} 
      alt="" 
      style="width: 40px; height: 40px; border-radius: 10px; border: 1px solid var(--glass-border);"
      onerror={handleAvatarError}
    />
    <div>
      <div style="font-weight: 800; font-size: 1.1rem; letter-spacing: -0.01em;">{repo.name}</div>
      <div style="font-size: 0.65rem; color: var(--efinity-text-muted); font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; margin-top: 2px;">
        {repo.url.replace('https://', '')}
      </div>
    </div>
  </div>

  <div style="display: flex; align-items: center; gap: 40px;">
    <div style="display: flex; gap: 12px;">
      <div class="stat-item" style="font-size: 0.75rem;"><b>{repo.stars}</b> stars</div>
      <div class="stat-item" style="font-size: 0.75rem;"><b>{repo.forks}</b> forks</div>
    </div>

    <div class="branch-badge">{repo.default_branch}</div>

    <div style="display: flex; align-items: center; gap: 12px; width: 140px; justify-content: flex-end;">
      <span style="font-size: 0.75rem; font-weight: 700; color: {repo.status === 'syncing' ? 'var(--status-green)' : 'var(--efinity-text-muted)'}">
        {repo.status === 'syncing' ? 'SYNCING...' : repo.status === 'error' ? 'FAILURE' : 'SECURED'}
      </span>
      <div class="health-score" style="width: 32px; height: 32px; font-size: 0.7rem; border-color: {repo.health_score > 70 ? 'var(--status-green)' : repo.health_score > 40 ? 'var(--status-yellow)' : 'var(--status-red)'}">
        {repo.health_score}%
      </div>
    </div>
  </div>
</div>
