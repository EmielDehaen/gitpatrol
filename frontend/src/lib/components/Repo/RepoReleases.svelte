<script lang="ts">
  import { marked } from '$lib/markdown';
  import type { Release } from '$lib/types';
  let { releases } = $props<{ releases: Release[] }>();
</script>

<div style="display: flex; flex-direction: column; gap: 24px;">
  {#each releases as release}
    <div style="background: rgba(255,255,255,0.02); border-radius: 20px; border: 1px solid var(--glass-border); padding: 24px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; border-bottom: 1px solid var(--glass-border); padding-bottom: 16px;">
        <div>
          <div style="font-size: 1.4rem; font-weight: 800; color: var(--efinity-blue);">{release.tag_name}</div>
          <div style="font-size: 0.8rem; color: var(--efinity-text-muted);">{release.name || ''} • {new Date(release.published_at).toLocaleDateString()}</div>
        </div>
      </div>
      <div class="readme-content" style="font-size: 0.9rem;">{@html marked.parse(release.body || '')}</div>
    </div>
  {:else}
    <p style="color: var(--efinity-text-muted); text-align: center; padding: 40px;">No historical chronicles (releases) found.</p>
  {/each}
</div>
