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
  }

  let repositories = $state<Repository[]>([]);
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
      fetchRepos();
    }
  }

  onMount(() => {
    fetchRepos();
    const ws = new WebSocket('ws://localhost:8080/ws');
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'status_update') {
        const repo = repositories.find(r => r.id === data.id);
        if (repo) {
          repo.status = data.status;
          repo.error_message = data.error;
          if (data.status === 'synced') fetchRepos();
        }
      }
    };
  });
</script>

<div class="container">
  <h1>GitPatrol ⚡</h1>

  <div class="card" style="border-style: dashed; border-color: #444; margin-bottom: 40px;">
    <h3>Nieuwe Repo Toevoegen</h3>
    <div style="display: flex; gap: 8px;">
      <input bind:value={newName} placeholder="Naam (bv. opengem)" style="flex: 1;" />
      <input bind:value={newUrl} placeholder="GitHub URL" style="flex: 2;" />
      <input bind:value={interval} type="number" placeholder="Minuten" style="width: 80px;" />
      <button onclick={addRepo}>ADD</button>
    </div>
  </div>

  {#if repositories.length === 0}
    <div style="text-align: center; color: #666; padding: 40px;">
      Nog geen repositories toegevoegd. Start met je eerste!
    </div>
  {:else}
    {#each repositories as repo (repo.id)}
      <div class="card">
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <div>
            <h3 style="margin: 0; display: flex; align-items: center; gap: 12px;">
              {repo.name}
              <span class="status-badge status-{repo.status}">{repo.status}</span>
            </h3>
            <p style="color: #666; margin: 4px 0 0 0; font-size: 14px;">{repo.url}</p>
          </div>
          <div style="text-align: right;">
            <div style="font-size: 12px; color: #444;">LAST SYNC</div>
            <div style="font-size: 14px; font-weight: 600;">{repo.last_sync ? new Date(repo.last_sync).toLocaleString() : 'Never'}</div>
          </div>
        </div>
        
        {#if repo.last_commit}
          <div style="margin-top: 16px; padding: 12px; background: #111; border-radius: 6px; font-family: monospace; font-size: 13px; color: var(--efinity-blue);">
            {repo.last_commit}
          </div>
        {/if}

        {#if repo.error_message}
          <div style="margin-top: 16px; padding: 12px; background: rgba(255,0,0,0.1); border-radius: 6px; color: #ff0000; font-size: 13px;">
            {repo.error_message}
          </div>
        {/if}
      </div>
    {/each}
  {/if}
</div>
