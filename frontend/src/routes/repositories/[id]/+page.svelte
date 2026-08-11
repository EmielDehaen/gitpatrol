<script lang="ts">
  import { page } from '$app/stores';
  import { onMount, tick } from 'svelte';
  import { fade } from 'svelte/transition';
  import { reposStore } from '$lib/repos.svelte';
  import { settingsApi } from '$lib/settings/settingsApi';
  import { parseMarkdown } from '$lib/markdown';
  import { API_URL, apiFetch } from '$lib/client';
  import { toastHandler } from '$lib/toast.svelte';
  import { getAvatarUrl, getRemainingTime, handleAvatarError, humanToMinutes, minutesToHuman } from '$lib/utils';
  import { authStore } from '$lib/auth.svelte';
  import type { Repository, Commit, Issue, Release } from '$lib/types';
  import { goto } from '$app/navigation';

  import Icon from '$lib/components/Icon.svelte';
  import RepoReadme from '$lib/components/Repo/RepoReadme.svelte';
  import RepoIssues from '$lib/components/Repo/RepoIssues.svelte';
  import RepoReleases from '$lib/components/Repo/RepoReleases.svelte';
  import RepoWiki from '$lib/components/Repo/RepoWiki.svelte';
  import RepoLogs from '$lib/components/Repo/RepoLogs.svelte';

  // Derived repo from store
  let repoId = $derived(Number($page.params.id));
  let repo = $derived(reposStore.repositories.find(r => r.id === repoId));

  let activeTab = $state<'readme' | 'issues' | 'releases' | 'wiki' | 'commits' | 'terminal' | 'settings'>('readme');
  let readmeContent = $state('Loading mission briefing...');
  let wikiContent = $state('');
  let issues = $state<Issue[]>([]);
  let releases = $state<Release[]>([]);
  let now = $state(Date.now());
  let exportDestination = $state('');

  // Settings State
  let intervalString = $state('');
  let autoPatrol = $state(false);

  // Terminal State
  let terminalLines = $state<string[]>([]);
  let terminalContainer = $state<HTMLDivElement>();

  async function fetchSettings() {
    const settings = await settingsApi.getSettings();
    if (settings) {
      exportDestination = settings.export_destination;
    }
  }

  // Terminal logic
  function pushTerminalLog(line: string) {
    const timestamp = new Date().toLocaleTimeString([], { hour12: false });
    terminalLines = [...terminalLines, `[${timestamp}] ${line}`];
    tick().then(() => {
      if (terminalContainer) {
        terminalContainer.scrollTop = terminalContainer.scrollHeight;
      }
    });
  }

  $effect(() => {
    if (repo) {
      intervalString = minutesToHuman(repo.interval_minutes);
      autoPatrol = repo.auto_patrol === 1;
      
      // Initialize terminal history based on repo status
      if (terminalLines.length === 0) {
        pushTerminalLog(`GitPatrol Terminal Session Started for ${repo.name}`);
        pushTerminalLog(`Target: ${repo.url}`);
        if (repo.last_sync) {
          pushTerminalLog(`Last known sync: ${new Date(repo.last_sync).toLocaleString()}`);
        }
        pushTerminalLog(`Current status: ${repo.status.toUpperCase()}`);
        if (repo.error_message) {
          pushTerminalLog(`ERROR: ${repo.error_message}`);
        }
      }
    }
  });

  onMount(() => {
    fetchSettings();
    const interval = setInterval(() => { now = Date.now(); }, 1000);

    // Terminal websocket hooking
    const wsBaseUrl = API_URL ? API_URL.replace('http', 'ws') : (window.location.protocol === 'https:' ? 'wss:' : 'ws:') + '//' + window.location.host;
    const wsUrl = wsBaseUrl + '/ws';
    const ws = new WebSocket(wsUrl);
    ws.onmessage = (event) => {
      if (authStore.isAuthenticated) {
        try {
          const data = JSON.parse(event.data);
          if (data.type === 'status_update' && Number(data.id) === repoId) {
            pushTerminalLog(`Received status update: ${data.status.toUpperCase()}`);
            if (data.status === 'syncing') pushTerminalLog(`> Executing git fetch origin...`);
            if (data.status === 'synced') pushTerminalLog(`> Sync successful. Remote mirrors updated.`);
            if (data.status === 'error') pushTerminalLog(`> ERROR: ${data.error}`);
          }
        } catch (e) {}
      }
    };

    return () => {
      clearInterval(interval);
      ws.close();
    };
  });

  async function fetchMetadata() {
    if (!repo) return;
    const assetBase = `/api/repositories/${repo.id}/assets/`;
    try {
      const issuesRes = await apiFetch(`${assetBase}metadata/issues.json`);
      issues = issuesRes.ok ? await issuesRes.json() : [];

      const releasesRes = await apiFetch(`${assetBase}metadata/releases.json`);
      releases = releasesRes.ok ? await releasesRes.json() : [];

      const wikiRes = await apiFetch(`${assetBase}wiki/Home.md`);
      if (wikiRes.ok) {
        wikiContent = await parseMarkdown(await wikiRes.text());
      } else {
        wikiContent = '<p style="color: var(--on-surface-variant)">No documentation (Wiki) found for this asset.</p>';
      }
    } catch (e) {
      toastHandler.showToast('Network error: could not load asset metadata.', 'error');
    }
  }

  async function fetchReadme() {
    if (!repo) return;
    readmeContent = 'Loading mission briefing...';
    try {
      const res = await apiFetch(`/api/repositories/${repo.id}/readme`);
      if (res.ok) {
        let text = await res.text();
        const assetBase = `${API_URL}/api/repositories/${repo.id}/assets/`;
        
        text = text.replace(/!\[([^\]]*)\]\((?!(?:http|https|ftp|data:))(?:\.\/)?([^)]+)\)/gi, `![$1](${assetBase}$2)`);
        text = text.replace(/<img([^>]*?)src=["'](?!(?:http|https|ftp))(?:\.\/)?([^"']+)["']/gi, (match, pre, path) => `<img${pre}src="${assetBase}${path}"`);
        text = text.replace(/<source([^>]*?)srcset=["'](?!(?:http|https|ftp))(?:\.\/)?([^"']+)["']/gi, (match, pre, path) => `<source${pre}srcset="${assetBase}${path}"`);
        text = text.replace(/\[([^\]]*)\]\((?!(?:http|https|ftp|#))(?:\.\/)?([^)]+)\)/gi, `[$1](${assetBase}$2)`);

        readmeContent = await parseMarkdown(text);
      } else {
        readmeContent = '<p style="color: var(--on-surface-variant)">No mission briefing available for this asset.</p>';
      }
      fetchMetadata();
    } catch (e) {
      readmeContent = '<p style="color: var(--on-surface-variant)">Failed to load mission briefing. Check your connection.</p>';
    }
  }

  function parseCommits(lastCommit: string): Commit[] {
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
    }).filter(c => c !== null) as Commit[];
  }

  async function syncNow() {
    if (!repo) return;
    try {
      pushTerminalLog(`Manual sync initiated by operator.`);
      await apiFetch(`/api/repositories/${repo.id}/sync`, { method: 'POST' });
      toastHandler.showToast('Sync initiated.', 'info');
    } catch (e) {
      toastHandler.showToast('Network error: could not initiate sync.', 'error');
    }
  }

  async function exportRepo() {
    if (!repo) return;
    if (!exportDestination) {
      toastHandler.showToast('No recovery vault configured. Setup tokens in Settings first.', 'error');
      return;
    }
    try {
      pushTerminalLog(`Initiating recovery export to ${exportDestination.toUpperCase()}...`);
      toastHandler.showToast(`Initiating recovery to ${exportDestination.toUpperCase()}...`, 'info');
      const res = await apiFetch(`/api/repositories/${repo.id}/export`, { 
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ destination: exportDestination })
      });
      if (res.ok) {
        const data = await res.json();
        pushTerminalLog(`Recovery successful! Destination URL generated.`);
        toastHandler.showToast('Repository recovered successfully!', 'success');
        if (data.destination_url) {
          window.open(data.destination_url, '_blank');
        }
      } else {
        const data = await res.json();
        pushTerminalLog(`ERROR: Recovery failed - ${data.error}`);
        toastHandler.showToast(data.error || 'Recovery failed.', 'error');
      }
    } catch (e) {
      toastHandler.showToast('Network error during recovery.', 'error');
    }
  }

  // Config Logic
  async function updateConfig() {
    if (!repo) return;
    try {
      pushTerminalLog(`Updating configuration...`);
      const res = await apiFetch(`/api/repositories/${repo.id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          interval_minutes: humanToMinutes(intervalString), 
          auto_patrol: autoPatrol ? 1 : 0
        })
      });
      if (res.ok) { 
        await reposStore.fetchRepos();
        pushTerminalLog(`Configuration applied successfully.`);
        toastHandler.showToast('Configuration updated.', 'success');
      } else {
        toastHandler.showToast('Failed to update configuration.', 'error');
      }
    } catch (e) {
      toastHandler.showToast('Network error.', 'error');
    }
  }

  async function deleteRepo() {
    if (!repo) return;
    if (!confirm('Are you sure you want to terminate this patrol?')) return;
    try {
      const res = await apiFetch(`/api/repositories/${repo.id}`, { method: 'DELETE' });
      if (res.ok) { 
        reposStore.fetchRepos(); 
        toastHandler.showToast('Patrol terminated.', 'info');
        goto('/');
      }
    } catch (e) {
      toastHandler.showToast('Network error.', 'error');
    }
  }

  $effect(() => {
    if (repo?.id) fetchReadme();
  });
</script>

<div class="page-container" transition:fade>
  {#if repo}
    <div class="header">
      <button class="icon-btn" onclick={() => goto('/')} style="margin-right: 24px;">
        <Icon name="x" size={24} />
      </button>
      
      <div style="display: flex; flex: 1; align-items: flex-start; justify-content: space-between;">
        <div style="display: flex; align-items: center; gap: 32px;">
          <img src={getAvatarUrl(repo.url)} onerror={handleAvatarError} alt={repo.name} class="repo-avatar" />
          <div>
            <div style="display: flex; align-items: center; gap: 16px;">
              <h1 class="display-lg" style="margin: 0; font-size: 3rem;">{repo.name}</h1>
              <div class="badge status-healthy" style="font-size: 0.9rem; padding: 6px 16px;">
                <span class="dot"></span>
                {repo.health_score}% HEALTH
              </div>
            </div>
            <a href={repo.url} target="_blank" rel="noopener noreferrer" style="color: var(--primary); text-decoration: none; font-family: monospace; font-size: 1rem; display: block; margin-top: 8px;">{repo.url} ↗</a>
          </div>
        </div>

        <div style="display: flex; gap: 16px;">
          <button class="secondary" style="border-color: rgba(77, 221, 187, 0.2);" onclick={exportRepo} data-tooltip="Recovery Export">
            <Icon name="download" />
          </button>
          <button class="secondary" style="border-color: rgba(77, 221, 187, 0.2);" onclick={syncNow} data-tooltip="Sync Now">
            <Icon name="refresh" />
          </button>
        </div>
      </div>
    </div>

    <div class="hero-stats" style="padding: 0 40px 40px 40px; margin-left: 96px;">
      <div class="stat-block">
        <span class="stat-label">NEXT SYNC</span>
        <span class="stat-value" style="font-size: 1.5rem;">{getRemainingTime(repo, false, now)}</span>
      </div>
      <div class="stat-block">
        <span class="stat-label">INTERVAL</span>
        <span class="stat-value" style="font-size: 1.5rem;">{minutesToHuman(repo.interval_minutes)}</span>
      </div>
    </div>

    <div class="content-area">
      <div class="tab-container" style="margin-bottom: 32px;">
        <button class="tab-btn" class:active={activeTab === 'readme'} onclick={() => activeTab = 'readme'}>BRIEFING</button>
        <button class="tab-btn" class:active={activeTab === 'issues'} onclick={() => activeTab = 'issues'}>INTEL <span class="tab-count">{issues.length}</span></button>
        <button class="tab-btn" class:active={activeTab === 'releases'} onclick={() => activeTab = 'releases'}>CHRONICLE <span class="tab-count">{releases.length}</span></button>
        <button class="tab-btn" class:active={activeTab === 'wiki'} onclick={() => activeTab = 'wiki'}>WIKI</button>
        <button class="tab-btn" class:active={activeTab === 'commits'} onclick={() => activeTab = 'commits'}>COMMITS</button>
        <button class="tab-btn" class:active={activeTab === 'terminal'} onclick={() => activeTab = 'terminal'}>TERMINAL</button>
        <button class="tab-btn" class:active={activeTab === 'settings'} onclick={() => activeTab = 'settings'}>SETTINGS</button>
      </div>

      <div class="tab-content-wrapper">
        {#if activeTab === 'readme'}
          <RepoReadme {readmeContent} />
        {:else if activeTab === 'issues'}
          <RepoIssues {issues} />
        {:else if activeTab === 'releases'}
          <RepoReleases {releases} />
        {:else if activeTab === 'wiki'}
          <RepoWiki {wikiContent} />
        {:else if activeTab === 'commits'}
          <RepoLogs commits={parseCommits(repo.last_commit || '')} />
        {:else if activeTab === 'terminal'}
          <div class="terminal-container" bind:this={terminalContainer}>
            {#each terminalLines as line}
              <div class="terminal-line">{line}</div>
            {/each}
            <div class="terminal-cursor">_</div>
          </div>
        {:else if activeTab === 'settings'}
          <div class="settings-card">
            <h2 style="font-size: 1.25rem; margin-bottom: 24px;">Patrol Configuration</h2>
            
            <label for='interval'>SCAN INTERVAL</label>
            <input id='interval' type="text" bind:value={intervalString} placeholder="e.g. 1h, 30m, 1d" />
            
            <div style="display: flex; align-items: center; gap: 12px; margin: 24px 0 40px 0;">
              <input id='auto_patrol' type="checkbox" bind:checked={autoPatrol} style="width: 20px; height: 20px; margin: 0; cursor: pointer;" />
              <label for='auto_patrol' style="margin: 0; cursor: pointer;">ENABLE AUTO PATROL</label>
            </div>

            <div style="display: flex; gap: 16px; margin-bottom: 32px;">
              <button style="flex: 1;" onclick={updateConfig}>SAVE CONFIGURATION</button>
            </div>

            <div class="divider"></div>

            <h2 style="font-size: 1.25rem; margin-bottom: 24px; color: var(--error);">Danger Zone</h2>
            <button class="secondary" style="width: 100%; color: var(--error); border-color: var(--error-container);" onclick={deleteRepo}>
              TERMINATE PATROL
            </button>
          </div>
        {/if}
      </div>
    </div>
  {:else}
    <div style="padding: 100px; text-align: center;">
      <h2 style="font-size: 2rem;">Repository not found</h2>
      <button onclick={() => goto('/')} style="margin-top: 24px;">RETURN TO OVERVIEW</button>
    </div>
  {/if}
</div>

<style>
  .header {
    padding: 40px;
    display: flex;
    align-items: center;
  }
  
  .repo-avatar {
    width: 80px;
    height: 80px;
    border-radius: 20px;
    border: none;
    background: var(--surface-container-highest);
  }

  .content-area {
    padding: 0 40px 80px 40px;
    max-width: 1400px;
  }

  .tab-content-wrapper {
    background: var(--surface-container-low);
    border-radius: 24px;
    padding: 40px;
    min-height: 500px;
  }

  .settings-card {
    max-width: 600px;
  }

  .divider {
    height: 1px;
    background: var(--surface-container-highest);
    margin: 32px 0;
  }

  /* Terminal UI */
  .terminal-container {
    background: #000000;
    border-radius: 12px;
    padding: 24px;
    font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
    font-size: 0.9rem;
    line-height: 1.6;
    color: #4ade80; /* Hacker green */
    height: 500px;
    overflow-y: auto;
    border: 1px solid rgba(77, 221, 187, 0.2);
    box-shadow: inset 0 0 40px rgba(0, 0, 0, 0.5);
  }

  .terminal-line {
    word-break: break-all;
    margin-bottom: 4px;
  }

  .terminal-cursor {
    display: inline-block;
    width: 10px;
    animation: blink 1s step-end infinite;
  }

  @keyframes blink {
    0%, 100% { opacity: 1; }
    50% { opacity: 0; }
  }
</style>
