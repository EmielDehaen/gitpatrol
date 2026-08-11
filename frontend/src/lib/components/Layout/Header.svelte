<script lang="ts">
  import { healthStore } from '$lib/health.svelte';
  import { incidentsStore } from '$lib/incidents.svelte';
  import Icon from '$lib/components/Icon.svelte';

  let { viewMode = $bindable(), showIncidentModal = $bindable(), searchQuery = $bindable() } = $props();
</script>

<header class="top-bar">
  <div class="search-bar">
    <Icon name="activity" />
    <input type="text" placeholder="Search clusters..." bind:value={searchQuery} />
  </div>

  <div class="top-bar-actions">
    <div class="view-toggle">
      <button class:active={viewMode === 'grid'} onclick={() => viewMode = 'grid'}>GRID</button>
      <button class:active={viewMode === 'list'} onclick={() => viewMode = 'list'}>LIST</button>
    </div>
    
    <button 
      class="notification-bell" 
      class:has-incidents={incidentsStore.incidents.length > 0} 
      onclick={() => showIncidentModal = true} 
      data-tooltip="Security Logs"
    >
      <Icon name="bell" />
      {#if incidentsStore.incidents.length > 0}
        <div class="bell-count">{incidentsStore.incidents.length}</div>
      {/if}
    </button>

    <div 
      class="badge status-{healthStore.healthStatus?.status || 'offline'}" 
      data-tooltip={healthStore.healthStatus ? `Internet: ${healthStore.healthStatus.checks.internet.connected ? 'OK' : 'OFF'} | Disk: ${healthStore.healthStatus.checks.disk.used_percent}` : 'Checking...'}
    >
      <span class="dot"></span>
      {healthStore.healthStatus ? (healthStore.healthStatus.status === 'healthy' ? 'SYSTEM ONLINE' : healthStore.healthStatus.status.toUpperCase()) : 'OFFLINE'}
    </div>
  </div>
</header>

<style>
  .top-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 24px 40px;
    margin-bottom: 24px;
  }
  
  .search-bar {
    display: flex;
    align-items: center;
    background: var(--surface-container-low);
    border-radius: 8px;
    padding: 0 16px;
    width: 300px;
    color: var(--on-surface-variant);
  }
  
  .search-bar input {
    background: transparent;
    border: none;
    margin: 0;
    padding: 12px;
    box-shadow: none;
    width: 100%;
    color: var(--on-surface);
  }

  .search-bar input:focus {
    box-shadow: none;
  }

  .top-bar-actions {
    display: flex;
    gap: 16px;
    align-items: center;
  }
</style>
