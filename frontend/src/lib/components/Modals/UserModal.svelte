<script lang="ts">
  import { api } from '$lib/api.svelte';
  import { fade } from 'svelte/transition';

  let { show = $bindable() } = $props();

  let editUsername = $state(api.user?.username || '');
  let oldPassword = $state('');
  let newPassword = $state('');
  let confirmPassword = $state('');

  async function updateProfile() {
    if (newPassword && newPassword !== confirmPassword) {
      api.showToast('Passwords do not match.', 'error');
      return;
    }
    const res = await api.apiFetch('/api/user', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        username: editUsername, 
        old_password: oldPassword, 
        new_password: newPassword 
      })
    });
    if (res.ok) {
      api.showToast('Profile updated.', 'success');
      show = false;
      oldPassword = ''; newPassword = ''; confirmPassword = '';
      api.checkAuth();
    } else {
      const data = await res.json();
      api.showToast(data.error || 'Update failed.', 'error');
    }
  }
</script>

<div class="modal-overlay" onclick={() => show = false} transition:fade>
  <div class="modal-content" style="max-width: 500px;" onclick={(e) => e.stopPropagation()}>
    <div class="modal-header">
      <h2 style="font-size: 1.5rem; margin: 0;">User Configuration</h2>
    </div>
    <div class="modal-body">
      <label>ADMIN USERNAME</label>
      <input type="text" bind:value={editUsername} placeholder="e.g. emiel" />
      
      <div style="border-top: 1px solid var(--glass-border); margin: 32px 0; padding-top: 32px;">
        <label style="color: var(--status-yellow);">CHANGE ACCESS CODE (OPTIONAL)</label>
        <input type="password" bind:value={oldPassword} placeholder="CURRENT PASSWORD" />
        <input type="password" bind:value={newPassword} placeholder="NEW PASSWORD" />
        <input type="password" bind:value={confirmPassword} placeholder="CONFIRM NEW PASSWORD" />
      </div>

      <div style="display: flex; gap: 16px;">
        <button style="flex: 2;" onclick={updateProfile}>UPDATE PROFILE</button>
        <button class="secondary" style="flex: 1;" onclick={() => show = false}>CANCEL</button>
      </div>
    </div>
  </div>
</div>
