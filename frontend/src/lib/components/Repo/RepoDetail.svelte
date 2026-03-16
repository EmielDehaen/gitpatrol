<script lang="ts">
  import { onMount } from 'svelte';
  import { marked } from 'marked';
  import { type Repository, type Commit } from '$lib/types';
  import { API_URL, api } from '$lib/api.svelte';
  import { fade } from 'svelte/transition';
  import { getAvatarUrl, getRemainingTime, handleAvatarError, minutesToHuman } from '$lib/utils';

  let { repo = $bindable(), selectedRepo = $bindable(), showConfigModal = $bindable() } = $props<{ repo: Repository | null, selectedRepo: Repository | null, showConfigModal: boolean }>();

  let activeTab = $state<'readme' | 'issues' | 'releases' | 'wiki' | 'logs'>('readme');
  let readmeContent = $state('Loading mission briefing...');
  let wikiContent = $state('');
  let issues = $state<any[]>([]);
  let releases = $state<any[]>([]);
  let readmeExpanded = $state(false);

  async function fetchMetadata() {
    if (!repo) return;
    const assetBase = `/api/repositories/${repo.id}/assets/`;
    try {
      const issuesRes = await api.apiFetch(`${assetBase}metadata/issues.json`);
      issues = issuesRes.ok ? await issuesRes.json() : [];

      const releasesRes = await api.apiFetch(`${assetBase}metadata/releases.json`);
      releases = releasesRes.ok ? await releasesRes.json() : [];

      const wikiRes = await api.apiFetch(`${assetBase}wiki/Home.md`);
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
      const res = await api.apiFetch(`/api/repositories/${repo.id}/readme`);
      if (res.ok) {
        let text = await res.text();
        const assetBase = `${API_URL}/api/repositories/${repo.id}/assets/`;
        
        text = text.replace(/!\[([^\]]*)\]\((?!(?:http|https|ftp|data:))(?:\.\/)?([^)]+)\)/gi, `![$1](${assetBase}$2)`);
        text = text.replace(/<img([^+]+)src=["'](?!(?:http|https|ftp))(?:\.\/)?([^"']+)["']/gi, (match, pre, path) => `<img${pre}src="${assetBase}${path}"`);
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
      await api.apiFetch(`/api/repositories/${repo.id}/sync`, { method: 'POST' });
      api.showToast('Sync initiated.', 'info');
    } catch (e) {}
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
                <!-- <div class="health-score" style="border-color: {(repo?.health_score || 0) > 70 ? 'var(--status-green)' : (repo?.health_score || 0) > 40 ? 'var(--status-yellow)' : 'var(--status-red)'}">
                  {repo?.health_score || 0}%
                </div> -->
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
          <div style="display: flex; flex-direction: column; align-items: flex-start;">Syncing in <b style="color: var(--efinity-text-main); font-size: .9rem;">{getRemainingTime(repo, false)}</b></div>
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
