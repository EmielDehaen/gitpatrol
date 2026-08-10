<script lang="ts">
  import { onMount } from 'svelte';
  import { marked } from 'marked';
  
  // Custom extension for GitHub Alerts ([!TIP], [!NOTE], etc.)
  marked.use({
    extensions: [{
      name: 'alert',
      level: 'block',
      start(src) { return src.match(/^> \[!/)?.index; },
      tokenizer(src) {
        const rule = /^> \[!(TIP|NOTE|IMPORTANT|WARNING|CAUTION)\][ \t]*\n((?:> .*(?:\n|$))*)/;
        const match = rule.exec(src);
        if (match) {
          return {
            type: 'alert',
            raw: match[0],
            alertType: match[1].toLowerCase(),
            text: match[2].replace(/^> /gm, '').trim()
          };
        }
      },
      renderer(token) {
        const icons = {
          tip: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px"><path d="M15 14c.2-1 .7-1.7 1.5-2.5 1-.9 1.5-2.2 1.5-3.5A5 5 0 0 0 8 8c0 1.3.5 2.6 1.5 3.5.8.8 1.3 1.5 1.5 2.5"/><path d="M9 18h6"/><path d="M10 22h4"/></svg>',
          note: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>',
          warning: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>',
          important: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>',
          caution: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>'
        };
        return `<div class="markdown-alert markdown-alert-${token.alertType}">
          <p class="markdown-alert-title">${icons[token.alertType] || icons.note}${token.alertType.toUpperCase()}</p>
          <div class="markdown-alert-content">${marked.parse(token.text)}</div>
        </div>`;
      }
    }]
  });

  import { type Repository, type Commit } from '$lib/types';
  import { API_URL, api, apiFetch, toastHandler } from '$lib/api.svelte';
  import { fade } from 'svelte/transition';
  import { getAvatarUrl, getRemainingTime, handleAvatarError, minutesToHuman } from '$lib/utils';

  let { repo = $bindable(), selectedRepo = $bindable(), showConfigModal = $bindable() } = $props<{ repo: Repository | null, selectedRepo: Repository | null, showConfigModal: boolean }>();

  let activeTab = $state<'readme' | 'issues' | 'releases' | 'wiki' | 'logs'>('readme');
  let readmeContent = $state('Loading mission briefing...');
  let wikiContent = $state('');
  let issues = $state<any[]>([]);
  let releases = $state<any[]>([]);
  let readmeExpanded = $state(false);
  let now = $state(Date.now());
  let exportDestination = $state('');

  async function fetchSettings() {
    const settings = await api.settings.getSettings();
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
        wikiContent = await marked.parse(await wikiRes.text());
      } else {
        wikiContent = '<p style="color: var(--efinity-text-muted)">No documentation (Wiki) found for this asset.</p>';
      }
    } catch (e) {}
  }

  async function fetchReadme() {
    if (!repo) return;
    readmeContent = 'Loading mission briefing...';
    readmeExpanded = false;
    try {
      const res = await apiFetch(`/api/repositories/${repo.id}/readme`);
      if (res.ok) {
        let text = await res.text();
        const assetBase = `${API_URL}/api/repositories/${repo.id}/assets/`;
        
        text = text.replace(/!\[([^\]]*)\]\((?!(?:http|https|ftp|data:))(?:\.\/)?([^)]+)\)/gi, `![$1](${assetBase}$2)`);
        text = text.replace(/<img([^>]*?)src=["'](?!(?:http|https|ftp))(?:\.\/)?([^"']+)["']/gi, (match, pre, path) => `<img${pre}src="${assetBase}${path}"`);
        text = text.replace(/<source([^>]*?)srcset=["'](?!(?:http|https|ftp))(?:\.\/)?([^"']+)["']/gi, (match, pre, path) => `<source${pre}srcset="${assetBase}${path}"`);
        text = text.replace(/\[([^\]]*)\]\((?!(?:http|https|ftp|#))(?:\.\/)?([^)]+)\)/gi, `[$1](${assetBase}$2)`);

        readmeContent = await marked.parse(text, { gfm: true, breaks: true });
      } else {
        readmeContent = '<p style="color: var(--efinity-text-muted)">No mission briefing available for this asset.</p>';
      }
      fetchMetadata();
    } catch (e) {}
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
    } catch (e) {}
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
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3v12m0 0l-4-4m4 4l4-4M4 17v2a2 2 0 002 2h12a2 2 0 002-2v-2"/></svg>
            </button>
            <button class="secondary" style="padding: 10px; border-radius: 12px; color: var(--status-green); border-color: rgba(0, 255, 136, 0.2);" onclick={syncNow} data-tooltip="Sync Now" aria-label='Sync now'>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21.5 2v6h-6M2.5 22v-6h6M2 12c0-4.4 3.6-8 8-8 3.3 0 6.1 2 7.3 4.9M22 12c0 4.4-3.6 8-8 8-3.3 0-6.1-2-7.3-4.9"/></svg>
            </button>
            <button class="secondary" style="padding: 10px; border-radius: 12px;" onclick={() => showConfigModal = true} data-tooltip="Settings" aria-label='Settings'>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
            </button>
            <button class="secondary" style="padding: 10px; border-radius: 12px;" onclick={() => selectedRepo = null} data-tooltip="Close" aria-label='Close'>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
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
          <div class="readme-container" class:readme-expanded={readmeExpanded}>
            <div class="readme-content">
              {@html readmeContent}
            </div>
            {#if !readmeExpanded}
              <div class="readme-fade">
                <button class="secondary" style="font-size: 0.65rem; padding: 12px 24px;" onclick={() => readmeExpanded = true}>READ FULL MISSION BRIEFING</button>
              </div>
            {/if}
          </div>
        {:else if activeTab === 'issues'}
          <div style="display: flex; flex-direction: column; gap: 16px;">
            {#each issues as issue}
              <div style="background: rgba(255,255,255,0.02); border-radius: 16px; border: 1px solid var(--glass-border); padding: 20px; opacity: {issue.state === 'closed' ? 0.6 : 1}">
                <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 8px;">
                  <div style="font-weight: 700; font-size: 1.1rem; color: #fff;">{issue.title}</div>
                  <span class="badge" style="color: {issue.state === 'open' ? 'var(--status-green)' : 'var(--efinity-text-muted)'}; background: {issue.state === 'open' ? 'rgba(0, 255, 136, 0.05)' : 'rgba(255,255,255,0.03)'}">{issue.state.toUpperCase()}</span>
                </div>
                <div style="font-size: 0.8rem; color: var(--efinity-text-muted);">#{issue.number} opened by {issue.user?.login} • {new Date(issue.created_at).toLocaleDateString()}</div>
              </div>
            {:else}
              <p style="color: var(--efinity-text-muted); text-align: center; padding: 40px;">No tactical issues found in this sector.</p>
            {/each}
          </div>
        {:else if activeTab === 'releases'}
          <div style="display: flex; flex-direction: column; gap: 24px;">
            {#each releases as release}
              <div style="background: rgba(255,255,255,0.02); border-radius: 20px; border: 1px solid var(--glass-border); padding: 24px;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; border-bottom: 1px solid var(--glass-border); padding-bottom: 16px;">
                  <div>
                    <div style="font-size: 1.4rem; font-weight: 800; color: var(--efinity-blue);">{release.tag_name}</div>
                    <div style="font-size: 0.8rem; color: var(--efinity-text-muted);">{release.name || ''} • {new Date(release.published_at).toLocaleDateString()}</div>
                  </div>
                </div>
                <div class="readme-content" style="font-size: 0.9rem;">{@html marked.parse(release.body || '')}</div>
              </div>
            {:else}
              <p style="color: var(--efinity-text-muted); text-align: center; padding: 40px;">No historical chronicles (releases) found.</p>
            {/each}
          </div>
        {:else if activeTab === 'wiki'}
          <div class="readme-container readme-expanded">
            <div class="readme-content">
              {@html wikiContent}
            </div>
          </div>
        {:else if activeTab === 'logs'}
          <div style="display: flex; flex-direction: column; gap: 12px;">
            {#each parseCommits(repo?.last_commit || '') as commit}
              <div style="background: rgba(255,255,255,0.02); border-radius: 16px; border: 1px solid var(--glass-border); padding: 20px;">
                <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 8px;">
                  {#if commit.branch}<span class="branch-badge">{commit.branch}</span>{/if}
                  <span style="font-weight: 600; font-size: 1rem;">{commit.message}</span>
                </div>
                <div style="font-size: 0.8rem; color: var(--efinity-text-muted); font-weight: 500;">
                  {commit.author} • {commit.date} • <span style="color: var(--efinity-blue); font-family: monospace;">{commit.hash.substring(0,7)}</span>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>
