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
    if (!res.ok) return;
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

  async function deleteRepo(id: number) {
    if (!confirm('Are you sure you want to terminate this patrol? All local data will be deleted.')) return;
    const res = await fetch(`${API_URL}/api/repositories/${id}`, {
      method: 'DELETE'
    });
    if (res.ok) {
      selectedRepo = null;
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
      const parts = line.split('|');
      if (parts.length < 4) return null;
      const [hash, author, date, message] = parts;
      return { hash, author, date, message };
    }).filter(c => c !== null);
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
      <h1>GitPatrol ⚡</h1>
      <p style="color: var(--efinity-text-muted); margin: 8px 0 0 0; font-weight: 600; letter-spacing: 0.05em; text-transform: uppercase; font-size: 0.75rem;">Tactical Asset Monitoring</p>
    </div>
    <div style="text-align: right;">
      <div class="badge" style="color: var(--status-green); background: rgba(0, 255, 136, 0.05); border: 1px solid rgba(0, 255, 136, 0.1);">
        <span style="width: 6px; height: 6px; background: var(--status-green); border-radius: 50%; box-shadow: 0 0 8px var(--status-green);"></span>
        Active Patrols: {repositories.length}
      </div>
    </div>
  </header>

  <div class="repo-grid">
    {#each repositories as repo (repo.id)}
      <div class="card" onclick={() => selectedRepo = repo}>
        <div class="health-score" style="color: {repo.health_score > 70 ? 'var(--status-green)' : repo.health_score > 40 ? 'var(--status-yellow)' : 'var(--status-red)'}; border-color: {repo.health_score > 70 ? 'var(--status-green)' : repo.health_score > 40 ? 'var(--status-yellow)' : 'var(--status-red)'}33">
          {repo.health_score}
        </div>
        
        <div class="card-header">
          <div>
            <h3 class="repo-name">{repo.name}</h3>
            <div class="repo-url">{repo.url.replace('https://github.com/', '')}</div>
          </div>
        </div>

        <div class="stats-row">
          <div class="stat-item"><b>{repo.stars}</b> stars</div>
          <div class="stat-item"><b>{repo.open_issues}</b> issues</div>
        </div>

        <div class="mini-chart">
          {#each getHistoryArray(repo.commit_history) as count}
            <div class="chart-bar" style="height: {Math.max(10, Math.min(100, (count / 10) * 100))}%; opacity: {count > 0 ? 0.8 : 0.2}"></div>
          {/each}
        </div>

        <div class="info-row">
          <div class="info-item">
            <label>Next Sync</label>
            <span>{repo.countdown || '--:--'}</span>
          </div>
          <div class="info-item" style="text-align: right;">
            <label>Status</label>
            <span style="color: {repo.status === 'synced' ? 'var(--status-green)' : repo.status === 'error' ? 'var(--status-red)' : '#fff'}">{repo.status}</span>
          </div>
        </div>
      </div>
    {/each}
  </div>
</div>

<!-- Add Repo FAB -->
<div class="fab" onclick={() => showAddModal = true}>
  <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
</div>

<!-- Detail Modal -->
{#if selectedRepo}
  <div class="modal-overlay" onclick={() => selectedRepo = null}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <div style="display: flex; justify-content: space-between; align-items: flex-start;">
          <div>
            <h2 style="margin: 0; font-size: 2.2rem; font-weight: 800;">{selectedRepo.name}</h2>
            <p style="color: var(--efinity-text-muted); margin: 8px 0 0 0; font-family: monospace;">{selectedRepo.url}</p>
          </div>
          <button class="secondary" onclick={() => selectedRepo = null}>CLOSE</button>
        </div>
      </div>
      <div class="modal-body">
        <h4 style="text-transform: uppercase; letter-spacing: 0.1em; color: var(--efinity-text-muted); font-size: 0.75rem; font-weight: 800; margin-bottom: 24px;">Recent Mission Logs</h4>
        <div style="display: flex; flex-direction: column; gap: 8px;">
          {#each parseCommits(selectedRepo.last_commit) as commit}
            <div class="commit-item" style="background: rgba(255,255,255,0.02); border-radius: 12px; border: 1px solid var(--glass-border);">
              <div class="commit-hash" style="background: rgba(0, 112, 243, 0.1); padding: 4px 8px; border-radius: 6px; height: fit-content;">{commit.hash.substring(0, 7)}</div>
              <div style="flex: 1;">
                <div style="font-weight: 600; font-size: 0.95rem;">{commit.message}</div>
                <div style="font-size: 0.75rem; color: var(--efinity-text-muted); margin-top: 6px; font-weight: 500;">{commit.author} • {commit.date}</div>
              </div>
            </div>
          {/each}
        </div>

        <div style="margin-top: 48px; padding-top: 32px; border-top: 1px solid rgba(255,0,0,0.1);">
          <h4 style="color: var(--status-red); text-transform: uppercase; font-size: 0.7rem; letter-spacing: 0.1em; margin-bottom: 16px;">Danger Zone</h4>
          <button style="background: rgba(255, 77, 77, 0.1); color: var(--status-red); box-shadow: none; border: 1px solid rgba(255, 77, 77, 0.2); width: 100%;" onclick={() => deleteRepo(selectedRepo!.id)}>
            TERMINATE PATROL
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Add Modal -->
{#if showAddModal}
  <div class="modal-overlay" onclick={() => showAddModal = false}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()} style="max-width: 540px; padding: 0;">
      <div class="modal-header" style="background: linear-gradient(to bottom, #111, #0a0a0a); padding: 40px;">
        <h2 style="margin: 0; font-size: 1.8rem; font-weight: 800;">Deploy New Patrol</h2>
        <p style="color: var(--efinity-text-muted); margin: 8px 0 0 0; font-size: 0.9rem;">Configure a new repository for automated monitoring.</p>
      </div>
      <div class="modal-body" style="padding: 40px;">
        <div style="margin-bottom: 24px;">
          <label>DISPLAY NAME</label>
          <input bind:value={newName} placeholder="e.g. efinity-frontend" />
        </div>
        
        <div style="margin-bottom: 24px;">
          <label>REPOSITORY URL</label>
          <input bind:value={newUrl} placeholder="https://github.com/..." />
        </div>
        
        <div style="margin-bottom: 32px;">
          <label>SYNC INTERVAL (MINUTES)</label>
          <input bind:value={interval} type="number" />
        </div>
        
        <div style="display: flex; gap: 16px;">
          <button style="flex: 2;" onclick={addRepo}>ACTIVATE PATROL</button>
          <button class="secondary" style="flex: 1;" onclick={() => showAddModal = false}>CANCEL</button>
        </div>
      </div>
    </div>
  </div>
{/if}
