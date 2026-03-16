<script lang="ts">
  import { api } from '$lib/api.svelte';
  import { fade } from 'svelte/transition';

  let username = $state('');
  let password = $state('');
  let loading = $state(false);

  async function handleSubmit() {
    loading = true;
    const success = await api.handleAuth(username, password);
    if (success) {
      username = '';
      password = '';
    }
    loading = false;
  }
</script>

<div class="modal-overlay" transition:fade>
  <div class="modal-content" style="max-width: 400px; padding: 40px;">
    <div style="text-align: center; margin-bottom: 40px;">
      <h1 style="font-size: 2.5rem; margin-bottom: 8px;">GitPatrol ⚡</h1>
      <p style="color: var(--efinity-text-muted); font-weight: 700; text-transform: uppercase; font-size: 0.7rem; letter-spacing: 0.1em;">
        {api.needsBootstrap ? 'System Initialization Required' : 'Authentication Required'}
      </p>
    </div>
    
    <label for='username'>USERNAME</label>
    <input id='username' type="text" bind:value={username} placeholder="username" />
    
    <label for='password'>PASSWORD</label>
    <input id='password' type="password" bind:value={password} placeholder="" onkeydown={(e) => e.key === 'Enter' && handleSubmit()} />
    
    <button style="width: 100%; margin-top: 32px;" onclick={handleSubmit} disabled={loading}>
      {loading ? 'PROCESSING...' : api.needsBootstrap ? 'BOOTSTRAP SYSTEM' : 'ACCESS DASHBOARD'}
    </button>

    {#if !api.needsBootstrap}
      <p style="text-align: center; font-size: 0.75rem; color: var(--efinity-text-muted); margin-top: 24px;">
        Contact your administrator for access.
      </p>
    {/if}
  </div>
</div>
