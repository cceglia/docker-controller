<script lang="ts">
  import type { UsageSample } from '../../bindings/dockercontroller/internal/compose/models'
  import { formatBytes, formatCores, formatPercent } from './format'
  import Meter from './Meter.svelte'

  let { usage }: {
    /** The latest sample; null while none has arrived (streaming just started). */
    usage: UsageSample | null
  } = $props()

  const rate = (bytesPerSec: number) => `${formatBytes(bytesPerSec, 'decimal')}/s`
</script>

<!-- The global Utilizzo risorse: it says whose numbers these are. -->
<section class="usage" aria-label="Utilizzo risorse di Docker">
  <div class="owner">
    <span class="owner-name">Docker</span>
    <span class="owner-scope">Utilizzo risorse</span>
  </div>

  {#if usage === null}
    <p class="state">In attesa della prima misura…</p>
  {:else if !usage.dockerReachable}
    <p class="state" title={usage.error}>
      Non disponibile
      <span class="state-detail">Docker non raggiungibile</span>
    </p>
  {:else}
    <div class="metrics">
      <div class="metric">
        <span class="label">Container</span>
        <span class="value num">{usage.containers}</span>
        <span class="note">in esecuzione</span>
      </div>

      <div class="metric">
        <span class="label">CPU</span>
        <span class="value num">{formatCores(usage.total.cpuPercent)}</span>
        <span class="note num">{formatPercent(usage.cpuMachinePercent)} della macchina</span>
        <Meter value={usage.cpuMachinePercent} label="CPU" tone="accent" wide />
      </div>

      <div class="metric">
        <span class="label">Memoria</span>
        <span class="value num">{formatBytes(usage.total.memoryBytes)}</span>
        <span class="note num">
          {formatPercent(usage.memoryMachinePercent)} di {formatBytes(usage.memoryTotalBytes)}
        </span>
        <Meter value={usage.memoryMachinePercent} label="Memoria" wide />
      </div>

      <div class="metric">
        <span class="label">Rete</span>
        <span class="pair">
          <span class="reading"><span class="dir">in</span><span class="value num">{rate(usage.rate.netInBytesPerSec)}</span></span>
          <span class="reading"><span class="dir">out</span><span class="value num">{rate(usage.rate.netOutBytesPerSec)}</span></span>
        </span>
      </div>

      <div class="metric">
        <span class="label">Disco</span>
        <span class="pair">
          <span class="reading"><span class="dir">lettura</span><span class="value num">{rate(usage.rate.blockReadBytesPerSec)}</span></span>
          <span class="reading"><span class="dir">scrittura</span><span class="value num">{rate(usage.rate.blockWriteBytesPerSec)}</span></span>
        </span>
      </div>
    </div>
  {/if}
</section>

<style>
  .usage {
    flex: none;
    display: flex;
    align-items: stretch;
    gap: var(--space-2);
    box-sizing: border-box;
    min-height: var(--strip-height);
    padding: var(--space-1) var(--space-2);
    background: var(--surface);
    border-bottom: 1px solid var(--border);
  }

  .owner {
    flex: none;
    display: grid;
    align-content: center;
    gap: 2px;
    padding-right: var(--space-2);
    border-right: 1px solid var(--rule);
  }
  .owner-name {
    font-size: var(--fs-body);
    font-weight: 600;
    line-height: 1.3;
    letter-spacing: -0.01em;
  }
  .owner-scope {
    font-size: var(--fs-small);
    line-height: 1.3;
    color: var(--muted);
  }

  .state {
    align-self: center;
    display: flex;
    align-items: baseline;
    gap: 8px;
    color: var(--muted);
    font-size: var(--fs-body);
  }
  .state-detail {
    color: var(--faint);
    font-size: var(--fs-small);
  }

  .metrics {
    flex: 1;
    min-width: 0;
    display: grid;
    grid-template-columns: 0.9fr 1.35fr 1.45fr 1.15fr 1.16fr;
    gap: var(--space-1);
    align-content: center;
  }

  /* one cell per measure; the hairline lives on the cell's left edge */
  .metric {
    min-width: 0;
    display: grid;
    align-content: center;
    gap: 2px;
    padding-left: var(--space-1);
    border-left: 1px solid var(--rule);
  }
  .metric:first-child {
    border-left: 0;
    padding-left: 0;
  }

  .label {
    font-size: var(--fs-small);
    line-height: 1.3;
    color: var(--muted);
  }
  .value {
    font-family: var(--font-mono);
    font-size: var(--fs-body);
    font-weight: 500;
    line-height: 1.4;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .note {
    font-size: var(--fs-small);
    line-height: 1.3;
    color: var(--muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .pair {
    display: grid;
    gap: 1px;
    min-width: 0;
  }
  .reading {
    display: flex;
    align-items: baseline;
    gap: 6px;
    min-width: 0;
  }
  .dir {
    flex: none;
    min-width: 5ch;
    color: var(--muted);
    font-size: var(--fs-small);
    line-height: 1.3;
  }
  .reading .value {
    font-size: var(--fs-small);
  }

  /* the window can be made narrower: the strip folds to two rows instead of
     truncating its notes */
  @media (max-width: 1100px) {
    .metrics {
      grid-template-columns: repeat(6, minmax(0, 1fr));
      row-gap: var(--space-1);
    }
    .metric {
      grid-column: span 2;
    }
    .metric:nth-child(4),
    .metric:nth-child(5) {
      grid-column: span 3;
    }
    .metric:nth-child(4) {
      border-left: 0;
      padding-left: 0;
    }
  }
</style>
