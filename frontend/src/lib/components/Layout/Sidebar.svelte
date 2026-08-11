<script lang="ts">
  import { page } from '$app/stores';
  import Icon from '$lib/components/Icon.svelte';
  import { authStore } from '$lib/auth.svelte';
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
      <span>Overview</span>
    </a>
    <a href="/repositories" class="nav-item" class:active={$page.url.pathname.startsWith('/repositories')}>
      <Icon name="folder" />
      <span>Repositories</span>
    </a>
    <a href="/mirroring" class="nav-item" class:active={$page.url.pathname.startsWith('/mirroring')}>
      <Icon name="git-merge" />
      <span>Mirroring</span>
    </a>
    <a href="/health" class="nav-item" class:active={$page.url.pathname.startsWith('/health')}>
      <Icon name="activity" />
      <span>Health</span>
    </a>
    <a href="/settings" class="nav-item" class:active={$page.url.pathname.startsWith('/settings')}>
      <Icon name="settings" />
      <span>Settings</span>
    </a>
  </nav>

  <div class="sidebar-footer">
    <div class="user-card-wrapper">
      <div class="user-card">
        <div class="avatar-circle">{authStore.user?.username ? authStore.user.username.charAt(0) : 'U'}</div>
        <div class="user-info">
          <span class="user-name">{authStore.user?.username || 'Operator'}</span>
          <span class="user-role">Tier: Enterprise</span>
        </div>
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

  .user-card-wrapper {
    display: flex;
    align-items: center;
    gap: 8px;
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

  .logout-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 44px;
    height: 44px;
    background: var(--surface-container-low);
    border: 1px solid var(--glass-border);
    border-radius: 12px;
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
