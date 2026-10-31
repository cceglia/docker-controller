<!-- A pair of rates, in/out or lettura/scrittura: one row per reading, the
     labels in a column of their own, so the values line up under each other.
     The formatting is the one of formatRate, the same for the global strip and
     for the Servizi. -->
<script lang="ts">
  import { formatRate } from './format'

  type RateReading = { label: string; bytesPerSec: number }

  let {
    readings,
  }: {
    /** The two readings of the pair, in display order. */
    readings: [RateReading, RateReading]
  } = $props()
</script>

<span class="pair">
  {#each readings as reading (reading.label)}
    <span class="reading">
      <span class="dir">{reading.label}</span>
      <span class="value">{formatRate(reading.bytesPerSec)}</span>
    </span>
  {/each}
</span>

<style>
  .pair {
    display: grid;
    grid-template-columns: max-content minmax(0, 1fr);
    column-gap: 6px;
    min-width: 0;
  }
  /* each reading is its own row, with the label and the value in the columns of the pair */
  .reading {
    display: contents;
  }
  .dir {
    color: var(--muted);
    font-size: var(--fs-small);
    line-height: 1.4;
  }
  .value {
    font-family: var(--font-mono);
    font-size: var(--fs-small);
    font-variant-numeric: tabular-nums;
    line-height: 1.4;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
</style>
