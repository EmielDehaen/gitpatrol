<script lang="ts">
  import { api } from '$lib/api.svelte';

  let { viewMode = $bindable(), showIncidentModal = $bindable(), showUserModal = $bindable() } = $props();
</script>

<header>
  <div>
    <h1>GitPatrol ⚡</h1>
    <p style="color: var(--efinity-text-muted); margin: 8px 0 0 0; font-weight: 600; text-transform: uppercase; font-size: 0.7rem; letter-spacing: 0.1em;">Tactical Asset Monitoring</p>
  </div>
  <div style="display: flex; gap: 20px; align-items: center;">
    <div class="view-toggle">
      <button class:active={viewMode === 'grid'} onclick={() => viewMode = 'grid'}>GRID</button>
      <button class:active={viewMode === 'list'} onclick={() => viewMode = 'list'}>LIST</button>
    </div>
    <button 
      class="notification-bell" 
      class:has-incidents={api.incidents.length > 0} 
      onclick={() => showIncidentModal = true} 
      data-tooltip="Security Logs"
    >
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
      {#if api.incidents.length > 0}
        <div class="bell-count">{api.incidents.length}</div>
      {/if}
    </button>
    <div class="badge" style="color: var(--status-green); background: rgba(0, 255, 136, 0.05);">
      <span style="width: 6px; height: 6px; background: var(--status-green); border-radius: 50%;"></span>
      SYSTEM ONLINE
    </div>
    <div class="user-profile" onclick={() => showUserModal = true} data-tooltip="User Settings" role='button' onkeypress={() => {}} tabindex=0>
      <div class="avatar-circle">{api.user?.username ? api.user.username.charAt(0) : 'U'}</div>
      <div class="user-info">
        <span class="user-name">{api.user?.username}</span>
        <span class="user-role">Administrator</span>
      </div>
    </div>
    <button class="notification-bell logout-btn" onclick={() => api.handleLogout()} data-tooltip="Logout" aria-label='logout'>
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M18.36 6.64a9 9 0 1 1-12.73 0"></path><line x1="12" y1="2" x2="12" y2="12"></line></svg>
    </button>
  </div>
</header>
