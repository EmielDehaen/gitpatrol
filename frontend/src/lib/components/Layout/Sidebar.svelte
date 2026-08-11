<script lang="ts">
  import { page } from '$app/stores';
  import Icon from '$lib/components/Icon.svelte';
  import { authStore } from '$lib/auth.svelte';
  import { incidentsStore } from '$lib/incidents.svelte';
  import { healthStore } from '$lib/health.svelte';
</script>

<aside class="sidebar">
  <div class="sidebar-header">
    <div class="logo">
      <span class="logo-text">GitPatrol</span>
      <Icon name="zap" />
    </div>
    <div class="logo-subtitle">SENTINEL STANDARD</div>
  </div>

  <nav class="sidebar-nav">
    <a href="/" class="nav-item" class:active={$page.url.pathname === '/'}>
      <Icon name="grid" />
      <span class="nav-label">Overview</span>
    </a>
    <a href="/health" class="nav-item" class:active={$page.url.pathname.startsWith('/health')}>
      <div style="display: flex; align-items: center; justify-content: space-between; width: 100%;">
        <div style="display: flex; align-items: center; gap: 12px;">
          <Icon name="activity" />
          <span class="nav-label">Health</span>
        </div>
        {#if incidentsStore.incidents.length > 0 || (healthStore.healthStatus && healthStore.healthStatus.status !== 'healthy')}
          <span class="pulse-dot" style="background: var(--error);"></span>
        {/if}
      </div>
    </a>
    <a href="/logs" class="nav-item" class:active={$page.url.pathname.startsWith('/logs')}>
      <Icon name="terminal" />
      <span class="nav-label">Logs</span>
    </a>
    <a href="/settings" class="nav-item" class:active={$page.url.pathname.startsWith('/settings')}>
      <Icon name="settings" />
      <span class="nav-label">Settings</span>
    </a>
  </nav>

  <div class="sidebar-footer">
    <div class="user-card">
      <div class="avatar-circle">{authStore.user?.username ? authStore.user.username.charAt(0) : 'U'}</div>
      <div class="user-info">
        <span class="user-name">{authStore.user?.username || 'Operator'}</span>
        <span class="user-role">Tier: Enterprise</span>
      </div>
      <button onclick={() => authStore.handleLogout()} class="logout-btn" title="Logout" aria-label="logout">
        <Icon name="power" size={16} />
      </button>
    </div>
  </div>
</aside>

<style>
  .sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    width: 280px;
    background: var(--surface-container-lowest);
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--glass-border);
    z-index: 100;
    transition: width 0.2s, transform 0.2s;

    @media (max-width: 1439px) {
      width: 250px;
    }

    @media (max-width: 1023px) {
      width: 80px;

      .logo-text, .logo-subtitle, .nav-label, .user-info {
        display: none;
      }
      .logo { justify-content: center; }
      .sidebar-header { padding: 32px 0; display: flex; flex-direction: column; align-items: center; }
      .nav-item { justify-content: center; padding: 16px 0; }
      .sidebar-footer { padding: 32px 16px; }
      .user-card { padding: 8px; justify-content: center; flex-direction: column; gap: 16px; background: transparent; border: none; }
      .logout-btn { background: var(--surface-container-low); border: 1px solid var(--glass-border); border-radius: 50%; }
    }

    @media (max-width: 767px) {
      top: auto;
      bottom: 0;
      width: 100%;
      height: 70px;
      flex-direction: row;
      border-right: none;
      border-top: 1px solid var(--glass-border);
      
      .sidebar-header, .sidebar-footer { display: none; }
      .sidebar-nav { flex-direction: row; justify-content: space-around; align-items: center; width: 100%; padding: 0 16px; }
      .nav-item { flex: 1; padding: 0; border-left: none; border-top: 3px solid transparent; height: 100%; display: flex; align-items: center; justify-content: center; }
      .nav-item.active { border-left: none; border-top-color: var(--primary); background: transparent; }
    }
  }

  .sidebar-header {
    padding: 32px 32px 40px 32px;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 1.5rem;
    font-weight: 800;
    font-family: 'Space Grotesk', sans-serif;
    color: var(--primary);
  }

  .logo-subtitle {
    font-size: 0.6rem;
    font-weight: 800;
    letter-spacing: 0.15em;
    color: var(--on-surface-variant);
    margin-top: 8px;
    text-transform: uppercase;
  }

  .sidebar-nav {
    display: flex;
    flex-direction: column;
    flex: 1;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px 32px;
    color: var(--on-surface-variant);
    text-decoration: none;
    font-weight: 600;
    font-size: 0.95rem;
    transition: all 0.2s;
    border-left: 3px solid transparent;

    &:hover {
      background: rgba(255, 255, 255, 0.02);
      color: var(--on-surface);
    }

    &.active {
      color: var(--primary);
      background: rgba(77, 221, 187, 0.05);
      border-left-color: var(--primary);
    }

    :global(svg) {
      width: 20px;
      height: 20px;
    }
  }

  .sidebar-footer {
    padding: 32px;
  }

  .user-card {
    display: flex;
    flex: 1;
    align-items: center;
    gap: 12px;
    background: var(--surface-container-low);
    padding: 12px;
    border-radius: 12px;
    border: 1px solid var(--glass-border);
  }

  .user-info {
    flex: 1;
  }

  .logout-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    background: transparent;
    border: none;
    border-radius: 8px;
    color: var(--on-surface-variant);
    cursor: pointer;
    transition: all 0.2s;
    padding: 0;
  }

  .logout-btn:hover {
    background: var(--error-container);
    color: var(--error);
    border-color: rgba(255, 77, 77, 0.3);
  }
</style>
