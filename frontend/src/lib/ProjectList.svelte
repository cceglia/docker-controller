<!-- The Progetti as the column's navigation: a row per Project with its status
     dot, its Display name, its compose name and the strip of its Container.
     Choosing a row opens the Project on the canvas. The open row keeps a tone
     step and a heavier name, never colour alone, and carries aria-current.
     A Project whose compose files are gone carries a compact warning pill.
     Below 900px the column keeps the dot and the strips; the row keeps its
     accessible name, count and warning through aria-label and its tooltip. -->
<script lang="ts">
  import type { Project } from '../../bindings/dockercontroller/internal/compose/models'
  import { SvelteMap } from 'svelte/reactivity'
  import ContainerStrip from './ContainerStrip.svelte'
  import { projectLabel, projectTone } from './status'
  import StatusDot from './StatusDot.svelte'

  let {
    projects,
    selected = null,
    onselect,
  }: {
    projects: Project[]
    /** The compose name of the open Project; null when none is open. */
    selected?: string | null
    /** Opens the detail of a Project. */
    onselect: (project: string) => void
  } = $props()

  // Projects by Group (Groups sorted), then those without a Group in a section of their own.
  const sections = $derived.by(() => {
    const byGroup = new SvelteMap<string, Project[]>()
    const ungrouped: Project[] = []
    for (const p of projects) {
      if (!p.group) ungrouped.push(p)
      else byGroup.set(p.group, [...(byGroup.get(p.group) ?? []), p])
    }
    const named = [...byGroup.entries()]
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([title, items]) => ({ title, items }))
    // Without any Group the list stays flat, with no section heading.
    if (named.length === 0) return [{ title: '', items: ungrouped }]
    return ungrouped.length > 0 ? [...named, { title: 'Senza Gruppo', items: ungrouped }] : named
  })

  // The generated bindings type the arrays as nullable; the backend never sends null.
  const containersOf = (project: Project) => (project.services ?? []).flatMap((s) => s.containers ?? [])

  const titleOf = (project: Project) =>
    project.displayName && project.displayName !== project.name
      ? `${project.displayName} (${project.name})`
      : project.name

  // The accessible name of a row: it survives the narrow column, which hides
  // the names, the count and the warning pill.
  const labelOf = (project: Project) => {
    const name = project.displayName || project.name
    const status = projectLabel(project.status)
    const count = `${project.running} Container in esecuzione su ${project.total}`
    return `${name}, ${status}, ${count}${project.composeMissing ? ', File compose mancante' : ''}`
  }
</script>

{#if projects.length === 0}
  <p class="empty">Nessun Progetto.</p>
{:else}
  <nav aria-label="Progetti">
    {#each sections as section (section.title)}
      <section class="group">
        {#if section.title}<h2 class="group-title">{section.title}</h2>{/if}
        <ul class="rows">
          {#each section.items as project (project.name)}
            {@const open = project.name === selected}
            {@const name = project.displayName || project.name}
            <li>
              <button
                class="row"
                class:open
                aria-current={open ? 'page' : undefined}
                aria-label={labelOf(project)}
                title={titleOf(project)}
                onclick={() => onselect(project.name)}
              >
                <span class="dot-cell"><StatusDot tone={projectTone(project.status)} /></span>
                <span class="name-cell">
                  <span class="name">{name}</span>
                  {#if project.composeMissing}
                    <span
                      class="pill"
                      title="File compose mancante: i file del Progetto non esistono più."
                    >File mancante</span>
                  {/if}
                </span>
                <span class="count mono" aria-hidden="true">{project.running}/{project.total}</span>
                <span class="compose mono">{project.name}</span>
                <span class="strip-cell"><ContainerStrip containers={containersOf(project)} compact /></span>
              </button>
            </li>
          {/each}
        </ul>
      </section>
    {/each}
  </nav>
{/if}

<style>
  nav {
    display: grid;
    gap: 8px;
  }
  .empty {
    padding: 0 8px;
    color: var(--muted);
    font-size: var(--fs-body);
  }
  .group-title {
    margin: 6px 8px 4px;
    font-size: var(--fs-small);
    font-weight: 500;
    color: var(--muted);
  }
  .rows {
    display: grid;
    gap: 1px;
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .row {
    display: grid;
    grid-template-columns: var(--dot-size) minmax(0, 1fr) auto;
    grid-template-areas:
      'dot name count'
      'dot compose strip';
    column-gap: 10px;
    row-gap: 1px;
    align-items: center;
    box-sizing: border-box;
    width: 100%;
    min-height: var(--row-height);
    padding: 4px 8px;
    border: 0;
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--text);
    font-family: var(--font-text);
    font-size: var(--fs-body);
    line-height: 1.3;
    text-align: left;
    cursor: pointer;
    transition: background-color 120ms ease;
  }
  .row:hover {
    background: var(--raised);
  }
  /* The open Project: a tone step and a heavier name, so it never relies on colour. */
  .row.open {
    background: var(--active);
  }
  .row.open .name {
    font-weight: 600;
  }
  .dot-cell {
    grid-area: dot;
    display: flex;
  }
  .name-cell {
    grid-area: name;
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }
  .name {
    flex: 1 1 auto;
    min-width: 0;
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  /* The pill marks a real problem only: a compose file that no longer exists. */
  .pill {
    flex: none;
    padding: 1px 4px;
    border-radius: var(--radius-control);
    background: var(--pill-warning-bg);
    color: var(--pill-warning-ink);
    font-size: var(--fs-small);
    font-weight: 500;
    line-height: 1.2;
    white-space: nowrap;
  }
  .compose {
    grid-area: compose;
    font-size: var(--fs-small);
    color: var(--muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .count {
    grid-area: count;
    justify-self: end;
    font-size: var(--fs-small);
    color: var(--muted);
  }
  .strip-cell {
    grid-area: strip;
    justify-self: end;
    display: flex;
  }

  /* Narrow column: the dot and the strips, centred on the row. */
  @media (max-width: 900px) {
    .row {
      grid-template-columns: var(--dot-size) minmax(0, 1fr);
      grid-template-areas: 'dot strip';
      justify-items: center;
      min-height: var(--control-height);
      padding: 6px 4px;
    }
    /* the names and the pill leave the row but stay in its accessible name */
    .name-cell,
    .compose {
      position: absolute;
      width: 1px;
      height: 1px;
      margin: -1px;
      overflow: hidden;
      clip-path: inset(50%);
      white-space: nowrap;
    }
    .count,
    .group-title {
      display: none;
    }
    .strip-cell {
      justify-self: center;
    }
  }
</style>
