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

  function suggestName(url: string) {
    if (!url) return '';
    let cleanUrl = url.trim().replace(/\/$/, '');
    let name = '';

    if (cleanUrl.includes('github.com/')) {
      const parts = cleanUrl.split('github.com/')[1].split('/');
      name = parts.length >= 2 ? parts[1] : parts[0];
    } else if (cleanUrl.includes('gitlab.com/')) {
      const parts = cleanUrl.split('gitlab.com/')[1].split('/');
      name = parts[parts.length - 1];
    } else if (cleanUrl.includes(':')) { // SSH
      const parts = cleanUrl.split(':');
      const repoParts = parts[parts.length - 1].split('/');
      name = repoParts[repoParts.length - 1];
    } else {
      const parts = cleanUrl.split('/');
      name = parts[parts.length - 1];
    }

    name = name.replace('.git', '');
    return name ? name.charAt(0).toUpperCase() + name.slice(1) : '';
  }

  $effect(() => {
    if (newUrl && !newName) {
      newName = suggestName(newUrl);
    }
  });

  const API_URL = 'http://localhost:8080';

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
    const res = await fetch(`${API_URL}/api/repositories`);
    if (!res.ok) return;
    repositories = await res.json();
  }

  async function fetchMetadata(repo: Repository) {
    const assetBase = `${API_URL}/api/repositories/${repo.id}/assets/`;
    
    // Fetch Issues
    const issuesRes = await fetch(`${assetBase}metadata/issues.json`);
    issues = issuesRes.ok ? await issuesRes.json() : [];

    // Fetch Releases
    const releasesRes = await fetch(`${assetBase}metadata/releases.json`);
    releases = releasesRes.ok ? await releasesRes.json() : [];

    // Fetch Wiki (Try Home.md)
    const wikiRes = await fetch(`${assetBase}wiki/Home.md`);
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
    const res = await fetch(`${API_URL}/api/repositories/${repo.id}/readme`);
    if (res.ok) {
      let text = await res.text();
      const assetBase = `${API_URL}/api/repositories/${repo.id}/assets/`;
      
      // Fix paths
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
      body: JSON.stringify({ 
        name: newName, 
        url: newUrl, 
        interval_minutes: humanToMinutes(intervalString), 
        auto_patrol: autoPatrol ? 1 : 0 
      })
    });
    if (res.ok) {
      newName = ''; newUrl = ''; autoPatrol = true; intervalString = '1h'; showAddModal = false; fetchRepos();
    }
  }

  async function deleteRepo(id: number) {
    if (!confirm('Are you sure you want to terminate this patrol?')) return;
    const res = await fetch(`${API_URL}/api/repositories/${id}`, { method: 'DELETE' });
    if (res.ok) { selectedRepo = null; showConfigModal = false; fetchRepos(); }
  }

  async function updateConfig() {
    if (!selectedRepo) return;
    const res = await fetch(`${API_URL}/api/repositories/${selectedRepo.id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        interval_minutes: humanToMinutes(intervalString), 
        auto_patrol: selectedRepo.auto_patrol 
      })
    });
    if (res.ok) { 
      await fetchRepos();
      selectedRepo = repositories.find(r => r.id === selectedRepo?.id) || null;
      showConfigModal = false; 
    }
  }

  async function syncRepoNow(repo: Repository) {
    await fetch(`${API_URL}/api/repositories/${repo.id}/sync`, { method: 'POST' });
  }

  function openConfig() {
    if (selectedRepo) {
      intervalString = minutesToHuman(selectedRepo.interval_minutes);
      showConfigModal = true;
    }
  }

  function getProgress(repo: Repository) {
    if (!repo.last_sync || repo.status === 'syncing' || repo.auto_patrol === 0) return 0;
    const lastSync = new Date(repo.last_sync).getTime();
    const nextSync = lastSync + repo.interval_minutes * 60000;
    const total = repo.interval_minutes * 60000;
    const remaining = nextSync - Date.now();
    if (remaining <= 0) return 0;
    return 100 * (remaining / total);
  }

  function getRemainingTime(repo: Repository) {
    if (repo.auto_patrol === 0) return 'Manual Patrol Only';
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
      const refs = parts[4] || '';
      let branch = '';
      if (refs) {
        const cleanRefs = refs.replace(/[()]/g, '').split(', ');
        const priorityRef = cleanRefs.find(r => !r.includes('HEAD') && !r.startsWith('tag:')) || cleanRefs[0];
        branch = priorityRef?.split(' -> ').pop()?.replace('remotes/origin/', '').replace('origin/', '').trim() || '';
      }
      return { hash: parts[0], author: parts[1], date: parts[2], message: parts[3], branch };
    }).filter(c => c !== null);
  }

  function handleAvatarError(e: Event) {
    const img = e.target as HTMLImageElement;
    img.src = "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png";
  }

  function getAvatarUrl(url: string) {
    const parts = url.replace('https://github.com/', '').split('/');
    if (parts.length > 0) return `${API_URL}/avatars/${parts[0]}.png`;
    return '';
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
            <div class="radial-timer" data-tooltip={getRemainingTime(repo)}>
              <svg width="40" height="40">
                <circle cx="20" cy="20" r="16" />
                <circle cx="20" cy="20" r="16" class="progress" class:active-pulse={repo.status !== 'syncing' && repo.auto_patrol === 1}
                  style="stroke-dasharray: 100; stroke-dashoffset: {100 - (repo.progress || 0)}" />
              </svg>
            </div>
            <div class="health-score" data-tooltip="Tactical Health Score" style="color: {repo.health_score > 70 ? 'var(--status-green)' : 'var(--status-yellow)'}; border-color: {repo.health_score > 70 ? 'var(--status-green)' : repo.health_score > 40 ? 'var(--status-yellow)' : 'var(--status-red)'}44">
              {repo.health_score}
            </div>
          </div>
          <div class="card-header">
            <div style="display: flex; gap: 16px; align-items: flex-start;">
              <img src={getAvatarUrl(repo.url)} onerror={handleAvatarError} alt="" style="width: 44px; height: 44px; border-radius: 12px; background: var(--glass); border: 1px solid var(--glass-border);" />
              <div>
                <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 4px;">
                  <h3 class="repo-name" style="margin: 0;">{repo.name}</h3>
                  {#if repo.status === 'synced'}
                    <span class="badge" style="color: var(--status-green); background: rgba(0, 255, 136, 0.05); font-size: 0.6rem; padding: 2px 8px;">SYNCED</span>
                  {/if}
                </div>
                <div class="repo-url">{repo.url.replace('https://github.com/', '')}</div>
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
        </div>
      {/each}
    </div>
  {:else}
    <div class="repo-list">
      {#each repositories as repo (repo.id)}
        <div class="list-item" onclick={() => selectedRepo = repo}>
          <div style="display: flex; align-items: center; gap: 24px;">
            <div class="radial-timer" data-tooltip={getRemainingTime(repo)} style="width: 32px; height: 32px;">
              <svg width="32" height="32"><circle cx="16" cy="16" r="12" /><circle cx="16" cy="16" r="12" class="progress" class:active-pulse={repo.status !== 'syncing' && repo.auto_patrol === 1} style="stroke-dasharray: 75; stroke-dashoffset: {75 - (repo.progress || 0) * 0.75}" /></svg>
            </div>
            <img src={getAvatarUrl(repo.url)} onerror={handleAvatarError} alt="" style="width: 32px; height: 32px; border-radius: 8px; border: 1px solid var(--glass-border);" />
            <div>
              <div style="display: flex; align-items: center; gap: 12px;">
                <div style="font-weight: 700; font-size: 1.1rem;">{repo.name}</div>
                {#if repo.status === 'synced'}<span class="badge" style="color: var(--status-green); background: rgba(0, 255, 136, 0.05); font-size: 0.6rem; padding: 2px 8px;">SYNCED</span>{/if}
              </div>
              <div style="font-size: 0.8rem; color: var(--efinity-text-muted);">{repo.url.replace('https://github.com/', '')}</div>
            </div>
          </div>
          <div style="display: flex; gap: 40px; align-items: center;">
            <div class="stat-item"><svg width="12" height="12" viewBox="0 0 24 24" fill="var(--status-yellow)"><path d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"/></svg><b>{repo.stars}</b></div>
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
  <div class="modal-overlay" onclick={() => { selectedRepo = null; showConfig = false; }}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <div style="display: flex; align-items: center; gap: 24px;">
            <img src={getAvatarUrl(selectedRepo.url)} onerror={handleAvatarError} alt="" style="width: 64px; height: 64px; border-radius: 16px; border: 1px solid var(--glass-border);" />
            <div>
              <div style="display: flex; align-items: center; gap: 16px;">
                <h2 style="margin: 0; font-size: 2.2rem; font-weight: 800;">{selectedRepo.name}</h2>
                <div class="radial-timer" data-tooltip={getRemainingTime(selectedRepo)} style="width: 32px; height: 32px;">
                  <svg width="32" height="32"><circle cx="16" cy="16" r="12" /><circle cx="16" cy="16" r="12" class="progress" class:active-pulse={selectedRepo.status !== 'syncing' && selectedRepo.auto_patrol === 1} style="stroke-dasharray: 75; stroke-dashoffset: {75 - (selectedRepo.progress || 0) * 0.75}" /></svg>
                </div>
                <div class="badge" style="color: {selectedRepo.health_score > 70 ? 'var(--status-green)' : 'var(--status-yellow)'}; background: rgba(255,255,255,0.03); font-size: 0.8rem; padding: 4px 12px; border: 1px solid rgba(255,255,255,0.05);">{selectedRepo.health_score}% HEALTH</div>
              </div>
              <a href={selectedRepo.url} target="_blank" rel="noopener noreferrer" style="color: var(--efinity-blue); text-decoration: none; font-family: monospace; font-size: 0.95rem; display: block; margin-top: 8px;">{selectedRepo.url} ↗</a>
            </div>
          </div>
          <div style="display: flex; gap: 12px; align-items: center;">
            {#if selectedRepo.auto_patrol === 0 || selectedRepo.status !== 'syncing'}
              <button class="secondary" style="padding: 10px; border-radius: 12px; color: var(--status-green); border-color: rgba(0, 255, 136, 0.2);" onclick={() => syncRepoNow(selectedRepo!)} data-tooltip="Sync Now"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21.5 2v6h-6M2.5 22v-6h6M2 12c0-4.4 3.6-8 8-8 3.3 0 6.1 2 7.3 4.9M22 12c0 4.4-3.6 8-8 8-3.3 0-6.1-2-7.3-4.9"/></svg></button>
            {/if}
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

        {#if activeTab === 'readme'}
          <div class="readme-container" class:readme-expanded={readmeExpanded}>
            <div style="max-height: {readmeExpanded ? 'none' : '300px'}; overflow: hidden;">
              <div class="readme-content">{@html readmeContent}</div>
            </div>
            {#if !readmeExpanded}
              <div class="readme-fade">
                <button class="secondary" style="font-size: 0.65rem; padding: 12px 24px;" onclick={() => readmeExpanded = true}>READ FULL MISSION BRIEFING</button>
              </div>
            {/if}
          </div>
        {:else if activeTab === 'issues'}
          <div style="display: flex; flex-direction: column; gap: 16px;">
            {#if issues.length === 0}
              <p style="color: var(--efinity-text-muted)">No tactical issues found in this sector.</p>
            {:else}
              {#each issues as issue}
                <div style="background: rgba(255,255,255,0.02); border-radius: 16px; border: 1px solid var(--glass-border); padding: 20px;">
                  <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 8px;">
                    <div style="font-weight: 700; font-size: 1.1rem; color: #fff;">{issue.title}</div>
                    <span class="badge" style="color: {issue.state === 'open' ? 'var(--status-green)' : 'var(--efinity-text-muted)'}; background: {issue.state === 'open' ? 'rgba(0, 255, 136, 0.05)' : 'rgba(255,255,255,0.03)'}">{issue.state.toUpperCase()}</span>
                  </div>
                  <div style="font-size: 0.8rem; color: var(--efinity-text-muted);">#{issue.number} opened by {issue.user?.login} • {new Date(issue.created_at).toLocaleDateString()}</div>
                </div>
              {/each}
            {/if}
          </div>
        {:else if activeTab === 'releases'}
          <div style="display: flex; flex-direction: column; gap: 24px;">
            {#if releases.length === 0}
              <p style="color: var(--efinity-text-muted)">No historical chronicles (releases) found.</p>
            {:else}
              {#each releases as release}
                <div style="background: rgba(255,255,255,0.02); border-radius: 20px; border: 1px solid var(--glass-border); padding: 24px;">
                  <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; border-bottom: 1px solid var(--glass-border); padding-bottom: 16px;">
                    <div>
                      <div style="font-size: 1.4rem; font-weight: 800; color: var(--efinity-blue);">{release.tag_name}</div>
                      <div style="font-size: 0.8rem; color: var(--efinity-text-muted);">{release.name} • {new Date(release.published_at).toLocaleDateString()}</div>
                    </div>
                  </div>
                  <div class="readme-content" style="font-size: 0.9rem;">{@html marked.parse(release.body || '')}</div>
                </div>
              {/each}
            {/if}
          </div>
        {:else if activeTab === 'wiki'}
          <div class="readme-content">{@html wikiContent}</div>
        {:else if activeTab === 'logs'}
          <div style="display: flex; flex-direction: column; gap: 12px;">
            {#each parseCommits(selectedRepo.last_commit) as commit}
              <div style="background: rgba(255,255,255,0.02); border-radius: 16px; border: 1px solid var(--glass-border); padding: 20px;">
                <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 8px;">{#if commit.branch}<span class="branch-badge">{commit.branch}</span>{/if}<span style="font-weight: 600; font-size: 1rem;">{commit.message}</span></div>
                <div style="font-size: 0.8rem; color: var(--efinity-text-muted); font-weight: 500;">{commit.author} • {commit.date} • <span style="color: var(--efinity-blue); font-family: monospace;">{commit.hash.substring(0,7)}</span></div>
              </div>
            {/each}
          </div>
        {/if}
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

        {#if selectedRepo.auto_patrol === 1}
          <label>SYNC INTERVAL (e.g. 1h 30m, 1d 2h)</label>
          <input type="text" bind:value={intervalString} placeholder="1h" />
        {:else}
          <div style="padding: 16px; background: rgba(255, 77, 77, 0.05); border-radius: 12px; color: var(--status-red); font-size: 0.8rem; border: 1px solid rgba(255, 77, 77, 0.1); margin-bottom: 24px;">
            ⚠️ Automated patrolling is disabled for this asset.
          </div>
        {/if}

        <div style="display: flex; gap: 16px; margin-top: 32px;">
          <button style="flex: 2;" onclick={updateConfig}>SAVE CHANGES</button>
          <button class="secondary" style="flex: 1;" onclick={() => showConfigModal = false}>CANCEL</button>
        </div>
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
        <input bind:value={newUrl} placeholder="https://github.com/hoppscotch/hoppscotch" />
        
        {#if newUrl.length > 3}
          <div transition:fade>
            <label>DISPLAY NAME (SUGGESTED)</label>
            <input bind:value={newName} placeholder="e.g. hoppscotch" />
          </div>
        {/if}

        <div style="margin-bottom: 24px; display: flex; align-items: center; justify-content: space-between; background: rgba(255,255,255,0.02); padding: 20px; border-radius: 16px; border: 1px solid var(--glass-border);">
          <div>
            <div style="font-weight: 700; font-size: 0.9rem;">KEEP ACTIVE PATROL</div>
            <div style="font-size: 0.75rem; color: var(--efinity-text-muted);">Continuously monitor and sync this asset.</div>
          </div>
          <input type="checkbox" bind:checked={autoPatrol} style="width: 24px; height: 24px; margin: 0; cursor: pointer; accent-color: var(--efinity-blue);" />
        </div>

        {#if autoPatrol}
          <label>SYNC INTERVAL (e.g. 1h 30m, 1d 2h)</label>
          <input type="text" bind:value={intervalString} placeholder="1h" />
        {/if}

        <div style="display: flex; gap: 16px; margin-top: 32px;">
          <button style="flex: 2;" onclick={addRepo}>ACTIVATE</button>
          <button class="secondary" style="flex: 1;" onclick={() => { showAddModal = false; newUrl = ''; newName = ''; }}>CANCEL</button>
        </div>
      </div>
    </div>
  </div>
{/if}
