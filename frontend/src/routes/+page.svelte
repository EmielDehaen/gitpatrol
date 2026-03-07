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
    countdown?: string;
  }

  let repositories = $state<Repository[]>([]);
  let showModal = $state(false);
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
      showModal = false;
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

  function getCommitAge(lastCommit: string) {
    if (!lastCommit) return 'Unknown';
    // Format: "hash date"
    const parts = lastCommit.split(' ');
    if (parts.length < 2) return 'Unknown';
    const commitDate = new Date(parts.slice(1).join(' '));
    const diff = Date.now() - commitDate.getTime();
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    if (days === 0) return 'Today';
    if (days < 30) return `${days}d ago`;
    return `${Math.floor(days/30)}mo ago`;
  }

  onMount(() => {
    fetchRepos();
    
    // WebSockets
    const ws = new WebSocket('ws://localhost:8080/ws');
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'status_update') {
        fetchRepos();
      }
    };

    // Countdown timer interval
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
    <h1>GitPatrol ⚡</h1>
    <div style="text-align: right;">
      <div style="font-size: 0.8rem; color: var(--efinity-text-muted); font-weight: 700; text-transform: uppercase;">System Status</div>
      <div style="color: var(--status-green); font-size: 0.9rem; font-weight: 600;">● Online</div>
    </div>
  </header>

  <div class="repo-grid">
    {#each repositories as repo (repo.id)}
      <div class="card">
        <div class="card-header">
          <div>
            <h3 class="repo-name">{repo.name}</h3>
            <div class="repo-url">{repo.url}</div>
          </div>
          <div class="badge" style="color: {repo.status === 'synced' ? 'var(--status-green)' : repo.status === 'error' ? 'var(--status-red)' : '#fff'}">
            <span class="status-dot" style="background: {repo.status === 'synced' ? 'var(--status-green)' : repo.status === 'error' ? 'var(--status-red)' : '#fff'}"></span>
            {repo.status}
          </div>
        </div>

        {#if repo.last_commit}
          <div class="commit-log">
            {repo.last_commit.split(' ')[0]} - {repo.last_commit.split(' ').slice(1, 4).join(' ')}
          </div>
        {/if}

        <div class="info-row">
          <div class="info-item">
            <label>Next Sync</label>
            <span>{repo.countdown || '--:--'}</span>
          </div>
          <div class="info-item">
            <label>Activity</label>
            <span style="color: {getCommitAge(repo.last_commit).includes('mo') ? 'var(--status-red)' : 'var(--efinity-text-main)'}">
              {getCommitAge(repo.last_commit)}
            </span>
          </div>
          <div class="info-item" style="text-align: right;">
            <label>Interval</label>
            <span>{repo.interval_minutes}m</span>
          </div>
        </div>

        {#if repo.error_message}
          <div style="margin-top: 16px; font-size: 0.75rem; color: var(--status-red); padding: 8px; background: rgba(255,0,0,0.1); border-radius: 4px;">
            {repo.error_message}
          </div>
        {/if}
      </div>
    {/each}
  </div>
</div>

<!-- FAB -->
<div class="fab" onclick={() => showModal = true}>
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
</div>

<!-- Modal -->
{#if showModal}
  <div class="modal-overlay" onclick={() => showModal = false}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()} role="presentation">
      <h2 style="margin-top: 0; margin-bottom: 32px;">Add Repository</h2>
      
      <label>Display Name</label>
      <input bind:value={newName} placeholder="e.g. opengem-core" />

      <label>GitHub Repository URL</label>
      <input bind:value={newUrl} placeholder="https://github.com/EmielDehaen/opengem" />

      <label>Sync Interval (Minutes)</label>
      <input bind:value={interval} type="number" />

      <div style="display: flex; gap: 12px; margin-top: 20px;">
        <button style="flex: 1; background: #222;" onclick={() => showModal = false}>Cancel</button>
        <button style="flex: 2;" onclick={addRepo}>Add Patrol</button>
      </div>
    </div>
  </div>
{/if}
