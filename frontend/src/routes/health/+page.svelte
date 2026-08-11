<script lang="ts">
  import { fade } from 'svelte/transition';
  import { healthStore } from '$lib/health.svelte';

  // Helper for byte conversion
  function formatBytes(bytes: number, decimals = 2) {
    if (!+bytes) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
  }
</script>

<div class="page-container" transition:fade>
  <div class="header">
    <h1 class="display-lg">System Health</h1>
    
    {#if healthStore.healthStatus}
      <div class="status-chip status-{healthStore.healthStatus.status}">
        <span class="dot"></span>
        {healthStore.healthStatus.status.toUpperCase()}
      </div>
    {:else}
      <div class="status-chip status-offline">
        <span class="dot"></span>
        CHECKING...
      </div>
    {/if}
  </div>

  <div class="metrics-grid">
    {#if healthStore.healthStatus}
      <!-- Network / Internet -->
      <div class="metric-card">
        <h3 class="metric-title">NETWORK LINK</h3>
        <div class="metric-value" style="color: {healthStore.healthStatus.checks.internet.connected ? 'var(--primary)' : 'var(--error)'};">
          {healthStore.healthStatus.checks.internet.connected ? 'ESTABLISHED' : 'DISCONNECTED'}
        </div>
        <div class="metric-sub">
          Latency: {healthStore.healthStatus.checks.internet.latency}
        </div>
      </div>

      <!-- Storage -->
      <div class="metric-card">
        <h3 class="metric-title">STORAGE LOAD</h3>
        <div class="metric-value">
          {healthStore.healthStatus.checks.disk.used_percent}
        </div>
        <div class="metric-sub">
          {formatBytes(healthStore.healthStatus.checks.disk.free_bytes)} free of {formatBytes(healthStore.healthStatus.checks.disk.total_bytes)}
        </div>
        <!-- Mini progress bar -->
        <div class="progress-bar-bg" style="margin-top: 16px;">
          <div class="progress-bar-fill" style="width: {healthStore.healthStatus.checks.disk.used_percent};"></div>
        </div>
      </div>

      <!-- Database -->
      <div class="metric-card">
        <h3 class="metric-title">DATABASE CLUSTER</h3>
        <div class="metric-value" style="color: {healthStore.healthStatus.checks.database ? 'var(--primary)' : 'var(--error)'};">
          {healthStore.healthStatus.checks.database ? 'ONLINE' : 'UNREACHABLE'}
        </div>
        <div class="metric-sub">
          SQLite Core Engine
        </div>
      </div>

      <!-- Workers -->
      <div class="metric-card">
        <h3 class="metric-title">SYNC WORKERS</h3>
        <div class="metric-value">
          {healthStore.healthStatus.checks.workers.active_tasks} / {healthStore.healthStatus.checks.workers.total_workers}
        </div>
        <div class="metric-sub">
          {healthStore.healthStatus.checks.workers.queued_tasks} tasks queued
        </div>
      </div>
    {:else}
      <div class="metric-card" style="grid-column: span 2; text-align: center;">
        <p style="color: var(--on-surface-variant);">Awaiting telemetry data...</p>
      </div>
    {/if}
  </div>
</div>

<style>
  .header {
    padding: 40px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .status-chip {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    border-radius: 999px;
    font-size: 0.8rem;
    font-weight: 700;
    letter-spacing: 0.1em;
  }
  
  .status-healthy { color: var(--primary); background: rgba(77, 221, 187, 0.1); }
  .status-degraded { color: var(--warning, #ffcc00); background: rgba(255, 204, 0, 0.1); }
  .status-critical { color: var(--error); background: var(--error-container); }
  .status-offline { color: var(--on-surface-variant); background: var(--surface-container-highest); }
  
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: currentColor;
    box-shadow: 0 0 8px currentColor;
  }

  .metrics-grid {
    padding: 0 40px 80px 40px;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 24px;
    max-width: 1200px;
  }

  .metric-card {
    background: var(--surface-container-low);
    border-radius: 24px;
    padding: 32px;
    display: flex;
    flex-direction: column;
  }

  .metric-title {
    font-size: 0.7rem;
    font-weight: 800;
    letter-spacing: 0.1em;
    color: var(--on-surface-variant);
    margin: 0 0 16px 0;
  }

  .metric-value {
    font-size: 2.5rem;
    font-weight: 700;
    font-family: 'Space Grotesk', sans-serif;
    color: var(--on-surface);
    line-height: 1;
    margin-bottom: 8px;
  }

  .metric-sub {
    font-size: 0.9rem;
    color: var(--on-surface-variant);
  }

  .progress-bar-bg {
    height: 6px;
    background: var(--surface-container-highest);
    border-radius: 3px;
    overflow: hidden;
  }

  .progress-bar-fill {
    height: 100%;
    background: var(--primary);
    border-radius: 3px;
    box-shadow: 0 0 12px rgba(77, 221, 187, 0.4);
  }
</style>
