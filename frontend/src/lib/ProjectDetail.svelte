<script lang="ts">
  import type { Project } from '../../bindings/dockercontroller/internal/compose/models'
  import { ActionKind, actionLabels } from './store.svelte'

  let {
    project,
    onaction,
    onopenport,
    onopenfolder,
    onback,
    unavailable = false,
  }: {
    project: Project
    onaction: (project: string, kind: ActionKind, service: string) => void
    /** Opens http://localhost:<port> in the browser. */
    onopenport: (port: number) => void
    onopenfolder: (project: string) => void
    onback: () => void
    /** Docker does not answer: Actions are disabled. */
    unavailable?: boolean
  } = $props()

  const kinds = [ActionKind.ActionStart, ActionKind.ActionStop, ActionKind.ActionRestart]
  // Generated bindings type arrays as nullable; the backend never sends null.
  const services = $derived(project.services ?? [])
  const busy = $derived(project.activity !== null)
</script>

<section class="detail">
  <header>
    <button onclick={onback}>← Progetti</button>
    <h2>{project.displayName || project.name}</h2>
    <button onclick={() => onopenfolder(project.name)} title={project.folder}>Apri cartella</button>
  </header>
  <p class="folder">{project.folder}</p>
  {#if project.activity}
    <p class="busy">
      ⏳ {actionLabels[project.activity.kind] ?? project.activity.kind}
      {project.activity.service ? `su ${project.activity.service} ` : ''}in corso
    </p>
  {/if}

  {#if services.length === 0}
    <p class="empty">Nessun Container: il Progetto non ha Servizi da mostrare.</p>
  {/if}
  {#each services as service (service.name)}
    <article class="service">
      <div class="head">
        <h3>{service.name}</h3>
        <div class="actions">
          {#each kinds as kind (kind)}
            <button disabled={unavailable || busy} onclick={() => onaction(project.name, kind, service.name)}>
              {actionLabels[kind]}
            </button>
          {/each}
        </div>
      </div>
      <ul class="containers">
        {#each (service.containers ?? []) as c (c.name)}
          <li>
            <span class="cname">{c.name}</span>
            <span class="image">{c.image}</span>
            <span class="state {c.state}">{c.state}</span>
            {#if c.health}<span class="health {c.health}">{c.health}</span>{/if}
            <span class="ports">
              {#each (c.ports ?? []) as port (`${port.published}/${port.protocol}`)}
                <button class="port" title="Apri http://localhost:{port.published}" onclick={() => onopenport(port.published)}>
                  {port.published}→{port.target}/{port.protocol}
                </button>
              {/each}
            </span>
          </li>
        {/each}
      </ul>
    </article>
  {/each}
</section>

<style>
  header { display: flex; align-items: center; gap: 0.75rem; }
  h2 { flex: 1; font-size: 1.1rem; margin: 0; }
  h3 { font-size: 0.95rem; margin: 0; }
  .folder { color: var(--muted); font-size: 0.8rem; margin: 0.25rem 0 0.75rem; }
  .busy { color: var(--warn); font-size: 0.85rem; }
  .empty { color: var(--muted); }
  .service { margin-bottom: 0.6rem; padding: 0.6rem 0.9rem; border-radius: 8px; background: var(--surface); border: 1px solid var(--border); }
  .head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 0.4rem; }
  .actions { display: flex; gap: 0.3rem; }
  .containers { list-style: none; margin: 0; padding: 0; display: grid; gap: 0.3rem; }
  .containers li { display: flex; flex-wrap: wrap; align-items: center; gap: 0.6rem; font-size: 0.82rem; }
  .cname { font-weight: 600; }
  .image { color: var(--muted); }
  .state.running, .health.healthy { color: var(--ok); }
  .state.exited, .state.dead, .health.unhealthy { color: var(--danger); }
  .health.starting, .state.restarting, .state.paused { color: var(--warn); }
  .ports { display: flex; flex-wrap: wrap; gap: 0.25rem; margin-left: auto; }
  button {
    font: inherit; font-size: 0.8rem; padding: 0.2rem 0.55rem; border-radius: 6px;
    border: 1px solid var(--border); background: var(--surface); color: var(--text); cursor: pointer;
  }
  button:disabled { opacity: 0.45; cursor: not-allowed; }
  .port { font-variant-numeric: tabular-nums; }
</style>
