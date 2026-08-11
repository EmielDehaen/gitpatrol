<script lang="ts">
  let { issues } = $props<{ issues: any[] }>();
</script>

<div style="display: flex; flex-direction: column; gap: 16px;">
  {#each issues as issue}
    <div style="background: rgba(255,255,255,0.02); border-radius: 16px; border: 1px solid var(--glass-border); padding: 20px; opacity: {issue.state === 'closed' ? 0.6 : 1}">
      <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 8px;">
        <div style="font-weight: 700; font-size: 1.1rem; color: #fff;">{issue.title}</div>
        <span class="badge" style="color: {issue.state === 'open' ? 'var(--status-green)' : 'var(--efinity-text-muted)'}; background: {issue.state === 'open' ? 'rgba(0, 255, 136, 0.05)' : 'rgba(255,255,255,0.03)'}">{issue.state.toUpperCase()}</span>
      </div>
      <div style="font-size: 0.8rem; color: var(--efinity-text-muted);">#{issue.number} opened by {issue.user?.login} • {new Date(issue.created_at).toLocaleDateString()}</div>
    </div>
  {:else}
    <p style="color: var(--efinity-text-muted); text-align: center; padding: 40px;">No tactical issues found in this sector.</p>
  {/each}
</div>
