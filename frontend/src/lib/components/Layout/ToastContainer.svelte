<script lang="ts">
  import { api } from '$lib/api.svelte';
  import { fade, fly } from 'svelte/transition';
</script>

<div class="toast-container">
  {#each api.toasts as toast (toast.id)}
    <div 
      class="toast {toast.type}" 
      in:fly={{ y: 20, duration: 400 }} 
      out:fade={{ duration: 200 }}
    >
      {#if toast.type === 'success'}
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
      {:else if toast.type === 'error'}
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
      {:else}
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>
      {/if}
      {toast.message}
    </div>
  {/each}
</div>
