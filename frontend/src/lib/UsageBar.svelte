<script lang="ts">
  import type { UsageSample } from '../../bindings/dockercontroller/internal/compose/models'
  import { formatBytes, formatCores, formatPercent } from './format'

  let { usage }: {
    /** The latest sample; null while none has arrived (streaming just started). */
    usage: UsageSample | null
  } = $props()

  const rate = (bytesPerSec: number) => `${formatBytes(bytesPerSec, 'decimal')}/s`
</script>

<div class="usage-bar" class:error={usage !== null && !usage.dockerReachable} aria-label="Utilizzo risorse di Docker">
  {#if usage === null}
    <span class="waiting">Utilizzo risorse: lettura in corso…</span>
  {:else if !usage.dockerReachable}
    <span class="failure" title={usage.error}>Utilizzo risorse non disponibile: Docker non raggiungibile</span>
  {:else}
    <div class="cell c4" title="{usage.cpus} CPU">
      <span class="label">CPU</span>
      <span class="main"><b>{formatCores(usage.total.cpuPercent)}</b><span class="sub">{formatPercent(usage.cpuMachinePercent)} della macchina</span></span>
    </div>
    <div class="cell c5" title="Memoria totale della macchina: {formatBytes(usage.memoryTotalBytes)}">
      <span class="label">Memoria</span>
      <span class="main"><b>{formatBytes(usage.total.memoryBytes)}</b><span class="sub">{formatPercent(usage.memoryMachinePercent)} di {formatBytes(usage.memoryTotalBytes)}</span></span>
    </div>
    <div class="cell c3">
      <span class="label">Container</span>
      <span class="main"><b>{usage.containers}</b><span class="sub">in esecuzione</span></span>
    </div>
    <div class="cell c6" title="Traffico istantaneo">
      <span class="label">Rete</span>
      <span class="pair">
        <span class="stat"><span class="dir">↓ in</span><b>{rate(usage.rate.netInBytesPerSec)}</b></span>
        <span class="stat"><span class="dir">↑ out</span><b>{rate(usage.rate.netOutBytesPerSec)}</b></span>
      </span>
    </div>
    <div class="cell c6" title="I/O istantaneo">
      <span class="label">Disco</span>
      <span class="pair">
        <span class="stat"><span class="dir">lettura</span><b>{rate(usage.rate.blockReadBytesPerSec)}</b></span>
        <span class="stat"><span class="dir">scrittura</span><b>{rate(usage.rate.blockWriteBytesPerSec)}</b></span>
      </span>
    </div>
  {/if}
</div>

<style>
  .usage-bar {
    display: grid; grid-template-columns: repeat(12, 1fr); margin: 0 0 1.25rem;
    border-radius: 10px; font-size: 0.85rem; overflow: hidden;
    background: var(--surface); border: 1px solid var(--border); box-shadow: var(--shadow);
  }
  .usage-bar.error { display: block; background: var(--warn-bg); color: var(--warn); border-color: transparent; padding: 0.6rem 1rem; }
  .waiting { display: block; padding: 0.6rem 1rem; color: var(--muted); }
  .cell { display: flex; flex-direction: column; gap: 0.2rem; padding: 0.6rem 1rem; min-width: 0; border-bottom: 1px solid var(--border); }
  .cell:nth-child(n + 4) { border-bottom: none; }
  .cell:not(:nth-child(3n + 1)):not(:nth-child(n + 4)) { border-left: 1px solid var(--border); }
  .cell:nth-child(5) { border-left: 1px solid var(--border); }
  .c3 { grid-column: span 3; } .c4 { grid-column: span 4; } .c5 { grid-column: span 5; } .c6 { grid-column: span 6; }
  .label { font-size: 0.68rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.08em; color: var(--muted); }
  .main { display: flex; align-items: baseline; gap: 0.5rem; white-space: nowrap; }
  .main b { font-size: 1.1rem; font-weight: 600; font-variant-numeric: tabular-nums; }
  .sub { color: var(--muted); font-size: 0.8rem; overflow: hidden; text-overflow: ellipsis; }
  .pair { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
  .stat { display: flex; align-items: baseline; gap: 0.5rem; white-space: nowrap; }
  .stat b { font-weight: 600; font-variant-numeric: tabular-nums; min-width: 6.5ch; }
  .dir { color: var(--muted); font-size: 0.8rem; min-width: 4.2rem; }
</style>
