<script lang="ts">
  import { apiFetch } from '$lib/client';
  import { toastHandler } from '$lib/toast.svelte';
  import { settingsApi } from '$lib/settings/settingsApi';
  import { authStore } from '$lib/auth.svelte';
  import { fade } from 'svelte/transition';

  let { show = $bindable() } = $props();

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
    if (githubToken || gitlabURL || gitlabToken || giteaURL || giteaToken || exportDestination) {
      updateSettings();
    }

    if (newPassword && newPassword !== confirmPassword) {
      toastHandler.showToast('Passwords do not match.', 'error');
      return;
    }

    if (!oldPassword || !newPassword || !confirmPassword) {
      return;
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
      show = false;
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
      toastHandler.showToast('Settings saved', 'success');
      githubToken = ''; gitlabToken = ''; giteaToken = ''; // Clear secret fields after save
      fetchSettings();
    } else {
      toastHandler.showToast('Settings not saved, try again', 'error');
    }
  }

  $effect(() => {
    fetchSettings();
  });
</script>

<div class="modal-overlay" onclick={() => show = false} onkeydown={(e) => e.key == 'ESCAPE' && (show = false)} role='presentation' transition:fade>
  <div class="modal-content" style="max-width: 500px;" onclick={(e) => e.stopPropagation()} role='none'>
    <div class="modal-header">
      <h2 style="font-size: 1.5rem; margin: 0;">User Configuration</h2>
    </div>
    <div class="modal-body">
      <label for='editUsername'>USERNAME</label>
      <input id='editUsername' type="text" bind:value={editUsername} placeholder="e.g. emiel" />
      
      <div style="border-top: 1px solid var(--glass-border); margin: 32px 0; padding-top: 32px;">
        <label for='oldPassword' style="color: var(--status-yellow);">CHANGE ACCESS CODE (OPTIONAL)</label>
        <input id='oldPassword' type="password" bind:value={oldPassword} placeholder="CURRENT PASSWORD" />
        <input id='newPassword' type="password" bind:value={newPassword} placeholder="NEW PASSWORD" />
        <input id='confirmPassword' type="password" bind:value={confirmPassword} placeholder="CONFIRM NEW PASSWORD" />
      </div>

      <div style="border-top: 1px solid var(--glass-border); margin: 32px 0; padding-top: 32px;">
        <label for='vault'>PRIMARY RECOVERY VAULT</label>
        <select id='vault' bind:value={exportDestination} style="width: 100%; margin-bottom: 24px;">
          <option value="">NONE CONFIGURED</option>
          <option value="github">GITHUB VAULT</option>
          <option value="gitlab">GITLAB VAULT</option>
          <option value="gitea">GITEA RECOVERY</option>
        </select>
        {#if exportDestination === 'gitlab'}
          <label for='gitlab_url'>GITLAB URL</label>
          <input id='gitlab_url' type='text' bind:value={gitlabURL} placeholder="https://gitlab.com" />
        {:else if exportDestination === 'gitea'}
          <label for='gitea_url'>GITEA URL</label>
          <input id='gitea_url' type='text' bind:value={giteaURL} placeholder="https://gitea.example.com" />
        {/if}
        <label for='github'>GITHUB TOKEN</label>
        <input id='github' type='text' bind:value={githubToken} placeholder={ hasGithubToken ? '*****************************************************************************' : 'Secret github token'} />
        <label for='gitlab'>GITLAB TOKEN</label>
        <input id='gitlab' type='text' bind:value={gitlabToken} placeholder={ hasGitlabToken ? '*****************************************************************************' : 'Secret gitlab token'} />
        <label for='gitea_token'>GITEA TOKEN</label>
        <input id='gitea_token' type='text' bind:value={giteaToken} placeholder={ hasGiteaToken ? '*****************************************************************************' : 'Secret gitea token'} />
      </div>

      <div style="display: flex; gap: 16px;">
        <button style="flex: 2;" onclick={updateProfile}>UPDATE PROFILE</button>
        <button class="secondary" style="flex: 1;" onclick={() => show = false}>CANCEL</button>
      </div>
    </div>
  </div>
</div>
