<script lang="ts">
  import { api, apiFetch, toastHandler } from '$lib/api.svelte';
  import { fade } from 'svelte/transition';

  let { show = $bindable() } = $props();

  let editUsername = $state(api.user?.username || '');
  let oldPassword = $state('');
  let newPassword = $state('');
  let confirmPassword = $state('');

  let githubToken = $state('');
  let gitlabToken = $state('');
  let hasGithubToken = $state(false);
  let hasGitlabToken = $state(false);

  async function updateProfile() {
    if (githubToken || gitlabToken) {
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
      api.checkAuth();
    } else {
      const data = await res.json();
      toastHandler.showToast(data.error || 'Update failed.', 'error');
    }
  }

  async function fetchSettings() {
    const settings = await api.settings.getSettings();
    if (settings) {
      hasGithubToken = settings.github_token_set;
      hasGitlabToken = settings.gitlab_token_set;
    }
  }

  async function updateSettings() {
    const res = await api.settings.setSettings(githubToken, gitlabToken);
    if (res) {
      toastHandler.showToast('Token saved', 'success');
    } else {
      toastHandler.showToast('Token not saved, try again', 'error');
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
        <label for='github'>Github Token</label>
        <input id='github' type='text' bind:value={githubToken} placeholder={ hasGithubToken ? '*****************************************************************************' : 'Secret github token'} />
        <label for='gitlab'>Gitlab Token</label>
        <input id='gitlab' type='text' bind:value={gitlabToken} placeholder={ hasGitlabToken ? '*****************************************************************************' : 'Secret github token'} />
      </div>

      <div style="display: flex; gap: 16px;">
        <button style="flex: 2;" onclick={updateProfile}>UPDATE PROFILE</button>
        <button class="secondary" style="flex: 1;" onclick={() => show = false}>CANCEL</button>
      </div>
    </div>
  </div>
</div>
