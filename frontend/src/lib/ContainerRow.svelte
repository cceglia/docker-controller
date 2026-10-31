<!-- One Container row: status dot, name, image, state, health and the published
     ports. The tone of the dot and of the health pill come from status.ts, the
     one place for those rules. -->
<script lang="ts">
  import type { Container } from '../../bindings/dockercontroller/internal/compose/models'
  import StatusDot from './StatusDot.svelte'
  import { containerTone, healthTone } from './status'

  let {
    container,
    onopenport,
  }: {
    container: Container
    /** Opens http://localhost:<port> in the browser. */
    onopenport: (port: number) => void
  } = $props()
</script>

<li class="container">
  <StatusDot tone={containerTone(container)} />
  <span class="cname mono" title={container.name}>{container.name}</span>
  <span class="image mono" title={container.image}>{container.image}</span>
  <span class="state">{container.state}</span>
  {#if container.health}
    {@const tone = healthTone(container.health)}
    {#if tone}
      <span class="pill {tone}">{container.health}</span>
    {:else}
      <span class="health">{container.health}</span>
    {/if}
  {/if}
  {#if (container.ports ?? []).length > 0}
    <span class="ports">
      {#each (container.ports ?? []) as port (`${port.published}/${port.protocol}`)}
        <button
          class="port"
          title="Apri http://localhost:{port.published}"
          aria-label="Apri http://localhost:{port.published} nel browser: porta {port.target}/{port.protocol}"
          onclick={() => onopenport(port.published)}
        >{port.published}→{port.target}/{port.protocol}</button>
      {/each}
    </span>
  {/if}
</li>

<style>
  /* Container rows: dot, name, image, state, health, published ports. */
  .container {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px var(--space-1);
    min-height: var(--row-height);
    padding: 6px 0;
    border-top: 1px solid var(--rule);
    font-size: var(--fs-small);
  }
  .cname {
    max-width: 100%;
    font-size: var(--fs-body);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .image {
    flex: 1 1 10rem;
    min-width: 0;
    color: var(--muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .state,
  .health {
    color: var(--muted);
  }
  /* A pill marks a problem only: unhealthy, or a healthcheck still starting. */
  .pill {
    flex: none;
    padding: 1px 4px;
    border-radius: var(--radius-control);
    font-size: var(--fs-small);
    font-weight: 500;
    line-height: 1.2;
    white-space: nowrap;
  }
  .pill.negative {
    background: var(--pill-negative-bg);
    color: var(--pill-negative-ink);
  }
  .pill.warning {
    background: var(--pill-warning-bg);
    color: var(--pill-warning-ink);
  }
  .ports {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    margin-left: auto;
  }
  .port {
    height: var(--control-height);
    padding: 0 8px;
    border: 1px solid var(--control);
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--text);
    font-family: var(--font-mono);
    font-size: var(--fs-small);
    font-variant-numeric: tabular-nums;
    line-height: 1;
    cursor: pointer;
    transition: background-color 120ms ease;
  }
  .port:hover {
    background: var(--raised);
  }
  .port:active {
    background: var(--active);
  }
</style>
