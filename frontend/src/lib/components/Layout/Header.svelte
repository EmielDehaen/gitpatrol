<script lang="ts">
  import { api } from '$lib/api.svelte';
  import Icon from '$lib/components/Icon.svelte';

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
      <Icon name="bell" />
      {#if api.incidents.length > 0}
        <div class="bell-count">{api.incidents.length}</div>
      {/if}
    </button>
    <div 
      class="badge status-{api.healthStatus?.status || 'offline'}" 
      data-tooltip={api.healthStatus ? `Internet: ${api.healthStatus.checks.internet.connected ? 'OK' : 'OFF'} | Disk: ${api.healthStatus.checks.disk.used_percent}` : 'Checking...'}
    >
      <span class="dot"></span>
      {api.healthStatus ? (api.healthStatus.status === 'healthy' ? 'SYSTEM ONLINE' : api.healthStatus.status.toUpperCase()) : 'OFFLINE'}
    </div>
    <div class="user-profile" onclick={() => showUserModal = true} data-tooltip="User Settings" role='button' onkeypress={() => {}} tabindex=0>
      <div class="avatar-circle">{api.user?.username ? api.user.username.charAt(0) : 'U'}</div>
      <div class="user-info">
        <span class="user-name">{api.user?.username}</span>
        <span class="user-role">Administrator</span>
      </div>
    </div>
    <button class="notification-bell logout-btn" onclick={() => api.handleLogout()} data-tooltip="Logout" aria-label='logout'>
      <Icon name="power" />
    </button>
  </div>
</header>
