<script lang="ts">
  let { wikiContent } = $props<{ wikiContent: string }>();
</script>

<div class="readme-container readme-expanded">
  <div class="readme-content">
    {@html wikiContent}
  </div>
</div>
