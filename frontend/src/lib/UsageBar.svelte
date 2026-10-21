<script lang="ts">
  import type { UsageSample } from '../../bindings/dockercontroller/internal/compose/models'
  import { formatBytes, formatCores, formatPercent } from './format'

  let { usage }: {
    /** The latest sample; null while none has arrived (streaming just started). */
    usage: UsageSample | null
  } = $props()
</script>

<div class="usage-bar" class:error={usage !== null && !usage.dockerReachable} aria-label="Utilizzo risorse di Docker">
  {#if usage === null}
    <span class="waiting">Utilizzo risorse: lettura in corso…</span>
  {:else if !usage.dockerReachable}
    <span class="failure" title={usage.error}>Utilizzo risorse non disponibile: Docker non raggiungibile</span>
  {:else}
    <span title="{usage.cpus} CPU">
      <b>CPU</b> {formatCores(usage.total.cpuPercent)} · {formatPercent(usage.cpuMachinePercent)}
    </span>
    <span title="Memoria totale della macchina: {formatBytes(usage.memoryTotalBytes)}">
      <b>Memoria</b> {formatBytes(usage.total.memoryBytes)} · {formatPercent(usage.memoryMachinePercent)} di {formatBytes(usage.memoryTotalBytes)}
    </span>
    <span><b>Rete</b> in {formatBytes(usage.total.netInBytes, 'decimal')} · out {formatBytes(usage.total.netOutBytes, 'decimal')}</span>
    <span><b>Disco</b> lettura {formatBytes(usage.total.blockReadBytes, 'decimal')} · scrittura {formatBytes(usage.total.blockWriteBytes, 'decimal')}</span>
    <span class="count">{usage.containers} Container in esecuzione</span>
  {/if}
</div>

<style>
  .usage-bar {
    display: flex; flex-wrap: wrap; gap: 0.25rem 1.25rem; align-items: baseline;
    margin: 0 0 1rem; padding: 0.5rem 0.9rem; border-radius: 8px; font-size: 0.85rem;
    background: var(--surface); border: 1px solid var(--border);
  }
  .usage-bar.error { background: var(--warn-bg); color: var(--warn); border-color: transparent; }
  b { font-weight: 600; margin-right: 0.2rem; }
  .waiting, .count { color: var(--muted); }
  .count { margin-left: auto; }
</style>
