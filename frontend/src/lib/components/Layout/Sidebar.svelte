<script lang="ts">
  import Icon from '$lib/components/Icon.svelte';
  import { authStore } from '$lib/auth.svelte';

  export let activePath = '/';
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
    <a href="/" class="nav-item" class:active={activePath === '/'}>
      <Icon name="grid" />
      <span>Overview</span>
    </a>
    <a href="/" class="nav-item">
      <Icon name="folder" />
      <span>Repositories</span>
    </a>
    <a href="/" class="nav-item">
      <Icon name="git-merge" />
      <span>Mirroring</span>
    </a>
    <a href="/" class="nav-item">
      <Icon name="activity" />
      <span>Health</span>
    </a>
    <a href="/" class="nav-item">
      <Icon name="settings" />
      <span>Settings</span>
    </a>
  </nav>

  <div class="sidebar-footer">
    <div class="user-card" onclick={() => {}} role="button" tabindex=0 onkeypress={() => {}}>
      <div class="avatar-circle">{authStore.user?.username ? authStore.user.username.charAt(0) : 'U'}</div>
      <div class="user-info">
        <span class="user-name">{authStore.user?.username || 'Operator'}</span>
        <span class="user-role">Tier: Enterprise</span>
      </div>
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

  .user-card {
    display: flex;
    align-items: center;
    gap: 12px;
    background: var(--surface-container-low);
    padding: 12px;
    border-radius: 12px;
    border: 1px solid var(--glass-border);
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      background: var(--surface-container-high);
      border-color: var(--primary);
    }
  }
</style>
