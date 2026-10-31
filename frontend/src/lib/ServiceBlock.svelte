<!-- One Servizio: a flat block with its name, the strip of its Containers and the
     Actions (Start, Stop, Restart and Down). Below, the Utilizzo risorse in one
     cell per measure, then one row per Container. Down is discreet at rest and
     asks the container for the confirmation. -->
<script lang="ts">
  import type { Service, ServiceUsage } from '../../bindings/dockercontroller/internal/compose/models'
  import ContainerRow from './ContainerRow.svelte'
  import ContainerStrip from './ContainerStrip.svelte'
  import { formatBytes, formatCores, formatPercent } from './format'
  import Meter from './Meter.svelte'
  import RatePair from './RatePair.svelte'
  import { ActionKind, actionLabels } from './store.svelte'

  let {
    service,
    usage,
    waiting,
    unavailable,
    busy,
    onaction,
    ondown,
    onopenport,
  }: {
    service: Service
    /** The Utilizzo risorse of this Service; null when no Container of it runs. */
    usage: ServiceUsage | null
    /** No sample has arrived yet: the measures wait for it. */
    waiting: boolean
    /** Docker does not answer: the Actions are disabled. */
    unavailable: boolean
    /** An Activity runs: the Actions are disabled. */
    busy: boolean
    /** Runs a Service Action. */
    onaction: (kind: ActionKind) => void
    /** The Servizio Down was chosen: the container asks for the confirmation. */
    ondown: () => void
    /** Opens http://localhost:<port> in the browser. */
    onopenport: (port: number) => void
  } = $props()

  // The Actions of a Servizio stay on its own block.
  const kinds = [ActionKind.ActionStart, ActionKind.ActionStop, ActionKind.ActionRestart]

  // Generated bindings type arrays as nullable; the backend never sends null.
  const containers = $derived(service.containers ?? [])
  const running = $derived(containers.some((c) => c.state === 'running'))
</script>

<article class="service">
  <div class="service-head">
    <div class="service-title">
      <h4>{service.name}</h4>
      <ContainerStrip {containers} />
    </div>
    <div class="service-actions">
      {#each kinds as kind (kind)}
        <button class="quiet" disabled={unavailable || busy} onclick={() => onaction(kind)}>
          {actionLabels[kind]}
        </button>
      {/each}
      <button class="quiet danger" disabled={unavailable || busy} onclick={ondown}>
        {actionLabels[ActionKind.ActionDownService]}
      </button>
    </div>
  </div>
  {#if usage}
    {@const u = usage}
    <div class="usage" aria-label="Utilizzo risorse di {service.name}">
      <div class="measure">
        <span class="measure-label">CPU</span>
        <span class="measure-value num">{formatPercent(u.measures.cpuPercent)}</span>
        <span class="measure-note num">{formatCores(u.measures.cpuPercent)}</span>
        <Meter value={u.measures.cpuPercent} label="CPU" tone="accent" wide />
      </div>
      <div class="measure">
        <span class="measure-label">Memoria</span>
        <span class="measure-value num">{formatBytes(u.measures.memoryBytes)}</span>
      </div>
      <div class="measure">
        <span class="measure-label">Rete</span>
        <RatePair
          readings={[
            { label: 'in', bytesPerSec: u.rate.netInBytesPerSec },
            { label: 'out', bytesPerSec: u.rate.netOutBytesPerSec },
          ]}
        />
      </div>
      <div class="measure">
        <span class="measure-label">Disco</span>
        <RatePair
          readings={[
            { label: 'lettura', bytesPerSec: u.rate.blockReadBytesPerSec },
            { label: 'scrittura', bytesPerSec: u.rate.blockWriteBytesPerSec },
          ]}
        />
      </div>
    </div>
  {:else if waiting && running}
    <p class="usage-none" aria-label="Utilizzo risorse di {service.name}" title="In attesa della prima misura">In attesa della prima misura…</p>
  {:else}
    <p class="usage-none" aria-label="Utilizzo risorse di {service.name}" title="Nessun Container in esecuzione">Utilizzo risorse: —</p>
  {/if}
  <ul class="containers">
    {#each containers as container (container.name)}
      <ContainerRow {container} {onopenport} />
    {/each}
  </ul>
</article>

<style>
  .service {
    display: grid;
    gap: 8px;
    padding: var(--space-1) 0 0;
    border-top: 1px solid var(--border);
  }
  .service-head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 4px var(--space-1);
  }
  .service-title {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px var(--space-1);
    min-width: 0;
  }
  .service-title h4 {
    font-size: var(--fs-section);
    font-weight: 600;
    overflow-wrap: anywhere;
  }
  .service-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }

  /* The quiet control of the Servizio: same values as the container, because
     the styles of a Svelte component do not reach its children. */
  .quiet {
    height: var(--control-height);
    padding: 0 8px;
    border: 0;
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--muted);
    font-family: var(--font-text);
    font-size: var(--fs-body);
    line-height: 1;
    cursor: pointer;
    transition: background-color 120ms ease, color 120ms ease;
  }
  .quiet:not(:disabled):hover {
    background: var(--raised);
    color: var(--text);
  }
  .quiet:not(:disabled):active {
    background: var(--active);
  }
  .quiet:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  /* Down is discreet at rest and takes the error colour only on hover and focus. */
  .quiet.danger {
    color: var(--muted);
  }
  .quiet.danger:not(:disabled):hover,
  .quiet.danger:not(:disabled):focus-visible {
    color: var(--negative);
  }

  /* Utilizzo risorse: one cell per measure, an instant value or a rate. */
  .usage {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr));
    gap: var(--space-1);
    padding: var(--space-1) 0;
    border-top: 1px solid var(--rule);
  }
  .measure {
    display: grid;
    align-content: start;
    gap: 2px;
    min-width: 0;
    padding-left: var(--space-1);
    border-left: 1px solid var(--rule);
  }
  .measure:first-child {
    padding-left: 0;
    border-left: 0;
  }
  .measure-label {
    color: var(--muted);
    font-size: var(--fs-small);
    line-height: 1.3;
  }
  .measure-value,
  .measure-note {
    font-family: var(--font-mono);
    font-variant-numeric: tabular-nums;
    line-height: 1.4;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .measure-value {
    font-size: var(--fs-body);
    font-weight: 500;
  }
  .measure-note {
    color: var(--muted);
    font-size: var(--fs-small);
    line-height: 1.3;
  }
  .usage-none {
    margin: 0;
    padding: var(--space-1) 0;
    border-top: 1px solid var(--rule);
    color: var(--muted);
    font-size: var(--fs-body);
  }

  .containers {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
  }
</style>
