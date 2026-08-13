<script lang="ts">
  import { fade } from 'svelte/transition';
  import { apiFetch } from '$lib/client';
  import { toastHandler } from '$lib/toast.svelte';
  import { settingsApi } from '$lib/settings/settingsApi';
  import { authStore } from '$lib/auth.svelte';

  let activeTab = $state('profile');

  let editUsername = $state(authStore.user?.username || '');
  let oldPassword = $state('');
  let newPassword = $state('');
  let confirmPassword = $state('');

  let githubToken = $state('');
  let gitlabURL = $state('');
  let gitlabToken = $state('');
  let giteaURL = $state('');
  let giteaToken = $state('');
  let exportDestination = $state('');

  let hasGithubToken = $state(false);
  let hasGitlabToken = $state(false);
  let hasGiteaToken = $state(false);

  async function updateProfile() {
    if (newPassword && newPassword !== confirmPassword) {
      toastHandler.showToast('Passwords do not match.', 'error');
      return;
    }

    if (!oldPassword || !newPassword || !confirmPassword) {
      // Just update username if passwords aren't provided? 
      // The API currently expects all 3 if updating password, but we'll send anyway.
    }

    const res = await apiFetch('/api/user', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        username: editUsername, 
        old_password: oldPassword, 
        new_password: newPassword 
      })
    });
    if (res.ok) {
      toastHandler.showToast('Profile updated.', 'success');
      oldPassword = ''; newPassword = ''; confirmPassword = '';
      authStore.checkAuth();
    } else {
      const data = await res.json();
      toastHandler.showToast(data.error || 'Update failed.', 'error');
    }
  }

  async function fetchSettings() {
    const settings = await settingsApi.getSettings();
    if (settings) {
      hasGithubToken = settings.github_token_set;
      gitlabURL = settings.gitlab_url;
      hasGitlabToken = settings.gitlab_token_set;
      giteaURL = settings.gitea_url;
      hasGiteaToken = settings.gitea_token_set;
      exportDestination = settings.export_destination;
    }
  }

  async function updateSettings() {
    const res = await settingsApi.setSettings(githubToken, gitlabURL, gitlabToken, giteaURL, giteaToken, exportDestination);
    if (res) {
      toastHandler.showToast('Vault integrations saved.', 'success');
      githubToken = ''; gitlabToken = ''; giteaToken = '';
      fetchSettings();
    } else {
      toastHandler.showToast('Failed to save integrations.', 'error');
    }
  }

  $effect(() => {
    fetchSettings();
  });
</script>

<div class="page-container" transition:fade>
  <div class="settings-header">
    <h1 class="display-lg">Settings</h1>
    <div class="tab-container" style="width: auto; margin-bottom: 0;">
      <button class="tab-btn" class:active={activeTab === 'profile'} onclick={() => activeTab = 'profile'}>PROFILE</button>
      <button class="tab-btn" class:active={activeTab === 'vault'} onclick={() => activeTab = 'vault'}>VAULT INTEGRATIONS</button>
    </div>
  </div>

  <div class="settings-content">
    {#if activeTab === 'profile'}
      <div class="settings-card">
        <h2 style="font-size: 1.25rem; margin-bottom: 24px;">Operator Identity</h2>
        <label for="editUsername">USERNAME</label>
        <input id="editUsername" type="text" bind:value={editUsername} placeholder="e.g. emiel" />
        
        <div class="divider"></div>
        
        <h2 style="font-size: 1.25rem; margin-bottom: 24px; color: var(--warning);">Change Access Code</h2>
        <label for="oldPassword">CURRENT PASSWORD</label>
        <input id="oldPassword" type="password" bind:value={oldPassword} placeholder="Enter current password" />
        
        <label for="newPassword">NEW PASSWORD</label>
        <input id="newPassword" type="password" bind:value={newPassword} placeholder="Enter new password" />
        
        <label for="confirmPassword">CONFIRM NEW PASSWORD</label>
        <input id="confirmPassword" type="password" bind:value={confirmPassword} placeholder="Confirm new password" />
        
        <div style="margin-top: 32px;">
          <button onclick={updateProfile}>SAVE PROFILE</button>
        </div>
      </div>
    {:else}
      <div class="settings-card">
        <h2 style="font-size: 1.25rem; margin-bottom: 24px;">Primary Recovery Vault</h2>
        <label for="vault">VAULT DESTINATION</label>
        <select id="vault" bind:value={exportDestination} style="width: 100%; margin-bottom: 24px; border: none; background: var(--surface-container-highest); color: var(--on-surface); padding: 12px; border-radius: 8px;">
          <option value="">NONE CONFIGURED</option>
          <option value="github">GITHUB VAULT</option>
          <option value="gitlab">GITLAB VAULT</option>
          <option value="gitea">GITEA RECOVERY</option>
        </select>

        {#if exportDestination === 'gitlab'}
          <label for="gitlab_url">GITLAB URL</label>
          <input id="gitlab_url" type="text" bind:value={gitlabURL} placeholder="https://gitlab.com" />
        {:else if exportDestination === 'gitea'}
          <label for="gitea_url">GITEA URL</label>
          <input id="gitea_url" type="text" bind:value={giteaURL} placeholder="https://gitea.example.com" />
        {/if}

        <div class="divider"></div>

        <h2 style="font-size: 1.25rem; margin-bottom: 24px;">Vault Credentials</h2>
        <label for="github">GITHUB TOKEN</label>
        <input id="github" type="password" bind:value={githubToken} placeholder={hasGithubToken ? '••••••••••••••••••••••••••••••••••••••••' : 'Secret github token'} />
        
        <label for="gitlab">GITLAB TOKEN</label>
        <input id="gitlab" type="password" bind:value={gitlabToken} placeholder={hasGitlabToken ? '••••••••••••••••••••••••••••••••••••••••' : 'Secret gitlab token'} />
        
        <label for="gitea_token">GITEA TOKEN</label>
        <input id="gitea_token" type="password" bind:value={giteaToken} placeholder={hasGiteaToken ? '••••••••••••••••••••••••••••••••••••••••' : 'Secret gitea token'} />

        <div style="margin-top: 32px;">
          <button onclick={updateSettings}>SAVE INTEGRATIONS</button>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .settings-header {
    padding: 40px;
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
  }
  
  .settings-content {
    padding: 0 40px 80px 40px;
    max-width: 800px;
  }

  .settings-card {
    background: var(--surface-container-low);
    border-radius: 24px;
    padding: 40px;
  }

  .divider {
    height: 1px;
    background: var(--surface-container-highest);
    margin: 32px 0;
  }
</style>
