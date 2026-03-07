<script lang="ts">
  import { onMount } from 'svelte';
  import { marked } from 'marked';

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
    progress?: number;
  }

  let repositories = $state<Repository[]>([]);
  let viewMode = $state<'grid' | 'list'>('grid');
  let selectedRepo = $state<Repository | null>(null);
  let readmeContent = $state('');
  let readmeExpanded = $state(false);
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

  async function fetchReadme(repo: Repository) {
    readmeContent = 'Loading mission briefing...';
    readmeExpanded = false;
    const res = await fetch(`${API_URL}/api/repositories/${repo.id}/readme`);
    if (res.ok) {
      let text = await res.text();
      
      const parts = repo.url.replace('https://github.com/', '').split('/');
      if (parts.length >= 2) {
        const user = parts[0];
        const repoName = parts[1].replace('.git', '');
        const branch = repo.default_branch || 'main';
        const rawBase = `https://raw.githubusercontent.com/${user}/${repoName}/${branch}/`;
        
        // Robust asset mapping
        text = text.replace(/!\[([^\]]*)\]\((?!(?:http|https|ftp|data:))(?:\.\/)?([^)]+)\)/gi, `![$1](${rawBase}$2)`);
        text = text.replace(/<img[^>]+src=["'](?!(?:http|https|ftp|data:))(?:\.\/)?([^"']+)["'][^>]*>/gi, (match) => {
          return match.replace(/src=["'](?:\.\/)?([^"']+)["']/i, (srcMatch, path) => `src="${rawBase}${path}"`);
        });
        text = text.replace(/\[([^\]]*)\]\((?!(?:http|https|ftp|#))(?:\.\/)?([^)]+)\)/gi, `[$1](${rawBase}$2)`);
      }

      readmeContent = await marked.parse(text, { gfm: true, breaks: true });
    } else {
      readmeContent = '<p style="color: var(--efinity-text-muted)">No mission briefing available for this asset.</p>';
    }
  }

  $effect(() => {
    if (selectedRepo) fetchReadme(selectedRepo);
  });

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
    if (parts.length > 0) {
      return `${API_URL}/avatars/${parts[0]}.png`;
    }
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
