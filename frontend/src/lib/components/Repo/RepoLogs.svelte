<script lang="ts">
  import type { Commit } from '$lib/types';
  let { commits } = $props<{ commits: Commit[] }>();
</script>

<div style="display: flex; flex-direction: column; gap: 12px;">
  {#each commits as commit}
    <div style="background: rgba(255,255,255,0.02); border-radius: 16px; border: 1px solid var(--glass-border); padding: 20px;">
      <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 8px;">
        {#if commit.branch}<span class="branch-badge">{commit.branch}</span>{/if}
        <span style="font-weight: 600; font-size: 1rem;">{commit.message}</span>
      </div>
      <div style="font-size: 0.8rem; color: var(--efinity-text-muted); font-weight: 500;">
        {commit.author} • {commit.date} • <span style="color: var(--efinity-blue); font-family: monospace;">{commit.hash.substring(0,7)}</span>
      </div>
    </div>
  {/each}
</div>
