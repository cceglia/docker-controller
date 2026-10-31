<!-- The strip of Container: one tack per Container, filled while it runs,
     empty while it is stopped, in the error colour when Docker reports it
     unhealthy or dead (the rule is in status.ts). Past twelve tacks the rest
     becomes "+n", so the shape of "3 of 5" reads at a glance and never grows
     with the Project.
     It is drawn from the Containers themselves, so it needs no caption; the
     screen reader text that describes it belongs to the row around it.
     The compact variant has narrower and closer tacks: the column of the
     Progetti uses it, so the name of the Project keeps its room. -->
<script lang="ts">
  import type { Container } from '../../bindings/dockercontroller/internal/compose/models'
  import { containerFaulty } from './status'

  let {
    containers,
    compact = false,
  }: {
    containers: Container[]
    /** Narrower and closer tacks, for the column of the Progetti. */
    compact?: boolean
  } = $props()

  /** Tacks drawn before the "+n" counter takes over. */
  const MAX = 12

  const shown = $derived(containers.slice(0, MAX))
  const more = $derived(containers.length - shown.length)
</script>

<span class="strip" class:compact aria-hidden="true">
  {#each shown as container, index (index)}
    <span class="tack" class:running={container.state === 'running'} class:faulty={containerFaulty(container)}></span>
  {/each}
  {#if more > 0}<span class="more mono">+{more}</span>{/if}
</span>

<style>
  .strip {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: flex-end;
    gap: var(--strip-gap);
    max-width: 100%;
  }
  .tack {
    flex: none;
    width: var(--strip-tack-width);
    height: var(--strip-tack-height);
    background: var(--active);
  }
  .tack.running {
    background: var(--text);
  }
  .tack.faulty {
    background: var(--negative);
  }
  .more {
    margin-left: var(--strip-gap);
    font-size: var(--fs-small);
    color: var(--muted);
  }
  /* The compact strips share the row with the name of the Project: the tacks
     get narrower and closer, and the height stays the one of a tack. */
  .strip.compact {
    gap: var(--strip-gap-compact);
  }
  .strip.compact .tack {
    width: var(--strip-tack-width-compact);
  }
  .strip.compact .more {
    margin-left: var(--strip-gap-compact);
  }

  /* Narrow column: the strip wraps inside the column and each row is centred. */
  @media (max-width: 900px) {
    .strip {
      justify-content: center;
    }
    .more {
      margin-left: 0;
    }
  }
</style>
