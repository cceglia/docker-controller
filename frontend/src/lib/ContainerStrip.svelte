<!-- The strip of Container: one tack per Container, filled while it runs,
     empty while it is stopped, in the error colour when Docker reports it
     unhealthy or dead. Past twelve tacks the rest becomes "+n", so the shape
     of "3 of 5" reads at a glance and never grows with the Project.
     It is drawn from the Containers themselves, so it needs no caption; the
     screen reader text that describes it belongs to the row around it. -->
<script lang="ts">
  import type { Container } from '../../bindings/dockercontroller/internal/compose/models'

  let {
    containers,
  }: {
    containers: Container[]
  } = $props()

  /** Tacks drawn before the "+n" counter takes over. */
  const MAX = 12

  // A Container is at fault when Docker reports it unhealthy or dead.
  const faulty = (c: Container) => c.health === 'unhealthy' || c.state === 'dead'

  const shown = $derived(containers.slice(0, MAX))
  const more = $derived(containers.length - shown.length)
</script>

<span class="strip" aria-hidden="true">
  {#each shown as container, index (index)}
    <span class="tack" class:running={container.state === 'running'} class:faulty={faulty(container)}></span>
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

  /* Narrow column: six tacks a row, the "+n" under them. */
  @media (max-width: 900px) {
    .strip {
      justify-content: center;
      max-width: calc(6 * var(--strip-tack-width) + 5 * var(--strip-gap));
    }
    .more {
      margin-left: 0;
    }
  }
</style>
