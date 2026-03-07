<script lang="ts">
  import { onMount } from 'svelte';

  interface Repository {
    id: number;
    name: string;
    url: string;
    interval_minutes: number;
    last_sync: string;
    status: string;
    last_commit: string;
    error_message: string;
    stars: number;
    forks: number;
    open_issues: number;
    commit_history: string;
    health_score: number;
    countdown?: string;
  }

  let repositories = $state<Repository[]>([]);
  let selectedRepo = $state<Repository | null>(null);
  let showAddModal = $state(false);
  let newName = $state('');
  let newUrl = $state('');
  let interval = $state(60);

  const API_URL = 'http://localhost:8080';

  async function fetchRepos() {
    const res = await fetch(`${API_URL}/api/repositories`);
    repositories = await res.json();
  }

  async function addRepo() {
    if (!newName || !newUrl) return;
    const res = await fetch(`${API_URL}/api/repositories`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: newName, url: newUrl, interval_minutes: interval })
    });
    if (res.ok) {
      newName = '';
      newUrl = '';
      showAddModal = false;
      fetchRepos();
    }
  }

  function calculateCountdown(repo: Repository) {
    if (!repo.last_sync || repo.status === 'syncing') return '--:--';
    const lastSync = new Date(repo.last_sync).getTime();
    const nextSync = lastSync + repo.interval_minutes * 60 * 1000;
    const diff = nextSync - Date.now();
    if (diff <= 0) return 'Syncing...';
    const mins = Math.floor(diff / 60000);
    const secs = Math.floor((diff % 60000) / 1000);
    return `${mins}m ${secs}s`;
  }

  function parseCommits(lastCommit: string) {
    if (!lastCommit) return [];
    return lastCommit.trim().split('\n').map(line => {
      const [hash, author, date, message] = line.split('|');
      return { hash, author, date, message };
    });
  }

  function getHistoryArray(historyStr: string) {
    try {
      return JSON.parse(historyStr || '[]');
    } catch {
      return [];
    }
  }

  onMount(() => {
    fetchRepos();
    const ws = new WebSocket('ws://localhost:8080/ws');
    ws.onmessage = () => fetchRepos();

    const timer = setInterval(() => {
      repositories = repositories.map(r => ({
        ...r,
        countdown: calculateCountdown(r)
      }));
    }, 1000);

    return () => clearInterval(timer);
  });
</script>

<div class="container">
  <header>
    <div>
      <h1>Patrol Control ⚡</h1>
      <p style="color: var(--efinity-text-muted); margin: 8px 0 0 0; font-weight: 600;">Monitoring {repositories.length} tactical assets</p>
    </div>
    <div style="text-align: right;">
      <div class="badge" style="color: var(--status-green)">
        <span style="width: 8px; height: 8px; background: var(--status-green); border-radius: 50%;"></span>
        Core Active
      </div>
    </div>
  </header>

  <div class="repo-grid">
    {#each repositories as repo (repo.id)}
      <div class="card" onclick={() => selectedRepo = repo}>
        <div class="health-score" style="color: {repo.health_score > 70 ? 'var(--status-green)' : repo.health_score > 40 ? 'var(--status-yellow)' : 'var(--status-red)'}; border-color: {repo.health_score > 70 ? 'var(--status-green)' : repo.health_score > 40 ? 'var(--status-yellow)' : 'var(--status-red)'}44">
          {repo.health_score}
        </div>
        
        <div class="card-header">
          <div>
            <h3 class="repo-name">{repo.name}</h3>
            <div class="repo-url">{repo.url}</div>
          </div>
        </div>

        <div class="stats-row">
          <div class="stat-item"><b>{repo.stars}</b> stars</div>
          <div class="stat-item"><b>{repo.open_issues}</b> issues</div>
        </div>

        <div class="mini-chart">
          {#each getHistoryArray(repo.commit_history) as count}
            <div class="chart-bar" style="height: {Math.min(100, (count / 10) * 100)}%;"></div>
          {/each}
        </div>

        <div class="info-row" style="margin-top: 24px;">
          <div class="info-item">
            <label>Next Sync</label>
            <span>{repo.countdown || '--:--'}</span>
          </div>
          <div class="info-item" style="text-align: right;">
            <label>Status</label>
            <span style="color: {repo.status === 'synced' ? 'var(--status-green)' : '#fff'}">{repo.status}</span>
          </div>
        </div>
      </div>
    {/each}
  </div>
</div>

<!-- Add Repo FAB -->
<div class="fab" onclick={() => showAddModal = true}>
  <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
</div>

<!-- Detail Modal -->
{#if selectedRepo}
  <div class="modal-overlay" onclick={() => selectedRepo = null}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <div style="display: flex; justify-content: space-between; align-items: flex-start;">
          <div>
            <h2 style="margin: 0; font-size: 2rem;">{selectedRepo.name}</h2>
            <p style="color: var(--efinity-text-muted); margin: 8px 0 0 0;">{selectedRepo.url}</p>
          </div>
          <button class="secondary" onclick={() => selectedRepo = null}>Close</button>
        </div>
      </div>
      <div class="modal-body">
        <h4 style="text-transform: uppercase; letter-spacing: 0.1em; color: var(--efinity-text-muted); font-size: 0.8rem;">Recent Activity</h4>
        <div style="margin-top: 20px;">
          {#each parseCommits(selectedRepo.last_commit) as commit}
            <div class="commit-item">
              <div class="commit-hash">{commit.hash.substring(0, 7)}</div>
              <div style="flex: 1;">
                <div style="font-weight: 600;">{commit.message}</div>
                <div style="font-size: 0.8rem; color: var(--efinity-text-muted); margin-top: 4px;">{commit.author} • {commit.date}</div>
              </div>
            </div>
          {/each}
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Add Modal -->
{#if showAddModal}
  <div class="modal-overlay" onclick={() => showAddModal = false}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()} style="max-width: 500px;">
      <div class="modal-header">
        <h2 style="margin: 0;">Add New Patrol</h2>
      </div>
      <div class="modal-body">
        <label>Patrol Name</label>
        <input bind:value={newName} placeholder="e.g. efinity-core" />
        
        <label>Repository URL</label>
        <input bind:value={newUrl} placeholder="https://github.com/..." />
        
        <label>Sync Interval (Minutes)</label>
        <input bind:value={interval} type="number" />
        
        <div style="display: flex; gap: 12px; margin-top: 20px;">
          <button style="flex: 1;" onclick={addRepo}>Activate Patrol</button>
          <button class="secondary" onclick={() => showAddModal = false}>Cancel</button>
        </div>
      </div>
    </div>
  </div>
{/if}
