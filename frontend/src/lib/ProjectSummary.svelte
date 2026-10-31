<!-- The canvas with no Project open: the state of the saved Progetti as real
     counts, and the way back into them. It is not an empty screen: the column
     already lists the Progetti, here they are numbered. -->
<script lang="ts">
  import { Status } from '../../bindings/dockercontroller/internal/compose/models'
  import type { Project } from '../../bindings/dockercontroller/internal/compose/models'

  let {
    projects,
  }: {
    projects: Project[]
  } = $props()

  const running = $derived(projects.reduce((n, p) => n + p.running, 0))
  const total = $derived(projects.reduce((n, p) => n + p.total, 0))
  const counts = $derived([
    { key: Status.StatusActive, label: 'Attivi', value: projects.filter((p) => p.status === Status.StatusActive).length },
    { key: Status.StatusPartial, label: 'Parziali', value: projects.filter((p) => p.status === Status.StatusPartial).length },
    { key: Status.StatusStopped, label: 'Fermi', value: projects.filter((p) => p.status === Status.StatusStopped).length },
    { key: Status.StatusRemoved, label: 'Rimossi', value: projects.filter((p) => p.status === Status.StatusRemoved).length },
  ])
</script>

<section class="summary" aria-labelledby="projects-summary-title">
  <header class="head">
    <h2 id="projects-summary-title">Progetti</h2>
    {#if projects.length === 0}
      <p class="lead">Nessun Progetto. Aggiungi il primo dalla colonna.</p>
    {:else if total === 0}
      <p class="lead">Nessun Container in esecuzione. Scegli un Progetto dalla colonna.</p>
    {:else}
      <p class="lead">
        <span class="mono">{running}</span> Container in esecuzione su <span class="mono">{total}</span>. Scegli un Progetto
        dalla colonna.
      </p>
    {/if}
  </header>

  {#if projects.length > 0}
    <div class="counts">
      {#each counts as count (count.key)}
        <div class="cell">
          <span class="label">{count.label}</span>
          <span class="value num mono">{count.value}</span>
        </div>
      {/each}
    </div>
  {/if}
</section>

<style>
  .summary {
    display: grid;
    gap: var(--space-2);
  }
  .head {
    display: grid;
    gap: 6px;
  }
  h2 {
    font-size: var(--fs-title);
    font-weight: 600;
    line-height: 1.25;
    letter-spacing: -0.01em;
  }
  .lead {
    color: var(--muted);
    font-size: var(--fs-body);
  }
  .counts {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1px;
    border: 1px solid var(--border);
    border-radius: var(--radius-card);
    background: var(--border);
    overflow: hidden;
  }
  .cell {
    display: grid;
    gap: 6px;
    padding: var(--space-1);
    background: var(--surface);
  }
  .label {
    font-size: var(--fs-small);
    color: var(--muted);
  }
  .value {
    font-size: var(--fs-figure);
    font-weight: 500;
    line-height: 1;
  }
</style>
