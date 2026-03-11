<script lang="ts">
  import { onMount } from 'svelte';
  import { marked } from 'marked';
  import { type Repository, type Commit } from '$lib/types';
  import { API_URL, api } from '$lib/api.svelte';
  import { fade } from 'svelte/transition';

  let { repo = $bindable(), selectedRepo = $bindable(), showConfigModal = $bindable() } = $props<{ repo: Repository, selectedRepo: Repository | null, showConfigModal: boolean }>();

  let activeTab = $state<'readme' | 'issues' | 'releases' | 'wiki' | 'logs'>('readme');
  let readmeContent = $state('Loading mission briefing...');
  let wikiContent = $state('');
  let issues = $state<any[]>([]);
  let releases = $state<any[]>([]);
  let readmeExpanded = $state(false);

  async function fetchMetadata() {
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
    readmeContent = 'Loading mission briefing...';
    readmeExpanded = false;
    try {
      const res = await api.apiFetch(`/api/repositories/${repo.id}/readme`);
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
    try {
      await api.apiFetch(`/api/repositories/${repo.id}/sync`, { method: 'POST' });
      api.showToast('Sync initiated.', 'info');
    } catch (e) {}
  }

  $effect(() => {
    if (repo.id) fetchReadme();
  });
</script>

<div class="modal-overlay" onclick={() => selectedRepo = null} transition:fade>
  <div class="modal-content" onclick={(e) => e.stopPropagation()}>
    <div class="modal-header">
      <div style="display: flex; justify-content: space-between; align-items: flex-start;">
        <div>
          <div style="display: flex; align-items: center; gap: 16px; margin-bottom: 12px;">
            <h2 style="font-size: 2.5rem; margin: 0; letter-spacing: -0.04em;">{repo.name}</h2>
            <div class="health-score" style="border-color: {repo.health_score > 70 ? 'var(--status-green)' : repo.health_score > 40 ? 'var(--status-yellow)' : 'var(--status-red)'}">
              {repo.health_score}%
            </div>
          </div>
          <p style="color: var(--efinity-text-muted); font-size: 0.9rem; font-weight: 600; font-family: monospace;">{repo.url}</p>
        </div>
        <div style="display: flex; gap: 16px;">
          <button class="secondary" onclick={syncNow}>FORCE SYNC</button>
          <button class="secondary" onclick={() => showConfigModal = true}>CONFIGURE</button>
          <button class="secondary" onclick={() => selectedRepo = null} style="padding: 16px;">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
          </button>
        </div>
      </div>
    </div>

    <div class="modal-body">
      <div class="tab-container">
        <button class="tab-btn" class:active={activeTab === 'readme'} onclick={() => activeTab = 'readme'}>Briefing</button>
        <button class="tab-btn" class:active={activeTab === 'issues'} onclick={() => activeTab = 'issues'}>
          Issues <span class="tab-count">{issues.length}</span>
        </button>
        <button class="tab-btn" class:active={activeTab === 'releases'} onclick={() => activeTab = 'releases'}>
          Releases <span class="tab-count">{releases.length}</span>
        </button>
        <button class="tab-btn" class:active={activeTab === 'wiki'} onclick={() => activeTab = 'wiki'}>Intelligence</button>
        <button class="tab-btn" class:active={activeTab === 'logs'} onclick={() => activeTab = 'logs'}>Tactical Log</button>
      </div>

      {#if activeTab === 'readme'}
        <div class="readme-container" class:readme-expanded={readmeExpanded}>
          <div class="readme-content">
            {@html readmeContent}
          </div>
          {#if !readmeExpanded}
            <div class="readme-fade">
              <button class="secondary" onclick={() => readmeExpanded = true}>READ FULL BRIEFING</button>
            </div>
          {/if}
        </div>
      {:else if activeTab === 'issues'}
        <div class="repo-list">
          {#each issues as issue}
            <div class="list-item" style="cursor: default; opacity: {issue.state === 'closed' ? 0.6 : 1}">
              <div style="display: flex; gap: 16px; align-items: flex-start;">
                <div style="color: {issue.state === 'open' ? 'var(--status-green)' : 'var(--status-red)'}; margin-top: 4px;">
                  {#if issue.state === 'open'}
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><circle cx="12" cy="12" r="1"></circle></svg>
                  {:else}
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><polyline points="12 8 12 12 15 15"></polyline></svg>
                  {/if}
                </div>
                <div>
                  <div style="font-weight: 800; font-size: 1rem; color: #fff;">{issue.title}</div>
                  <div style="font-size: 0.7rem; color: var(--efinity-text-muted); margin-top: 4px; font-weight: 700;">#{issue.number} opened by {issue.user.login}</div>
                </div>
              </div>
              <div style="display: flex; gap: 8px;">
                {#each issue.labels as label}
                  <span class="branch-badge" style="background: rgba(255,255,255,0.05); color: #fff; border: 1px solid var(--glass-border);">
                    {label.name}
                  </span>
                {/each}
              </div>
            </div>
          {:else}
            <p style="color: var(--efinity-text-muted); text-align: center; padding: 40px;">No issues found in this sector.</p>
          {/each}
        </div>
      {:else if activeTab === 'releases'}
        <div class="repo-list">
          {#each releases as release}
            <div class="list-item" style="cursor: default;">
              <div style="display: flex; gap: 20px; align-items: center;">
                <div style="color: var(--efinity-blue);">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path><polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline><line x1="12" y1="22.08" x2="12" y2="12"></line></svg>
                </div>
                <div>
                  <div style="font-weight: 800; font-size: 1.1rem; color: #fff;">{release.name || release.tag_name}</div>
                  <div style="font-size: 0.75rem; color: var(--efinity-text-muted); margin-top: 4px; font-weight: 700;">Deployed on {new Date(release.published_at).toLocaleDateString()}</div>
                </div>
              </div>
              <span class="branch-badge" style="font-size: 0.8rem; padding: 6px 12px; background: var(--efinity-blue); color: #fff;">{release.tag_name}</span>
            </div>
          {:else}
            <p style="color: var(--efinity-text-muted); text-align: center; padding: 40px;">No official releases detected.</p>
          {/each}
        </div>
      {:else if activeTab === 'wiki'}
        <div class="readme-container readme-expanded">
          <div class="readme-content">
            {@html wikiContent}
          </div>
        </div>
      {:else if activeTab === 'logs'}
        <div class="repo-list">
          {#each parseCommits(repo.last_commit) as commit}
            <div class="list-item" style="cursor: default;">
              <div style="display: flex; gap: 20px; align-items: center; flex: 1;">
                <div style="font-family: monospace; color: var(--efinity-blue); font-weight: 800; font-size: 0.8rem; background: rgba(0, 112, 243, 0.05); padding: 4px 8px; border-radius: 6px;">
                  {commit.hash.substring(0, 7)}
                </div>
                <div style="flex: 1;">
                  <div style="font-weight: 800; font-size: 1rem; color: #fff;">{commit.message}</div>
                  <div style="font-size: 0.7rem; color: var(--efinity-text-muted); margin-top: 4px; font-weight: 700;">
                    {commit.author} • {commit.date}
                  </div>
                </div>
              </div>
              {#if commit.branch}
                <span class="branch-badge">{commit.branch}</span>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>
</div>
