<script lang="ts">
  import { onMount } from 'svelte';
  import { settingsApi } from '$lib/settings/settingsApi';
  import { parseMarkdown } from '$lib/markdown';

  import { type Repository, type Commit, type Issue, type Release } from '$lib/types';
  import { API_URL, apiFetch } from '$lib/client';
  import { toastHandler } from '$lib/toast.svelte';
  import { fade } from 'svelte/transition';
  import { getAvatarUrl, getRemainingTime, handleAvatarError, minutesToHuman } from '$lib/utils';
  import RepoReadme from './RepoReadme.svelte';
  import RepoIssues from './RepoIssues.svelte';
  import RepoReleases from './RepoReleases.svelte';
  import RepoWiki from './RepoWiki.svelte';
  import RepoLogs from './RepoLogs.svelte';
  import Icon from '$lib/components/Icon.svelte';

  let { repo = $bindable(), selectedRepo = $bindable(), showConfigModal = $bindable() } = $props<{ repo: Repository | null, selectedRepo: Repository | null, showConfigModal: boolean }>();

  let activeTab = $state<'readme' | 'issues' | 'releases' | 'wiki' | 'logs'>('readme');
  let readmeContent = $state('Loading mission briefing...');
  let wikiContent = $state('');
  let issues = $state<Issue[]>([]);
  let releases = $state<Release[]>([]);
  let now = $state(Date.now());
  let exportDestination = $state('');

  async function fetchSettings() {
    const settings = await settingsApi.getSettings();
    exportDestination = settings.export_destination;
  }

  onMount(() => {
    fetchSettings();
    const interval = setInterval(() => {
      now = Date.now();
    }, 1000);
    return () => clearInterval(interval);
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
        wikiContent = '<p style="color: var(--efinity-text-muted)">No documentation (Wiki) found for this asset.</p>';
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
        readmeContent = '<p style="color: var(--efinity-text-muted)">No mission briefing available for this asset.</p>';
      }
      fetchMetadata();
    } catch (e) {
      readmeContent = '<p style="color: var(--efinity-text-muted)">Failed to load mission briefing. Check your connection.</p>';
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
      await apiFetch(`/api/repositories/${repo.id}/sync`, { method: 'POST' });
      toastHandler.showToast('Sync initiated.', 'info');
    } catch (e) {
      toastHandler.showToast('Network error: could not initiate sync.', 'error');
    }
  }

  async function exportRepo() {
    if (!repo) return;
    if (!exportDestination) {
      toastHandler.showToast('No recovery vault configured. Setup tokens in User Configuration first.', 'error');
      return;
    }

    try {
      toastHandler.showToast(`Initiating recovery to ${exportDestination.toUpperCase()}...`, 'info');
      const res = await apiFetch(`/api/repositories/${repo.id}/export`, { 
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ destination: exportDestination })
      });
      if (res.ok) {
        const data = await res.json();
        toastHandler.showToast('Repository recovered successfully!', 'success');
        if (data.destination_url) {
          window.open(data.destination_url, '_blank');
        }
      } else {
        const data = await res.json();
        toastHandler.showToast(data.error || 'Recovery failed.', 'error');
      }
    } catch (e) {
      toastHandler.showToast('Network error during recovery.', 'error');
    }
  }

  $effect(() => {
    if (repo?.id) fetchReadme();
  });
</script>

<div class="modal-overlay" onkeydown={(e) => e.key === 'Escape' && (selectedRepo = null)} onclick={() => selectedRepo = null} role="presentation" transition:fade>
  {#if repo}
    <div class="modal-content" onclick={(e) => e.stopPropagation()} role='none'>
      <div class="modal-header">
        <div style="display: flex; justify-content: space-between; align-items: flex-start;">
          <div style="display: flex; align-items: center; gap: 2rem;">
            <img src={getAvatarUrl(repo?.url)} onerror={handleAvatarError} alt={repo?.name} style="width: 64px; height: 64px; border-radius: 16px; border: 1px solid var(--glass-border);" />
            <div>
              <div style="display: flex; align-items: center; gap: 16px;">
                <h2 style="font-size: 2.5rem; margin: 0; letter-spacing: -0.04em;">{repo?.name || ''}</h2>
                <div class="badge" style="color: {repo?.health_score > 70 ? 'var(--status-green)' : 'var(--status-yellow)'}; background: rgba(255,255,255,0.03); font-size: 0.8rem; padding: 4px 12px; border: 1px solid rgba(255,255,255,0.05);">
                  {repo?.health_score || 0}% HEALTH
                </div>
              </div>
              {#if repo?.url}
                <a href={repo?.url} target="_blank" rel="noopener noreferrer" style="color: var(--efinity-blue); text-decoration: none; font-family: monospace; font-size: 0.95rem; display: block;">{repo?.url} ↗</a>
              {/if}
            </div>
          </div>
          <div style="display: flex; gap: 16px;">
            <button class="secondary" style="padding: 10px; border-radius: 12px; color: var(--efinity-blue); border-color: rgba(0, 153, 255, 0.2);" onclick={exportRepo} data-tooltip="Recovery Bridge (One-Click Export)" aria-label='Export Repository'>
              <Icon name="download" />
            </button>
            <button class="secondary" style="padding: 10px; border-radius: 12px; color: var(--status-green); border-color: rgba(0, 255, 136, 0.2);" onclick={syncNow} data-tooltip="Sync Now" aria-label='Sync now'>
              <Icon name="refresh" />
            </button>
            <button class="secondary" style="padding: 10px; border-radius: 12px;" onclick={() => showConfigModal = true} data-tooltip="Settings" aria-label='Settings'>
              <Icon name="settings" />
            </button>
            <button class="secondary" style="padding: 10px; border-radius: 12px;" onclick={() => selectedRepo = null} data-tooltip="Close" aria-label='Close'>
              <Icon name="x" strokeWidth={3} />
            </button>
          </div>
        </div>
        <div style="display: flex; justify-content: space-between; align-items: center; color: var(--efinity-text-muted); margin-top: 1rem; font-size: .8rem;">
          <div style="display: flex; flex-direction: column; align-items: flex-start;">Syncing in <b style="color: var(--efinity-text-main); font-size: .9rem;">{getRemainingTime(repo, false, now)}</b></div>
          <div style="display: flex; flex-direction: column; align-items: flex-end;">Interval <b style="color: var(--efinity-text-main); font-size: .9rem;">{minutesToHuman(repo?.interval_minutes)}</b></div>
        </div>
      </div>

      <div class="modal-body">
        <div class="tab-container">
          <button class="tab-btn" class:active={activeTab === 'readme'} onclick={() => activeTab = 'readme'}>BRIEFING</button>
          <button class="tab-btn" class:active={activeTab === 'issues'} onclick={() => activeTab = 'issues'}>
            INTEL <span class="tab-count">{issues.length}</span>
          </button>
          <button class="tab-btn" class:active={activeTab === 'releases'} onclick={() => activeTab = 'releases'}>
            CHRONICLE <span class="tab-count">{releases.length}</span>
          </button>
          <button class="tab-btn" class:active={activeTab === 'wiki'} onclick={() => activeTab = 'wiki'}>WIKI</button>
          <button class="tab-btn" class:active={activeTab === 'logs'} onclick={() => activeTab = 'logs'}>LOGS</button>
        </div>

        {#if activeTab === 'readme'}
          <RepoReadme {readmeContent} />
        {:else if activeTab === 'issues'}
          <RepoIssues {issues} />
        {:else if activeTab === 'releases'}
          <RepoReleases {releases} />
        {:else if activeTab === 'wiki'}
          <RepoWiki {wikiContent} />
        {:else if activeTab === 'logs'}
          <RepoLogs commits={parseCommits(repo?.last_commit || '')} />
        {/if}
      </div>
    </div>
  {/if}
</div>
