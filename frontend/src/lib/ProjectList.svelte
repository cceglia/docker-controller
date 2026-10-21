<script lang="ts">
  import { Status } from '../../bindings/dockercontroller/internal/compose/models'
  import type { Project } from '../../bindings/dockercontroller/internal/compose/models'
  import { ActionKind, actionLabels } from './store.svelte'

  let {
    projects,
    onaction,
  }: { projects: Project[]; onaction: (project: string, kind: ActionKind) => void } = $props()

  const kinds = [
    ActionKind.ActionUp,
    ActionKind.ActionStop,
    ActionKind.ActionRestart,
    ActionKind.ActionPull,
    ActionKind.ActionBuild,
  ]

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
        <div class="actions">
          {#if project.activity}
            <span class="busy">⏳ {actionLabels[project.activity.kind] ?? project.activity.kind} in corso</span>
          {/if}
          {#each kinds as kind (kind)}
            <button disabled={project.activity !== null} onclick={() => onaction(project.name, kind)}>
              {actionLabels[kind]}
            </button>
          {/each}
        </div>
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
  .actions { display: flex; align-items: center; gap: 0.3rem; }
  .busy { color: var(--warn); font-size: 0.8rem; margin-right: 0.3rem; }
  button {
    font: inherit; font-size: 0.8rem; padding: 0.2rem 0.55rem; border-radius: 6px;
    border: 1px solid var(--border); background: var(--surface); color: var(--text); cursor: pointer;
  }
  button:disabled { opacity: 0.45; cursor: not-allowed; }
  .empty { color: var(--muted); }
</style>
