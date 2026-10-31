<!-- The app's mark: three stacked blocks, the top one lit in the data hue. -->
<script lang="ts">
  let { size = 20 }: { size?: number } = $props()
</script>

<svg class="mark" width={size} height={size} viewBox="0 0 20 20" role="img" aria-label="Docker Controller">
  <rect x="2" y="3" width="16" height="4" rx="1.5" fill="var(--accent)" />
  <rect x="2" y="8.5" width="16" height="3.5" rx="1.5" fill="currentColor" />
  <rect x="2" y="13.5" width="11" height="3.5" rx="1.5" fill="currentColor" />
</svg>

<style>
  .mark {
    flex: none;
    display: block;
  }
</style>
