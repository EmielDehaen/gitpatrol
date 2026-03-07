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
    progress?: number;
  }

  let repositories = $state<Repository[]>([]);
  let viewMode = $state<'grid' | 'list'>('grid');
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
      newName = ''; newUrl = ''; showAddModal = false; fetchRepos();
    }
  }

  async function deleteRepo(id: number) {
    if (!confirm('Are you sure you want to terminate this patrol?')) return;
    const res = await fetch(`${API_URL}/api/repositories/${id}`, { method: 'DELETE' });
    if (res.ok) { selectedRepo = null; fetchRepos(); }
  }

  function getProgress(repo: Repository) {
    if (!repo.last_sync || repo.status === 'syncing') return 0;
    const lastSync = new Date(repo.last_sync).getTime();
    const nextSync = lastSync + repo.interval_minutes * 60000;
    const total = repo.interval_minutes * 60000;
    const remaining = nextSync - Date.now();
    if (remaining <= 0) return 0;
    // Calculation for SVG stroke-dashoffset (Circle radius is 16, circumference is ~100)
    const percentage = remaining / total;
    return 100 * percentage;
  }

  function getRemainingTime(repo: Repository) {
    if (!repo.last_sync || repo.status === 'syncing') return 'Syncing...';
    const lastSync = new Date(repo.last_sync).getTime();
    const nextSync = lastSync + repo.interval_minutes * 60000;
    const remaining = nextSync - Date.now();
    if (remaining <= 0) return 'Syncing...';
    const mins = Math.floor(remaining / 60000);
    const secs = Math.floor((remaining % 60000) / 1000);
    return `Next sync in: ${mins}m ${secs}s`;
  }

  function parseCommits(lastCommit: string) {
    if (!lastCommit) return [];
    return lastCommit.trim().split('\n').map(line => {
      const parts = line.split('|');
      if (parts.length < 4) return null;
      return { 
        hash: parts[0], 
        author: parts[1], 
        date: parts[2], 
        message: parts[3],
        branch: parts[4]?.replace('HEAD -> ', '').split(',')[0].trim() || 'main'
      };
    }).filter(c => c !== null);
  }

  onMount(() => {
    fetchRepos();
    const ws = new WebSocket('ws://localhost:8080/ws');
    ws.onmessage = () => fetchRepos();
    const timer = setInterval(() => {
      repositories = repositories.map(r => ({ ...r, progress: getProgress(r) }));
    }, 1000);
    return () => clearInterval(timer);
  });
</script>

<div class="container">
  <header>
    <div>
      <h1>GitPatrol ⚡</h1>
      <p style="color: var(--efinity-text-muted); margin: 8px 0 0 0; font-weight: 600; text-transform: uppercase; font-size: 0.7rem; letter-spacing: 0.1em;">Tactical Asset Monitoring</p>
    </div>
    
    <div style="display: flex; gap: 24px; align-items: center;">
      <div class="view-toggle">
        <button class:active={viewMode === 'grid'} onclick={() => viewMode = 'grid'}>GRID</button>
        <button class:active={viewMode === 'list'} onclick={() => viewMode = 'list'}>LIST</button>
      </div>
      <div class="badge" style="color: var(--status-green); background: rgba(0, 255, 136, 0.05);">
        <span style="width: 6px; height: 6px; background: var(--status-green); border-radius: 50%;"></span>
        SYSTEM ONLINE
      </div>
    </div>
  </header>

  {#if viewMode === 'grid'}
    <div class="repo-grid">
      {#each repositories as repo (repo.id)}
        <div class="card" onclick={() => selectedRepo = repo}>
          <div style="position: absolute; top: 32px; right: 32px; display: flex; align-items: center; gap: 16px;">
            <div class="radial-timer" title={getRemainingTime(repo)}>
              <svg width="40" height="40">
                <circle cx="20" cy="20" r="16" />
                <circle cx="20" cy="20" r="16" class="progress" class:active-pulse={repo.status !== 'syncing'}
                  style="stroke-dasharray: 100; stroke-dashoffset: {100 - (repo.progress || 0)}" />
              </svg>
            </div>
            <div class="health-score" style="color: {repo.health_score > 70 ? 'var(--status-green)' : 'var(--status-yellow)'}; border-color: {repo.health_score > 70 ? 'var(--status-green)' : repo.health_score > 40 ? 'var(--status-yellow)' : 'var(--status-red)'}44">
              {repo.health_score}
            </div>
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
            {#each JSON.parse(repo.commit_history || '[]') as count}
              <div class="chart-bar" style="height: {Math.max(10, Math.min(100, (count / 10) * 100))}%"></div>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <div class="repo-list">
      {#each repositories as repo (repo.id)}
        <div class="list-item" onclick={() => selectedRepo = repo}>
          <div style="display: flex; align-items: center; gap: 24px;">
            <div class="radial-timer" style="width: 32px; height: 32px;">
              <svg width="32" height="32"><circle cx="16" cy="16" r="12" /><circle cx="16" cy="16" r="12" class="progress" style="stroke-dasharray: 75; stroke-dashoffset: {(repo.progress || 0) * 0.75}" /></svg>
            </div>
            <div>
              <div style="font-weight: 700; font-size: 1.1rem;">{repo.name}</div>
              <div style="font-size: 0.8rem; color: var(--efinity-text-muted);">{repo.url}</div>
            </div>
          </div>
          <div style="display: flex; gap: 40px; align-items: center;">
            <div class="stat-item"><b>{repo.health_score}</b> Health</div>
            <div class="badge" style="color: {repo.status === 'synced' ? 'var(--status-green)' : '#fff'}">{repo.status}</div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<div class="fab" onclick={() => showAddModal = true}>
  <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
</div>

{#if selectedRepo}
  <div class="modal-overlay" onclick={() => selectedRepo = null}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <div>
            <h2 style="margin: 0; font-size: 2.2rem; font-weight: 800;">{selectedRepo.name}</h2>
            <p style="color: var(--efinity-text-muted); margin: 8px 0 0 0; font-family: monospace; font-size: 0.9rem;">{selectedRepo.url}</p>
          </div>
          <button class="secondary" onclick={() => selectedRepo = null}>CLOSE</button>
        </div>
      </div>
      <div class="modal-body">
        <h4 style="text-transform: uppercase; letter-spacing: 0.1em; color: var(--efinity-text-muted); font-size: 0.75rem; font-weight: 800; margin-bottom: 24px;">Recent Mission Logs</h4>
        <div style="display: flex; flex-direction: column; gap: 12px;">
          {#each parseCommits(selectedRepo.last_commit) as commit}
            <div style="background: rgba(255,255,255,0.02); border-radius: 16px; border: 1px solid var(--glass-border); padding: 20px;">
              <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 8px;">
                <span class="branch-badge">{commit.branch}</span>
                <span style="font-weight: 600; font-size: 1rem;">{commit.message}</span>
              </div>
              <div style="font-size: 0.8rem; color: var(--efinity-text-muted); font-weight: 500;">{commit.author} • {commit.date} • <span style="color: var(--efinity-blue); font-family: monospace;">{commit.hash.substring(0,7)}</span></div>
            </div>
          {/each}
        </div>
        <div style="margin-top: 48px; border-top: 1px solid var(--glass-border); padding-top: 32px;">
          <h4 style="color: var(--status-red); text-transform: uppercase; font-size: 0.7rem; letter-spacing: 0.1em; margin-bottom: 16px; font-weight: 800;">Danger Zone</h4>
          <button style="width: 100%; background: rgba(255,77,77,0.05); color: var(--status-red); border: 1px solid rgba(255,77,77,0.15); box-shadow: none;" onclick={() => deleteRepo(selectedRepo!.id)}>TERMINATE PATROL</button>
        </div>
      </div>
    </div>
  </div>
{/if}

{#if showAddModal}
  <div class="modal-overlay" onclick={() => showAddModal = false}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()} style="max-width: 540px;">
      <div class="modal-header">
        <h2 style="margin: 0; font-size: 1.8rem; font-weight: 800;">Deploy New Patrol</h2>
        <p style="color: var(--efinity-text-muted); margin: 8px 0 0 0; font-size: 0.9rem;">Configure a new asset for monitoring.</p>
      </div>
      <div class="modal-body">
        <label>DISPLAY NAME</label>
        <input bind:value={newName} placeholder="e.g. efinity-frontend" />
        <label>REPOSITORY URL</label>
        <input bind:value={newUrl} placeholder="https://github.com/..." />
        <label>SYNC INTERVAL (MINUTES)</label>
        <input type="number" bind:value={interval} />
        <div style="display: flex; gap: 16px; margin-top: 12px;">
          <button style="flex: 2;" onclick={addRepo}>ACTIVATE</button>
          <button class="secondary" style="flex: 1;" onclick={() => showAddModal = false}>CANCEL</button>
        </div>
      </div>
    </div>
  </div>
{/if}
