<script lang="ts">
  let { readmeContent } = $props<{ readmeContent: string }>();
  let readmeExpanded = $state(false);
</script>

<div class="readme-container" class:readme-expanded={readmeExpanded}>
  <div class="readme-content">
    {@html readmeContent}
  </div>
  {#if !readmeExpanded}
    <div class="readme-fade">
      <button class="secondary" style="font-size: 0.65rem; padding: 12px 24px;" onclick={() => readmeExpanded = true}>READ FULL MISSION BRIEFING</button>
    </div>
  {/if}
</div>
