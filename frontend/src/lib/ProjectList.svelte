<script lang="ts">
  import { Status } from '../../bindings/dockercontroller/internal/compose/models'
  import type { Project } from '../../bindings/dockercontroller/internal/compose/models'

  let { projects }: { projects: Project[] } = $props()

  const labels: Record<string, string> = {
    [Status.StatusActive]: 'Attivo',
    [Status.StatusPartial]: 'Parziale',
    [Status.StatusStopped]: 'Fermo',
    [Status.StatusRemoved]: 'Rimosso',
  }
</script>

{#if projects.length === 0}
  <p class="empty">Nessun Progetto.</p>
{:else}
  <ul class="list">
    {#each projects as project (project.name)}
      <li class="row">
        <div class="main">
          <span class="name">{project.displayName || project.name}</span>
          <span class="folder">{project.folder}</span>
        </div>
        <span class="count" title="Container in esecuzione sul totale">
          {project.running}/{project.total}
        </span>
        <span class="status {project.status}">{labels[project.status] ?? project.status}</span>
      </li>
    {/each}
  </ul>
{/if}

<style>
  .list { list-style: none; margin: 0; padding: 0; display: grid; gap: 0.5rem; }
  .row {
    display: flex; align-items: center; gap: 1rem;
    padding: 0.6rem 0.9rem; border-radius: 8px;
    background: var(--surface); border: 1px solid var(--border);
  }
  .main { flex: 1; min-width: 0; display: flex; flex-direction: column; }
  .name { font-weight: 600; }
  .folder { color: var(--muted); font-size: 0.8rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .count { font-variant-numeric: tabular-nums; color: var(--muted); }
  .status { min-width: 5.5rem; text-align: center; padding: 0.15rem 0.6rem; border-radius: 999px; font-size: 0.8rem; border: 1px solid currentColor; }
  .active { color: var(--ok); }
  .partial { color: var(--warn); }
  .stopped { color: var(--muted); }
  .removed { color: var(--muted); border-style: dashed; }
  .empty { color: var(--muted); }
</style>
