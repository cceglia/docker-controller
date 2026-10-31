<!-- The open Project. The identity block carries the name, the status as a dot
     plus text, the strip of its Container, its compose name and its folder, and
     the Project Actions: one primary that follows the status (Up while Fermo or
     Rimosso, Stop while Attivo, Restart while Parziale) and the others quiet.
     Down, Down con volumi and Rimuovi dalla lista sit behind the "…" menu with
     the same confirmations as before. Below: the cells of Cartella, Nome
     compose and Gruppo with their commands, then the Profili as boxes.
     Below: one flat block per Servizio and one row per Container. The log
     panel lives in the bottom drawer of the guscio, not here. -->
<script lang="ts">
  import type { Project } from '../../bindings/dockercontroller/internal/compose/models'
  import type { ServiceUsage } from '../../bindings/dockercontroller/internal/compose/models'
  import { Status } from '../../bindings/dockercontroller/internal/compose/models'
  import ConfirmModal from './ConfirmModal.svelte'
  import ContainerStrip from './ContainerStrip.svelte'
  import ProjectMenu from './ProjectMenu.svelte'
  import ServiceBlock from './ServiceBlock.svelte'
  import { projectLabel, projectTone } from './status'
  import StatusDot from './StatusDot.svelte'
  import { ActionKind, actionLabels } from './store.svelte'

  let {
    project,
    onaction,
    onopenport,
    onopenfolder,
    onopenterminal,
    onremove,
    onsettings,
    onprofiles,
    groups,
    onopenlogs,
    unavailable = false,
    usage,
  }: {
    project: Project
    onaction: (project: string, kind: ActionKind, service: string) => void
    /** Opens http://localhost:<port> in the browser. */
    onopenport: (port: number) => void
    onopenfolder: (project: string) => void
    /** Opens a terminal in the Project folder. */
    onopenterminal: (project: string) => void
    /** Removes the (Rimosso) Project from the saved list. */
    onremove: (project: string) => void
    /** Saves the Display name and Group of the Project ('' clears one). */
    onsettings: (project: string, displayName: string, group: string) => void
    /** Saves the complete set of active Profiles; it launches nothing. */
    onprofiles: (project: string, profiles: string[]) => void
    /** The existing Groups. */
    groups: string[]
    /** Opens the log stream of the Project in the drawer, restricted to a Service ('' for all). */
    onopenlogs: (project: string, service: string) => void
    /** Docker does not answer: Actions are disabled. */
    unavailable?: boolean
    /** Utilizzo risorse by Service name: measures and rates. A Service without running Containers is absent. null until the first sample. */
    usage: Record<string, ServiceUsage> | null
  } = $props()

  // The Actions of the Project. The Actions of a Servizio live in ServiceBlock.
  const projectKinds = [
    ActionKind.ActionUp,
    ActionKind.ActionStop,
    ActionKind.ActionRestart,
    ActionKind.ActionPull,
    ActionKind.ActionBuild,
  ]

  // Generated bindings type arrays as nullable; the backend never sends null.
  const services = $derived(project.services ?? [])
  const profiles = $derived(project.profiles ?? [])
  const activeProfiles = $derived(project.activeProfiles ?? [])
  const containers = $derived(services.flatMap((s) => s.containers ?? []))
  const busy = $derived(project.activity !== null)
  const removable = $derived(project.status === Status.StatusRemoved && project.total === 0)
  const statusLabel = $derived(projectLabel(project.status))

  // One Action leads; which one follows the status. The others stay available.
  const primary = $derived(
    project.status === Status.StatusActive
      ? ActionKind.ActionStop
      : project.status === Status.StatusPartial
        ? ActionKind.ActionRestart
        : ActionKind.ActionUp,
  )
  const others = $derived(projectKinds.filter((kind) => kind !== primary))

  // The destructive Action waiting for a confirmation, if any.
  let pending: 'down' | 'downVolumes' | 'remove' | null = $state(null)
  // The Service whose Down is waiting for a confirmation, if any.
  let pendingService: string | null = $state(null)
  const name = $derived(project.displayName || project.name)

  // Choosing a menu item opens its confirmation on its own; the menu closes itself.
  function choose(what: 'down' | 'downVolumes' | 'remove') {
    pending = what
  }

  // Settings form: edited values, reset whenever another Project is opened.
  const NEW_GROUP = '\u0000new'
  let editing = $state(false)
  let displayNameInput = $state('')
  let groupChoice = $state('')
  let newGroupInput = $state('')

  function startEditing() {
    displayNameInput = project.displayName
    groupChoice = project.group
    newGroupInput = ''
    editing = true
  }

  function saveSettings() {
    const group = groupChoice === NEW_GROUP ? newGroupInput : groupChoice
    editing = false
    onsettings(project.name, displayNameInput, group)
  }

  function toggleProfile(profile: string, on: boolean) {
    const rest = activeProfiles.filter((p) => p !== profile)
    onprofiles(project.name, on ? [...rest, profile] : rest)
  }

  function confirmDownService() {
    const service = pendingService
    pendingService = null
    if (service) onaction(project.name, ActionKind.ActionDownService, service)
  }

  function confirm() {
    const what = pending
    pending = null
    if (what === 'down') onaction(project.name, ActionKind.ActionDown, '')
    else if (what === 'downVolumes') onaction(project.name, ActionKind.ActionDownVolumes, '')
    else if (what === 'remove') onremove(project.name)
  }
</script>

<section class="detail">
  <header class="identity">
    <div class="head-main">
      <div class="title-cell">
        <h2>{name}</h2>
        <ContainerStrip {containers} />
      </div>

      <p class="status-line">
        <StatusDot tone={projectTone(project.status)} />
        <span>{statusLabel}</span>
        <span class="count mono" title="Container in esecuzione sul totale">{project.running}/{project.total} in esecuzione</span>
      </p>

      <p class="ids">
        <span class="mono">{project.name}</span>
        <span class="mono path" title={project.folder}>{project.folder}</span>
      </p>
    </div>

    <div class="commands">
      <button class="primary" disabled={unavailable || busy} onclick={() => onaction(project.name, primary, '')}>
        {actionLabels[primary]}
      </button>
      {#each others as kind (kind)}
        <button class="quiet" disabled={unavailable || busy} onclick={() => onaction(project.name, kind, '')}>
          {actionLabels[kind]}
        </button>
      {/each}

      <ProjectMenu {unavailable} {busy} {removable} onchoose={choose} />
    </div>
  </header>

  {#if project.activity}
    <p class="activity">
      <span class="activity-dot" aria-hidden="true"></span>
      <span>
        {actionLabels[project.activity.kind] ?? project.activity.kind}{project.activity.service
          ? ` su ${project.activity.service}`
          : ''} in corso
      </span>
    </p>
  {/if}

  {#if project.composeMissing}
    <div class="notice">
      <p>
        File compose mancante: i file del Progetto non esistono più.{removable ? ' Puoi toglierlo dalla lista.' : ''}
      </p>
      {#if removable}
        <button class="notice-action" disabled={unavailable || busy} onclick={() => (pending = 'remove')}>
          Rimuovi dalla lista
        </button>
      {/if}
    </div>
  {/if}

  <section class="data" aria-label="Dati del Progetto">
    {#if editing}
      <form class="settings" onsubmit={(e) => { e.preventDefault(); saveSettings() }}>
        <label>
          Nome visualizzato
          <input bind:value={displayNameInput} placeholder={project.name} />
        </label>
        <label>
          Gruppo
          <select bind:value={groupChoice}>
            <option value="">Nessun Gruppo</option>
            {#each groups as g (g)}
              <option value={g}>{g}</option>
            {/each}
            <option value={NEW_GROUP}>Nuovo Gruppo…</option>
          </select>
        </label>
        {#if groupChoice === NEW_GROUP}
          <label>
            Nome del nuovo Gruppo
            <input bind:value={newGroupInput} />
          </label>
        {/if}
        <div class="form-actions">
          <button type="submit" disabled={groupChoice === NEW_GROUP && newGroupInput.trim() === ''}>Salva</button>
          <button type="button" onclick={() => (editing = false)}>Annulla</button>
        </div>
      </form>
    {:else}
      <dl class="grid">
        <div class="cell wide">
          <dt>Cartella</dt>
          <dd class="mono" title={project.folder}>{project.folder}</dd>
        </div>
        <div class="cell">
          <dt>Nome compose</dt>
          <dd class="mono">{project.name}</dd>
        </div>
        <div class="cell">
          <dt>Gruppo</dt>
          <dd>{project.group || 'Nessun Gruppo'}</dd>
        </div>
      </dl>
    {/if}

    <div class="cell-commands">
      <button
        class="quiet"
        aria-label="Apri la cartella nel gestore file"
        title="Apri la cartella nel gestore file"
        onclick={() => onopenfolder(project.name)}
      >Cartella</button>
      <button
        class="quiet"
        aria-label="Apri un terminale in {project.folder}"
        title="Apri un terminale in {project.folder}"
        onclick={() => onopenterminal(project.name)}
      >Terminale</button>
      {#if !editing}
        <button class="quiet" aria-label="Modifica nome e Gruppo" title="Modifica nome e Gruppo" onclick={startEditing}>Modifica</button>
      {/if}
    </div>
  </section>

  {#if profiles.length > 0}
    <section class="profiles">
      <h3 class="label">Profili</h3>
      <div class="boxes">
        {#each profiles as profile (profile)}
          <label class="box">
            <input
              type="checkbox"
              checked={activeProfiles.includes(profile)}
              onchange={(e) => toggleProfile(profile, e.currentTarget.checked)}
            />
            <span>{profile}</span>
          </label>
        {/each}
      </div>
      <p class="hint">Vengono usati da Up, Stop, Restart, Pull e Build. Cambiarli non avvia né ferma nulla. Down agisce su tutti i Profili.</p>
    </section>
  {/if}

  <div class="section-head">
    <h3 class="section-title">Servizi <span class="section-count mono">{services.length}</span></h3>
    <button class="quiet" onclick={() => onopenlogs(project.name, '')}>Mostra log</button>
  </div>

  {#if services.length === 0}
    <p class="empty">
      {#if unavailable}
        Nessun Servizio: Docker non risponde, quindi l'elenco non è disponibile.
      {:else if project.composeMissing}
        Nessun Servizio: i file compose del Progetto non esistono più.
      {:else}
        Nessun Servizio: Docker non riporta Container per questo Progetto. Avvialo con Up per crearli.
      {/if}
    </p>
  {/if}
  {#each services as service (service.name)}
    <ServiceBlock
      {service}
      usage={usage?.[service.name] ?? null}
      waiting={usage === null}
      {unavailable}
      {busy}
      onaction={(kind) => onaction(project.name, kind, service.name)}
      ondown={() => (pendingService = service.name)}
      {onopenport}
    />
  {/each}
</section>

{#if pendingService}
  <ConfirmModal
    title="Down di {pendingService}"
    message="Ferma e rimuove i Container del Servizio. I volumi restano."
    confirmLabel="Down"
    onconfirm={confirmDownService}
    oncancel={() => (pendingService = null)}
  />
{/if}

{#if pending === 'down'}
  <ConfirmModal
    title="Down di {name}"
    message="Ferma e rimuove i Container del Progetto. I volumi restano. Il Progetto resta in lista come Rimosso."
    confirmLabel="Down"
    onconfirm={confirm}
    oncancel={() => (pending = null)}
  />
{:else if pending === 'downVolumes'}
  <ConfirmModal
    title="Down con volumi di {name}"
    message="Rimuove i Container e ELIMINA i volumi del Progetto: i dati nei volumi andranno persi per sempre."
    confirmLabel="Down con volumi"
    confirmText={project.name}
    danger
    onconfirm={confirm}
    oncancel={() => (pending = null)}
  />
{:else if pending === 'remove'}
  <ConfirmModal
    title="Rimuovi {name} dalla lista"
    message="Il Progetto sparisce dalla lista dell'app. Docker e i file non vengono toccati."
    confirmLabel="Rimuovi dalla lista"
    onconfirm={confirm}
    oncancel={() => (pending = null)}
  />
{/if}

<style>
  .detail {
    display: grid;
    gap: var(--space-2);
  }

  /* ---------- identity: name, strip, status, compose name and folder, Actions ---------- */
  .identity {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-1);
    padding-bottom: var(--space-1);
    border-bottom: 1px solid var(--border);
  }
  .head-main {
    display: grid;
    flex: 1 1 18rem;
    gap: 4px;
    min-width: 0;
  }
  .title-cell {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px var(--space-1);
    min-width: 0;
  }
  h2 {
    font-size: var(--fs-title);
    font-weight: 600;
    line-height: 1.25;
    letter-spacing: -0.01em;
    overflow-wrap: anywhere;
  }
  .status-line {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: var(--fs-body);
  }
  .count {
    color: var(--muted);
    font-size: var(--fs-small);
  }
  .ids {
    display: flex;
    flex-wrap: wrap;
    gap: 0 var(--space-1);
    min-width: 0;
    color: var(--muted);
    font-size: var(--fs-small);
  }
  .path {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .commands {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px;
  }

  /* one Action leads; the others are text, so their weight stays low */
  .primary {
    height: var(--control-height);
    padding: 0 var(--space-1);
    border: 0;
    border-radius: var(--radius-control);
    background: var(--primary);
    color: var(--primary-ink);
    font-family: var(--font-text);
    font-size: var(--fs-body);
    font-weight: 500;
    line-height: 1;
    cursor: pointer;
  }
  .primary:not(:disabled):hover {
    opacity: 0.9;
  }
  .primary:not(:disabled):active {
    opacity: 0.8;
  }
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

  /* ---------- Activity and the missing compose files ---------- */
  .activity {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: var(--fs-body);
  }
  .activity-dot {
    flex: none;
    box-sizing: border-box;
    width: var(--dot-size);
    height: var(--dot-size);
    border-radius: var(--dot-size);
    background: var(--accent);
  }
  .notice {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-1);
    padding: var(--space-1);
    border-radius: var(--radius-card);
    background: var(--pill-warning-bg);
    color: var(--pill-warning-ink);
    font-size: var(--fs-body);
  }
  /* the text takes the room and wraps; the button keeps its place on the right */
  .notice p {
    flex: 1 1 20rem;
    min-width: 0;
  }
  .notice-action {
    height: var(--control-height);
    padding: 0 var(--space-1);
    border-radius: var(--radius-control);
    border: 1px solid currentColor;
    background: transparent;
    color: inherit;
    font-family: var(--font-text);
    font-size: var(--fs-body);
    cursor: pointer;
  }

  /* ---------- the cells: Cartella, Nome compose, Gruppo ---------- */
  .data {
    display: grid;
    gap: 4px;
  }
  .grid {
    display: flex;
    flex-wrap: wrap;
    gap: 1px;
    margin: 0;
    border: 1px solid var(--border);
    border-radius: var(--radius-card);
    background: var(--border);
    overflow: hidden;
  }
  .cell {
    display: grid;
    flex: 1 1 10rem;
    gap: 4px;
    min-width: 0;
    padding: var(--space-1);
    background: var(--surface);
  }
  /* a folder is the longest value: it takes twice the room of the others */
  .cell.wide {
    flex: 2 1 18rem;
  }
  dt {
    color: var(--muted);
    font-size: var(--fs-small);
  }
  dd {
    min-width: 0;
    overflow-wrap: anywhere;
    font-size: var(--fs-body);
  }
  .cell-commands {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }
  .settings {
    display: grid;
    gap: 8px;
    width: min(24rem, 100%);
  }
  .settings label {
    display: grid;
    gap: 4px;
    font-size: var(--fs-small);
    color: var(--muted);
  }
  .settings input,
  .settings select {
    min-height: var(--control-height);
    padding: 0 8px;
    border-radius: var(--radius-control);
    border: 1px solid var(--control);
    background: var(--surface);
    color: var(--text);
    font-family: var(--font-text);
    font-size: var(--fs-body);
  }
  .form-actions {
    display: flex;
    gap: 8px;
  }

  /* ---------- Profili: boxes that change nothing until an Action runs ---------- */
  .profiles {
    display: grid;
    gap: 8px;
  }
  .label {
    font-size: var(--fs-small);
    font-weight: 500;
    color: var(--muted);
  }
  .boxes {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .box {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-height: var(--control-height);
    padding: 0 var(--space-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    background: var(--surface);
    font-size: var(--fs-body);
    cursor: pointer;
  }
  .box input {
    margin: 0;
    accent-color: var(--accent);
  }
  .hint {
    color: var(--muted);
    font-size: var(--fs-small);
  }

  /* ---------- Servizi: flat blocks, one hairline apart ---------- */
  h3 {
    font-size: var(--fs-section);
    margin: 0;
  }
  .empty {
    padding: var(--space-1) 0;
    border-top: 1px solid var(--border);
    color: var(--muted);
    font-size: var(--fs-body);
  }
  .section-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-1);
    margin-top: 4px;
  }
  .section-title {
    font-size: var(--fs-section);
    font-weight: 600;
  }
  .section-count {
    margin-left: 4px;
    color: var(--muted);
    font-size: var(--fs-small);
    font-weight: 400;
  }
  /* The shared control: the Settings form buttons use it; the rest sets its own. */
  button {
    height: var(--control-height);
    padding: 0 var(--space-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    background: var(--surface);
    color: var(--text);
    font-family: var(--font-text);
    font-size: var(--fs-body);
    line-height: 1;
    cursor: pointer;
  }
  button:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
</style>
