<script lang="ts">
  import { toastHandler } from '$lib/toast.svelte';
  import { fade, fly } from 'svelte/transition';
  import Icon from '$lib/components/Icon.svelte';
</script>

<div class="toast-container">
  {#each toastHandler.toasts as toast (toast.id)}
    <div 
      class="toast {toast.type}" 
      in:fly={{ y: 20, duration: 400 }} 
      out:fade={{ duration: 200 }}
    >
      {#if toast.type === 'success'}
        <Icon name="check" size={18} strokeWidth={3} />
      {:else if toast.type === 'error'}
        <Icon name="alert-circle" size={18} strokeWidth={3} />
      {:else}
        <Icon name="info" size={18} strokeWidth={3} />
      {/if}
      {toast.message}
    </div>
  {/each}
</div>
