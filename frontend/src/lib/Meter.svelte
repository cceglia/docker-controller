<!-- A thin proportion indicator: a 4px track with its true ratio filled.
     CPU carries the screen's one data hue; everything else fills in ink. -->
<script lang="ts">
  let {
    value,
    label,
    tone = 'neutral',
    wide = false,
  }: {
    /** 0-100. */
    value: number
    /** The accessible name of the measure. */
    label: string
    tone?: 'neutral' | 'accent'
    /** Fill the cell it sits in instead of keeping the token width. */
    wide?: boolean
  } = $props()

  const percent = $derived(Math.round(Math.max(0, Math.min(100, value))))
</script>

<span
  class="meter"
  class:accent={tone === 'accent'}
  class:wide
  role="progressbar"
  aria-label={label}
  aria-valuemin={0}
  aria-valuemax={100}
  aria-valuenow={percent}
>
  <span class="bar" style:width="{percent}%"></span>
</span>

<style>
  .meter {
    display: inline-block;
    flex: none;
    width: var(--meter-width);
    height: var(--meter-height);
    border-radius: calc(var(--meter-height) / 2);
    background: var(--active);
    overflow: hidden;
  }
  .meter.wide {
    display: block;
    width: 100%;
    max-width: var(--meter-max-width);
  }
  .bar {
    display: block;
    height: 100%;
    border-radius: inherit;
    background: var(--text);
  }
  .accent .bar {
    background: var(--accent);
  }
</style>
