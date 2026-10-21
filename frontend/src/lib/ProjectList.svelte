<script lang="ts">
  import { Status } from '../../bindings/dockercontroller/internal/compose/models'
  import type { Project } from '../../bindings/dockercontroller/internal/compose/models'
  import { ActionKind, actionLabels } from './store.svelte'

  let {
    projects,
    onaction,
    onselect,
    unavailable = false,
  }: {
    projects: Project[]
    onaction: (project: string, kind: ActionKind) => void
    /** Opens the detail of a Project. */
    onselect: (project: string) => void
    /** Docker does not answer: the list is greyed out and Actions are disabled. */
    unavailable?: boolean
  } = $props()

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
  <ul class="list" class:unavailable>
    {#each projects as project (project.name)}
      <li class="row">
        <div class="main">
          <button class="name" title="Apri il dettaglio" onclick={() => onselect(project.name)}>
            {project.displayName || project.name}
          </button>
          <span class="folder">{project.folder}</span>
          {#if project.composeMissing}<span class="missing">File compose mancante</span>{/if}
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
            <button disabled={unavailable || project.activity !== null} onclick={() => onaction(project.name, kind)}>
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
  button.name { font-weight: 600; font-size: 1rem; align-self: flex-start; border: none; background: none; padding: 0; text-align: left; }
  button.name:hover { text-decoration: underline; }
  .folder { color: var(--muted); font-size: 0.8rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .missing { color: var(--warn); font-size: 0.8rem; }
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
  .unavailable { opacity: 0.5; filter: grayscale(1); }
  .empty { color: var(--muted); }
</style>
