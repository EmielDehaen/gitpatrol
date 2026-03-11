<script lang="ts">
  import { onMount } from 'svelte';
  import { marked } from 'marked';
  import { fade } from 'svelte/transition';

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
    default_branch: string;
    auto_patrol: number;
    progress?: number;
  }

  let repositories = $state<Repository[]>([]);
  let viewMode = $state<'grid' | 'list'>('grid');
  let selectedRepo = $state<Repository | null>(null);
  let activeTab = $state<'readme' | 'issues' | 'releases' | 'wiki' | 'logs'>('readme');
  let showConfigModal = $state(false);
  let readmeContent = $state('');
  let wikiContent = $state('');
  let issues = $state<any[]>([]);
  let releases = $state<any[]>([]);
  let readmeExpanded = $state(false);
  let showAddModal = $state(false);
  let newUrl = $state('');
  let newName = $state('');
  let intervalString = $state('1h');
  let autoPatrol = $state(true);
  let incidents = $state<any[]>([]);
  let showIncidentModal = $state(false);

  // Auth State
  let isAuthenticated = $state(false);
  let needsBootstrap = $state(false);
  let authUsername = $state('');
  let authPassword = $state('');
  let authLoading = $state(true);

  // User Settings State
  let showUserModal = $state(false);
  let editUsername = $state('');
  let oldPassword = $state('');
  let newPassword = $state('');
  let confirmPassword = $state('');

  interface Toast { id: number; message: string; type: 'success' | 'error' | 'info'; }
  let toasts = $state<Toast[]>([]);
  let toastId = 0;

  function showToast(message: string, type: 'success' | 'error' | 'info' = 'info') {
    const id = toastId++;
    toasts = [...toasts, { id, message, type }];
    setTimeout(() => {
      toasts = toasts.filter(t => t.id !== id);
    }, 5000);
  }

  function getNormalizedUrl(url: string) {
    if (!url || url.length < 3) return '';
    let u = url.trim().replace(/\.git$/, '');
    if (u.startsWith('git@')) {
      u = u.replace(':', '/').replace('git@', 'https://');
    }
    if (u.includes('/tree/')) u = u.split('/tree/')[0];
    if (u.includes('/blob/')) u = u.split('/blob/')[0];
    if (!u.startsWith('http') && u.includes('/')) {
      const parts = u.split('/');
      if (parts.length === 2) u = 'https://github.com/' + u;
    }
    return u;
  }

  function suggestName(url: string) {
    if (!url) return '';
    let cleanUrl = getNormalizedUrl(url);
    const parts = cleanUrl.split('/');
    let name = parts[parts.length - 1];
    return name ? name.charAt(0).toUpperCase() + name.slice(1) : '';
  }

  $effect(() => {
    if (newUrl && !newName) {
      newName = suggestName(newUrl);
    }
  });

  const API_URL = 'http://localhost:8080';

  async function checkAuth() {
    try {
      const res = await fetch(`${API_URL}/api/auth/status`, { credentials: 'include' });
      const data = await res.json();
      needsBootstrap = data.needs_bootstrap;
      isAuthenticated = data.logged_in;
      if (isAuthenticated) {
        const meRes = await fetch(`${API_URL}/api/me`, { credentials: 'include' });
        if (meRes.ok) {
          const meData = await meRes.json();
          authUsername = meData.username;
          editUsername = meData.username;
        }
        fetchRepos();
        fetchIncidents();
      }
    } catch (e) {
      console.error('Auth check failed', e);
    } finally {
      authLoading = false;
    }
  }

  async function handleAuth() {
    const endpoint = needsBootstrap ? '/api/auth/register' : '/api/auth/login';
    const res = await fetch(`${API_URL}${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: authUsername, password: authPassword }),
      credentials: 'include'
    });
    
    if (res.ok) {
      authUsername = ''; authPassword = '';
      showToast(needsBootstrap ? 'System bootstrapped!' : 'Welcome back.', 'success');
      await checkAuth();
    } else {
      const data = await res.json();
      showToast(data.error || 'Authentication failed.', 'error');
    }
  }

  async function handleLogout() {
    await fetch(`${API_URL}/api/auth/logout`, { method: 'POST', credentials: 'include' });
    isAuthenticated = false;
    repositories = [];
    showToast('Logged out.', 'info');
  }

  async function updateProfile() {
    if (newPassword && newPassword !== confirmPassword) {
      showToast('Passwords do not match.', 'error');
      return;
    }
    const res = await fetch(`${API_URL}/api/user`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        username: editUsername, 
        old_password: oldPassword, 
        new_password: newPassword 
      }),
      credentials: 'include'
    });
    if (res.ok) {
      showToast('Profile updated.', 'success');
      showUserModal = false;
      oldPassword = ''; newPassword = ''; confirmPassword = '';
      checkAuth();
    } else {
      const data = await res.json();
      showToast(data.error || 'Update failed.', 'error');
    }
  }

  function minutesToHuman(minutes: number): string {
    if (minutes <= 0) return '0m';
    const d = Math.floor(minutes / 1440);
    const h = Math.floor((minutes % 1440) / 60);
    const m = minutes % 60;
    let parts = [];
    if (d > 0) parts.push(`${d}d`);
    if (h > 0) parts.push(`${h}h`);
    if (m > 0) parts.push(`${m}m`);
    return parts.join(' ') || '0m';
  }

  function humanToMinutes(str: string): number {
    const regex = /(?:(\d+)d)?\s*(?:(\d+)h)?\s*(?:(\d+)m)?/i;
    const match = str.match(regex);
    if (!match) return 60;
    const d = parseInt(match[1] || '0');
    const h = parseInt(match[2] || '0');
    const m = parseInt(match[3] || '0');
    const total = (d * 1440) + (h * 60) + m;
    return total > 0 ? total : 60;
  }

  async function fetchRepos() {
    const res = await fetch(`${API_URL}/api/repositories`, { credentials: 'include' });
    if (!res.ok) {
      if (res.status === 401) isAuthenticated = false;
      return;
    }
    repositories = await res.json();
  }

  async function fetchIncidents() {
    const res = await fetch(`${API_URL}/api/incidents`, { credentials: 'include' });
    if (!res.ok) return;
    incidents = await res.json();
  }

  async function clearIncidents() {
    const res = await fetch(`${API_URL}/api/incidents`, { method: 'DELETE', credentials: 'include' });
    if (res.ok) {
      incidents = [];
      showIncidentModal = false;
      showToast('All incidents cleared.', 'info');
    }
  }

  async function fetchMetadata(repo: Repository) {
    const assetBase = `${API_URL}/api/repositories/${repo.id}/assets/`;
    const issuesRes = await fetch(`${assetBase}metadata/issues.json`, { credentials: 'include' });
    issues = issuesRes.ok ? await issuesRes.json() : [];
    const releasesRes = await fetch(`${assetBase}metadata/releases.json`, { credentials: 'include' });
    releases = releasesRes.ok ? await releasesRes.json() : [];
    const wikiRes = await fetch(`${assetBase}wiki/Home.md`, { credentials: 'include' });
    if (wikiRes.ok) {
      wikiContent = await marked.parse(await wikiRes.text());
    } else {
      wikiContent = '<p style="color: var(--efinity-text-muted)">No documentation (Wiki) found for this asset.</p>';
    }
  }

  async function fetchReadme(repo: Repository) {
    readmeContent = 'Loading mission briefing...';
    readmeExpanded = false;
    activeTab = 'readme';
    const res = await fetch(`${API_URL}/api/repositories/${repo.id}/readme`, { credentials: 'include' });
    if (res.ok) {
      let text = await res.text();
      const assetBase = `${API_URL}/api/repositories/${repo.id}/assets/`;
      text = text.replace(/!\[([^\]]*)\]\((?!(?:http|https|ftp|data:))(?:\.\/)?([^)]+)\)/gi, `![$1](${assetBase}$2)`);
      text = text.replace(/<img([^>]+)src=["'](?!(?:http|https|ftp))(?:\.\/)?([^"']+)["']/gi, (match, pre, path) => `<img${pre}src="${assetBase}${path}"`);
      text = text.replace(/\[([^\]]*)\]\((?!(?:http|https|ftp|#))(?:\.\/)?([^)]+)\)/gi, `[$1](${assetBase}$2)`);
      readmeContent = await marked.parse(text, { gfm: true, breaks: true });
    } else {
      readmeContent = '<p style="color: var(--efinity-text-muted)">No mission briefing available for this asset.</p>';
    }
    fetchMetadata(repo);
  }

  $effect(() => { if (selectedRepo) fetchReadme(selectedRepo); });

  async function addRepo() {
    if (!newName || !newUrl) return;
    const res = await fetch(`${API_URL}/api/repositories`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: newName, url: newUrl, interval_minutes: humanToMinutes(intervalString), auto_patrol: autoPatrol ? 1 : 0 }),
      credentials: 'include'
    });
    if (res.ok) {
      newName = ''; newUrl = ''; autoPatrol = true; intervalString = '1h'; showAddModal = false; fetchRepos();
      showToast('Patrol deployed successfully!', 'success');
    } else {
      const err = await res.json();
      showToast(err.error || 'Failed to deploy patrol.', 'error');
    }
  }

  async function deleteRepo(id: number) {
    if (!confirm('Are you sure you want to terminate this patrol?')) return;
    const res = await fetch(`${API_URL}/api/repositories/${id}`, { method: 'DELETE', credentials: 'include' });
    if (res.ok) { selectedRepo = null; showConfigModal = false; fetchRepos(); showToast('Patrol terminated.', 'info'); }
  }

  async function updateConfig() {
    if (!selectedRepo) return;
    const res = await fetch(`${API_URL}/api/repositories/${selectedRepo.id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ interval_minutes: humanToMinutes(intervalString), auto_patrol: selectedRepo.auto_patrol }),
      credentials: 'include'
    });
    if (res.ok) { fetchRepos(); selectedRepo = repositories.find(r => r.id === selectedRepo?.id) || null; showConfigModal = false; showToast('Configuration updated.', 'success'); }
  }

  async function syncRepoNow(repo: Repository) {
    await fetch(`${API_URL}/api/repositories/${repo.id}/sync`, { method: 'POST', credentials: 'include' });
  }

  function openConfig() { if (selectedRepo) { intervalString = minutesToHuman(selectedRepo.interval_minutes); showConfigModal = true; } }

  function isSyncing(repo: Repository) {
    if (repo.status === 'syncing') return true;
    if (repo.auto_patrol === 0 || !repo.last_sync) return false;
    const lastSync = new Date(repo.last_sync).getTime();
    const nextSync = lastSync + repo.interval_minutes * 60000;
    return (nextSync - Date.now()) <= 0;
  }

  function getProgress(repo: Repository) {
    if (!repo.last_sync || repo.status === 'syncing' || repo.auto_patrol === 0) return 0;
    const lastSync = new Date(repo.last_sync).getTime();
    const nextSync = lastSync + repo.interval_minutes * 60000;
    const total = repo.interval_minutes * 60000;
    const remaining = nextSync - Date.now();
    return remaining <= 0 ? 0 : 100 * (remaining / total);
  }

  function getRemainingTime(repo: Repository) {
    if (repo.auto_patrol === 0) return 'Manual Patrol Only';
    if (repo.status === 'syncing' || isSyncing(repo)) return 'Syncing...';
    if (!repo.last_sync) return 'Pending first patrol';
    const lastSync = new Date(repo.last_sync).getTime();
    const nextSync = lastSync + repo.interval_minutes * 60000;
    const remaining = nextSync - Date.now();
    if (remaining <= 0) return 'Syncing...';
    const minutes = Math.floor(remaining / 60000);
    const seconds = Math.floor((remaining % 60000) / 1000);
    return `Next patrol in ${minutes}m ${seconds}s`;
  }

  function parseCommits(data: string) {
    if (!data) return [];
    return data.split('\n').map(line => {
      const parts = line.split('|');
      if (parts.length < 4) return null;
      const refs = parts[4] || '';
      let branch = '';
      if (refs) {
        const cleanRefs = refs.replace(/[()]/g, '').split(', ');
        branch = cleanRefs.find(r => !r.includes('HEAD') && !r.includes('tag:')) || '';
      }
      return { hash: parts[0], author: parts[1], date: parts[2], message: parts[3], branch };
    }).filter(c => c !== null);
  }

  function parseHistory(history: string) {
    try {
      if (!history || history === '[]' || history === 'null') return Array(14).fill(0);
      const parsed = JSON.parse(history);
      return Array.isArray(parsed) ? parsed : Array(14).fill(0);
    } catch (e) { return Array(14).fill(0); }
  }

  function handleAvatarError(e: Event) {
    const img = e.target as HTMLImageElement;
    img.src = `https://ui-avatars.com/api/?name=GP&background=0070f3&color=fff`;
  }

  function getAvatarUrl(url: string) {
    const parts = url.replace('https://github.com/', '').replace('https://gitlab.com/', '').split('/');
    if (parts.length > 0) return `${API_URL}/avatars/${parts[0]}.png`;
    return '';
  }

  onMount(() => {
    checkAuth();
    const ws = new WebSocket('ws://localhost:8080/ws');
    ws.onmessage = () => { if (isAuthenticated) { fetchRepos(); fetchIncidents(); } };
    const timer = setInterval(() => { if (isAuthenticated) { repositories = repositories.map(r => ({ ...r, progress: getProgress(r) })); } }, 1000);
    return () => clearInterval(timer);
  });
</script>

{#if authLoading}
  <div style="height: 100vh; display: flex; align-items: center; justify-content: center; background: var(--efinity-dark);">
    <div class="badge" style="color: var(--efinity-blue); background: rgba(0, 112, 243, 0.05); padding: 12px 24px;">ESTABLISHING NEURAL LINK...</div>
  </div>
{:else if !isAuthenticated}
  <div style="height: 100vh; display: flex; align-items: center; justify-content: center; background: var(--efinity-dark);">
    <div class="modal-content" style="max-width: 400px; padding: 48px; border: 1px solid var(--glass-border);" transition:fade>
      <div style="text-align: center; margin-bottom: 32px;">
        <h1 style="font-size: 2.5rem; margin: 0;">GitPatrol ⚡</h1>
        <p style="color: var(--efinity-text-muted); margin-top: 8px;">{needsBootstrap ? 'System Initial Setup' : 'Tactical Access Required'}</p>
      </div>
      <label>USERNAME</label>
      <input bind:value={authUsername} placeholder="e.g. emiel" />
      <label>PASSWORD</label>
      <input type="password" bind:value={authPassword} placeholder="••••••••" onkeydown={(e) => e.key === 'Enter' && handleAuth()} />
      <button style="width: 100%; margin-top: 32px;" onclick={handleAuth}>{needsBootstrap ? 'BOOTSTRAP SYSTEM' : 'ACCESS DASHBOARD'}</button>
      {#if !needsBootstrap}<p style="text-align: center; font-size: 0.75rem; color: var(--efinity-text-muted); margin-top: 24px;">Contact your administrator for access.</p>{/if}
    </div>
  </div>
{:else}
<div class="container" transition:fade>
  <header>
    <div>
      <h1>GitPatrol ⚡</h1>
      <p style="color: var(--efinity-text-muted); margin: 8px 0 0 0; font-weight: 600; text-transform: uppercase; font-size: 0.7rem; letter-spacing: 0.1em;">Tactical Asset Monitoring</p>
    </div>
    <div style="display: flex; gap: 24px; align-items: center;">
      <div class="user-profile" onclick={() => showUserModal = true} data-tooltip="User Settings">
        <div class="avatar-circle">{authUsername ? authUsername.charAt(0) : 'U'}</div>
        <div class="user-info">
          <span class="user-name">{authUsername}</span>
          <span class="user-role">Administrator</span>
        </div>
      </div>
      <div class="notification-bell" class:has-incidents={incidents.length > 0} onclick={() => showIncidentModal = true} data-tooltip="Security Logs">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
        {#if incidents.length > 0}<div class="bell-count">{incidents.length}</div>{/if}
      </div>
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

  <main>
    <div style="margin-bottom: 48px;">
      <h2 style="margin: 0; font-size: 2.2rem; font-weight: 800;">Insured Assets</h2>
      <p style="color: var(--efinity-text-muted); margin: 8px 0 0 0;">Monitoring {repositories.length} active sectors.</p>
    </div>

  {#if viewMode === 'grid'}
    <div class="repo-grid">
      {#each repositories as repo (repo.id)}
        <div class="card" onclick={() => selectedRepo = repo}>
          <div style="position: absolute; top: 32px; right: 32px; display: flex; align-items: center; gap: 16px;">
            <div class="radial-timer" class:is-syncing={isSyncing(repo)} data-tooltip={getRemainingTime(repo)}>
              <svg width="40" height="40">
                <circle cx="20" cy="20" r="16" />
                <circle cx="20" cy="20" r="16" class="progress" class:active-pulse={repo.status !== 'syncing' && repo.auto_patrol === 1}
                  style="stroke-dasharray: 100; stroke-dashoffset: {isSyncing(repo) ? 0 : 100 - (repo.progress || 0)}" />
              </svg>
            </div>
            <div class="health-score" data-tooltip="Tactical Health Score" style="color: {repo.health_score > 70 ? 'var(--status-green)' : 'var(--status-yellow)'}; border-color: {repo.health_score > 70 ? 'var(--status-green)' : repo.health_score > 40 ? 'var(--status-yellow)' : 'var(--status-red)'}44">
              {repo.health_score}
            </div>
          </div>
          <div style="display: flex; align-items: center; gap: 20px; margin-bottom: 32px;">
            <img src={getAvatarUrl(repo.url)} onerror={handleAvatarError} alt="" style="width: 48px; height: 48px; border-radius: 12px; border: 1px solid var(--glass-border);" />
            <div>
              <div style="display: flex; align-items: center; gap: 10px;">
                <div style="font-weight: 800; font-size: 1.2rem; color: #fff;">{repo.name}</div>
                {#if repo.status === 'synced'}<span class="badge" style="color: var(--status-green); background: rgba(0, 255, 136, 0.05); font-size: 0.6rem; padding: 2px 8px;">SYNCED</span>{:else if repo.status === 'error'}<span class="badge" style="color: var(--status-red); background: rgba(255, 77, 77, 0.05); font-size: 0.6rem; padding: 2px 8px;">ERROR</span>{/if}
              </div>
              <div class="repo-url">{repo.url.replace('https://github.com/', '')}</div>
            </div>
          </div>
          <div class="history-container">
            {#each parseHistory(repo.commit_history) as count}
              <div class="history-bar" style="height: {Math.max(5, Math.min(100, count * 20))}%"></div>
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
            <div class="radial-timer" class:is-syncing={isSyncing(repo)} data-tooltip={getRemainingTime(repo)} style="width: 32px; height: 32px;">
              <svg width="32" height="32"><circle cx="16" cy="16" r="12" /><circle cx="16" cy="16" r="12" class="progress" class:active-pulse={repo.status !== 'syncing' && repo.auto_patrol === 1} style="stroke-dasharray: 75; stroke-dashoffset: {isSyncing(repo) ? 0 : 75 - (repo.progress || 0) * 0.75}" /></svg>
            </div>
            <img src={getAvatarUrl(repo.url)} onerror={handleAvatarError} alt="" style="width: 32px; height: 32px; border-radius: 8px; border: 1px solid var(--glass-border);" />
            <div>
              <div style="display: flex; align-items: center; gap: 12px;">
                <div style="font-weight: 700; font-size: 1.1rem;">{repo.name}</div>
                {#if repo.status === 'synced'}<span class="badge" style="color: var(--status-green); background: rgba(0, 255, 136, 0.05); font-size: 0.6rem; padding: 2px 8px;">SYNCED</span>{/if}
                {#if repo.status === 'error'}<span class="badge" style="color: var(--status-red); background: rgba(255, 77, 77, 0.05); font-size: 0.6rem; padding: 2px 8px;">ERROR</span>{/if}
              </div>
              <div style="font-size: 0.8rem; color: var(--efinity-text-muted);">{repo.url.replace('https://github.com/', '')}</div>
            </div>
          </div>
          <div style="display: flex; gap: 40px; align-items: center;">
            <div style="text-align: right;"><div style="font-size: 0.65rem; color: var(--efinity-text-muted); text-transform: uppercase; font-weight: 800; letter-spacing: 0.05em; margin-bottom: 4px;">Health</div><div style="font-weight: 800; color: {repo.health_score > 70 ? 'var(--status-green)' : 'var(--status-yellow)'};">{repo.health_score}%</div></div>
            <div style="text-align: right;"><div style="font-size: 0.65rem; color: var(--efinity-text-muted); text-transform: uppercase; font-weight: 800; letter-spacing: 0.05em; margin-bottom: 4px;">Last Sync</div><div style="font-weight: 800;">{repo.last_sync ? new Date(repo.last_sync).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) : 'Never'}</div></div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
  </main>
  <button class="fab" onclick={() => showAddModal = true} data-tooltip="Deploy New Patrol"><svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg></button>
</div>
{/if}

{#if showUserModal}
  <div class="modal-overlay" onclick={() => showUserModal = false}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()} style="max-width: 450px;">
      <div class="modal-header"><h2 style="margin: 0; font-size: 1.8rem; font-weight: 800;">User Settings</h2><p style="color: var(--efinity-text-muted); margin: 4px 0 0 0; font-size: 0.8rem;">Update your tactical profile.</p></div>
      <div class="modal-body">
        <label>USERNAME</label>
        <input bind:value={editUsername} />
        <div style="margin: 32px 0; height: 1px; background: var(--glass-border);"></div>
        <h4 style="margin-top: 0; font-size: 0.75rem; color: var(--efinity-blue);">CHANGE PASSWORD</h4>
        <label>CURRENT PASSWORD</label>
        <input type="password" bind:value={oldPassword} placeholder="Required for changes" />
        <label>NEW PASSWORD</label>
        <input type="password" bind:value={newPassword} placeholder="Leave empty to keep current" />
        <label>CONFIRM NEW PASSWORD</label>
        <input type="password" bind:value={confirmPassword} placeholder="Repeat new password" />
        <div style="display: flex; gap: 16px; margin-top: 32px;">
          <button style="flex: 2;" onclick={updateProfile}>UPDATE PROFILE</button>
          <button class="secondary" style="flex: 1;" onclick={handleLogout}>LOGOUT</button>
        </div>
      </div>
    </div>
  </div>
{/if}

{#if selectedRepo}
  <div class="modal-overlay" onclick={() => { selectedRepo = null; showConfigModal = false; }}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 32px;">
          <div style="display: flex; align-items: center; gap: 24px;">
            <img src={getAvatarUrl(selectedRepo.url)} onerror={handleAvatarError} alt="" style="width: 64px; height: 64px; border-radius: 16px; border: 1px solid var(--glass-border);" />
            <div>
              <div style="display: flex; align-items: center; gap: 16px;">
                <h2 style="margin: 0; font-size: 2.2rem; font-weight: 800;">{selectedRepo.name}</h2>
                <div class="radial-timer" class:is-syncing={isSyncing(selectedRepo)} data-tooltip={getRemainingTime(selectedRepo)} style="width: 32px; height: 32px;">
                  <svg width="32" height="32"><circle cx="16" cy="16" r="12" /><circle cx="16" cy="16" r="12" class="progress" class:active-pulse={selectedRepo.status !== 'syncing' && selectedRepo.auto_patrol === 1} style="stroke-dasharray: 75; stroke-dashoffset: {isSyncing(selectedRepo) ? 0 : 75 - (selectedRepo.progress || 0) * 0.75}" /></svg>
                </div>
                <div class="badge" style="color: {selectedRepo.health_score > 70 ? 'var(--status-green)' : 'var(--status-yellow)'}; background: rgba(255,255,255,0.03); font-size: 0.8rem; padding: 4px 12px; border: 1px solid rgba(255,255,255,0.05);">{selectedRepo.health_score}% HEALTH</div>
              </div>
              <a href={selectedRepo.url} target="_blank" rel="noopener noreferrer" style="color: var(--efinity-blue); text-decoration: none; font-family: monospace; font-size: 0.95rem; display: block; margin-top: 8px;">{selectedRepo.url} ↗</a>
              {#if selectedRepo.status === 'error'}<div style="margin-top: 16px; padding: 12px 16px; background: rgba(255, 77, 77, 0.05); border: 1px solid rgba(255, 77, 77, 0.1); border-radius: 12px; color: var(--status-red); font-size: 0.85rem; font-weight: 600; display: flex; align-items: center; gap: 10px;"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>{selectedRepo.error_message}</div>{/if}
            </div>
          </div>
          <div style="display: flex; gap: 12px; align-items: center;">
            {#if selectedRepo.auto_patrol === 0 || selectedRepo.status !== 'syncing'}<button class="secondary" style="padding: 10px; border-radius: 12px; color: var(--status-green); border-color: rgba(0, 255, 136, 0.2);" onclick={() => syncRepoNow(selectedRepo!)} data-tooltip="Sync Now"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21.5 2v6h-6M2.5 22v-6h6M2 12c0-4.4 3.6-8 8-8 3.3 0 6.1 2 7.3 4.9M22 12c0 4.4-3.6 8-8 8-3.3 0-6.1-2-7.3-4.9"/></svg></button>{/if}
            <button class="secondary" style="padding: 10px; border-radius: 12px;" onclick={openConfig} data-tooltip="Settings"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg></button>
            <button class="secondary" style="padding: 10px; border-radius: 12px; color: var(--status-red); border-color: rgba(255, 77, 77, 0.2);" onclick={() => deleteRepo(selectedRepo!.id)} data-tooltip="Delete Patrol"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"/><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg></button>
            <button class="secondary" style="padding: 10px; border-radius: 12px; margin-left: 12px;" onclick={() => { selectedRepo = null; showConfigModal = false; }} data-tooltip="Close"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg></button>
          </div>
        </div>
      </div>
      <div class="modal-body">
        <div class="tab-container">
          <button class="tab-btn" class:active={activeTab === 'readme'} onclick={() => activeTab = 'readme'}>BRIEFING</button>
          <button class="tab-btn" class:active={activeTab === 'issues'} onclick={() => activeTab = 'issues'}>INTEL <span class="tab-count">{issues.length}</span></button>
          <button class="tab-btn" class:active={activeTab === 'releases'} onclick={() => activeTab = 'releases'}>CHRONICLE <span class="tab-count">{releases.length}</span></button>
          <button class="tab-btn" class:active={activeTab === 'wiki'} onclick={() => activeTab = 'wiki'}>WIKI</button>
          <button class="tab-btn" class:active={activeTab === 'logs'} onclick={() => activeTab = 'logs'}>LOGS</button>
        </div>
        {#if activeTab === 'readme'}<div class="readme-container" class:readme-expanded={readmeExpanded}><div style="max-height: {readmeExpanded ? 'none' : '300px'}; overflow: hidden;"><div class="readme-content">{@html readmeContent}</div></div>{#if !readmeExpanded}<div class="readme-fade"><button class="secondary" style="font-size: 0.65rem; padding: 12px 24px;" onclick={() => readmeExpanded = true}>READ FULL MISSION BRIEFING</button></div>{/if}</div>
        {:else if activeTab === 'issues'}<div style="display: flex; flex-direction: column; gap: 16px;">{#if issues.length === 0}<p style="color: var(--efinity-text-muted)">No tactical issues found in this sector.</p>{:else}{#each issues as issue}<div style="background: rgba(255,255,255,0.02); border-radius: 16px; border: 1px solid var(--glass-border); padding: 20px;"><div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 8px;"><div style="font-weight: 700; font-size: 1.1rem; color: #fff;">{issue.title}</div><span class="badge" style="color: {issue.state === 'open' ? 'var(--status-green)' : 'var(--efinity-text-muted)'}; background: {issue.state === 'open' ? 'rgba(0, 255, 136, 0.05)' : 'rgba(255,255,255,0.03)'}">{issue.state.toUpperCase()}</span></div><div style="font-size: 0.8rem; color: var(--efinity-text-muted);">#{issue.number} opened by {issue.user?.login} • {new Date(issue.created_at).toLocaleDateString()}</div></div>{/each}{/if}</div>
        {:else if activeTab === 'releases'}<div style="display: flex; flex-direction: column; gap: 24px;">{#if releases.length === 0}<p style="color: var(--efinity-text-muted)">No historical chronicles (releases) found.</p>{:else}{#each releases as release}<div style="background: rgba(255,255,255,0.02); border-radius: 20px; border: 1px solid var(--glass-border); padding: 24px;"><div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; border-bottom: 1px solid var(--glass-border); padding-bottom: 16px;"><div><div style="font-size: 1.4rem; font-weight: 800; color: var(--efinity-blue);">{release.tag_name}</div><div style="font-size: 0.8rem; color: var(--efinity-text-muted);">{release.name} • {new Date(release.published_at).toLocaleDateString()}</div></div></div><div class="readme-content" style="font-size: 0.9rem;">{@html marked.parse(release.body || '')}</div></div>{/each}{/if}</div>
        {:else if activeTab === 'wiki'}<div class="readme-content">{@html wikiContent}</div>
        {:else if activeTab === 'logs'}<div style="display: flex; flex-direction: column; gap: 12px;">{#each parseCommits(selectedRepo.last_commit) as commit}<div style="background: rgba(255,255,255,0.02); border-radius: 16px; border: 1px solid var(--glass-border); padding: 20px;"><div style="display: flex; align-items: center; gap: 12px; margin-bottom: 8px;">{#if commit.branch}<span class="branch-badge">{commit.branch}</span>{/if}<span style="font-weight: 600; font-size: 1rem;">{commit.message}</span></div><div style="font-size: 0.8rem; color: var(--efinity-text-muted); font-weight: 500;">{commit.author} • {commit.date} • <span style="color: var(--efinity-blue); font-family: monospace;">{commit.hash.substring(0,7)}</span></div></div>{/each}</div>{/if}
      </div>
    </div>
  </div>
{/if}

{#if showConfigModal && selectedRepo}
  <div class="modal-overlay" onclick={() => showConfigModal = false} style="z-index: 2000;">
    <div class="modal-content" onclick={(e) => e.stopPropagation()} style="max-width: 500px; border: 1px solid var(--efinity-blue);">
      <div class="modal-header">
        <h2 style="margin: 0; font-size: 1.5rem; font-weight: 800;">Asset Configuration</h2>
        <p style="color: var(--efinity-text-muted); margin: 4px 0 0 0; font-size: 0.8rem;">{selectedRepo.name}</p>
      </div>
      <div class="modal-body">
        <div style="margin-bottom: 24px; display: flex; align-items: center; justify-content: space-between; background: rgba(255,255,255,0.02); padding: 20px; border-radius: 16px; border: 1px solid var(--glass-border);">
          <div>
            <div style="font-weight: 700; font-size: 0.9rem;">ACTIVE PATROL</div>
            <div style="font-size: 0.75rem; color: var(--efinity-text-muted);">Enable automated synchronization.</div>
          </div>
          <input type="checkbox" checked={selectedRepo.auto_patrol === 1} onchange={(e) => selectedRepo!.auto_patrol = e.currentTarget.checked ? 1 : 0} style="width: 24px; height: 24px; margin: 0; cursor: pointer; accent-color: var(--efinity-blue);" />
        </div>
        {#if selectedRepo.auto_patrol === 1}<label>SYNC INTERVAL (e.g. 1h 30m, 1d 2h)</label><input type="text" bind:value={intervalString} placeholder="1h" />{:else}<div style="padding: 16px; background: rgba(255, 77, 77, 0.05); border: 1px solid rgba(255, 77, 77, 0.1); border-radius: 12px; color: var(--status-red); font-size: 0.8rem; margin-bottom: 24px;">⚠️ Automated patrolling is disabled for this asset.</div>{/if}
        <div style="display: flex; gap: 16px; margin-top: 32px;"><button style="flex: 2;" onclick={updateConfig}>SAVE CHANGES</button><button class="secondary" style="flex: 1;" onclick={() => showConfigModal = false}>CANCEL</button></div>
      </div>
    </div>
  </div>
{/if}

{#if showAddModal}
  <div class="modal-overlay" onclick={() => { showAddModal = false; newUrl = ''; newName = ''; }}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()} style="max-width: 540px;">
      <div class="modal-header"><h2 style="margin: 0; font-size: 1.8rem; font-weight: 800;">Deploy New Patrol</h2><p style="color: var(--efinity-text-muted); margin: 8px 0 0 0; font-size: 0.9rem;">Configure a new asset for monitoring.</p></div>
      <div class="modal-body">
        <label>REPOSITORY URL</label>
        <input bind:value={newUrl} placeholder="https://github.com/hoppscotch/hoppscotch" style="margin-bottom: 8px;" />
        {#if getNormalizedUrl(newUrl)}<div style="margin-bottom: 24px;"><a href={getNormalizedUrl(newUrl)} target="_blank" rel="noopener noreferrer" style="font-size: 0.75rem; color: var(--efinity-blue); text-decoration: none; font-family: monospace; opacity: 0.8;"><svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" style="margin-right: 4px;"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6M15 3h6v6M10 14L21 3"/></svg>{getNormalizedUrl(newUrl)}</a></div>{/if}
        {#if newUrl.length > 3}<div transition:fade><label>DISPLAY NAME</label><input bind:value={newName} placeholder="e.g. Hoppscotch" /></div>{/if}
        <div style="margin-bottom: 24px; display: flex; align-items: center; justify-content: space-between; background: rgba(255,255,255,0.02); padding: 20px; border-radius: 16px; border: 1px solid var(--glass-border);"><div><div style="font-weight: 700; font-size: 0.9rem;">KEEP ACTIVE PATROL</div><div style="font-size: 0.75rem; color: var(--efinity-text-muted);">Continuously monitor and sync this asset.</div></div><input type="checkbox" bind:checked={autoPatrol} style="width: 24px; height: 24px; margin: 0; cursor: pointer; accent-color: var(--efinity-blue);" /></div>
        {#if autoPatrol}<label>SYNC INTERVAL (e.g. 1h 30m, 1d 2h)</label><input type="text" bind:value={intervalString} placeholder="1h" />{/if}
        <div style="display: flex; gap: 16px; margin-top: 32px;"><button style="flex: 2;" onclick={addRepo}>ACTIVATE</button><button class="secondary" style="flex: 1;" onclick={() => { showAddModal = false; newUrl = ''; newName = ''; }}>CANCEL</button></div>
      </div>
    </div>
  </div>
{/if}

{#if showIncidentModal}
  <div class="modal-overlay" onclick={() => showIncidentModal = false}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()} style="max-width: 600px;">
      <div class="modal-header" style="display: flex; justify-content: space-between; align-items: center;"><div><h2 style="margin: 0; font-size: 1.8rem; font-weight: 800;">Security Logs</h2><p style="color: var(--efinity-text-muted); margin: 4px 0 0 0; font-size: 0.8rem;">Historical synchronization incidents.</p></div><button class="secondary" onclick={clearIncidents} style="font-size: 0.7rem; padding: 10px 20px;">CLEAR ALL</button></div>
      <div class="modal-body" style="max-height: 60vh;">{#if incidents.length === 0}<div style="text-align: center; padding: 40px; color: var(--efinity-text-muted);"><svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" style="margin-bottom: 16px; opacity: 0.3;"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg><p>No incidents detected. All systems nominal.</p></div>{:else}{#each incidents as incident}<div class="incident-item"><div style="display: flex; justify-content: space-between; margin-bottom: 8px;"><div style="font-weight: 800; color: #fff; font-size: 0.9rem;">{incident.repo_name}</div><div style="font-size: 0.7rem; color: var(--efinity-text-muted);">{new Date(incident.created_at).toLocaleString()}</div></div><div style="font-size: 0.85rem; color: var(--status-red); line-height: 1.4;">{incident.message}</div></div>{/each}{/if}</div>
    </div>
  </div>
{/if}

<div class="toast-container">
  {#each toasts as toast (toast.id)}
    <div class="toast {toast.type}" transition:fade>
      {#if toast.type === 'error'}<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>{/if}
      {#if toast.type === 'success'}<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><path d="M20 6L9 17l-5-5"/></svg>{/if}
      {toast.message}
    </div>
  {/each}
</div>
