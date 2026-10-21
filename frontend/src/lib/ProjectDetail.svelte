<script lang="ts">
  import type { Project } from '../../bindings/dockercontroller/internal/compose/models'
  import { Status } from '../../bindings/dockercontroller/internal/compose/models'
  import ConfirmModal from './ConfirmModal.svelte'
  import LogPanel from './LogPanel.svelte'
  import { ActionKind, actionLabels } from './store.svelte'
  import type { LogView } from './store.svelte'

  let {
    project,
    onaction,
    onopenport,
    onopenfolder,
    onopenterminal,
    onback,
    onremove,
    onsettings,
    onprofiles,
    groups,
    logs,
    onopenlogs,
    oncloselogs,
    unavailable = false,
  }: {
    project: Project
    onaction: (project: string, kind: ActionKind, service: string) => void
    /** Opens http://localhost:<port> in the browser. */
    onopenport: (port: number) => void
    onopenfolder: (project: string) => void
    /** Opens a terminal in the Project folder. */
    onopenterminal: (project: string) => void
    onback: () => void
    /** Removes the (Rimosso) Project from the saved list. */
    onremove: (project: string) => void
    /** Saves the Display name and Group of the Project ('' clears one). */
    onsettings: (project: string, displayName: string, group: string) => void
    /** Saves the complete set of active Profiles; it launches nothing. */
    onprofiles: (project: string, profiles: string[]) => void
    /** The existing Groups. */
    groups: string[]
    /** The open log stream, null when closed. */
    logs: LogView | null
    /** Opens the log stream of the Project, restricted to a Service ('' for all). */
    onopenlogs: (project: string, service: string) => void
    oncloselogs: () => void
    /** Docker does not answer: Actions are disabled. */
    unavailable?: boolean
  } = $props()

  const kinds = [ActionKind.ActionStart, ActionKind.ActionStop, ActionKind.ActionRestart]
  // Generated bindings type arrays as nullable; the backend never sends null.
  const services = $derived(project.services ?? [])
  const profiles = $derived(project.profiles ?? [])
  const activeProfiles = $derived(project.activeProfiles ?? [])
  const busy = $derived(project.activity !== null)
  const removable = $derived(project.status === Status.StatusRemoved && project.total === 0)

  // The destructive Action waiting for a confirmation, if any.
  let pending: 'down' | 'downVolumes' | 'remove' | null = $state(null)
  const name = $derived(project.displayName || project.name)

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

  function confirm() {
    const what = pending
    pending = null
    if (what === 'down') onaction(project.name, ActionKind.ActionDown, '')
    else if (what === 'downVolumes') onaction(project.name, ActionKind.ActionDownVolumes, '')
    else if (what === 'remove') onremove(project.name)
  }
</script>

<section class="detail">
  <header>
    <button onclick={onback}>← Progetti</button>
    <h2>{project.displayName || project.name}</h2>
    <button onclick={() => onopenfolder(project.name)} title={project.folder}>Apri cartella</button>
    <button onclick={() => onopenterminal(project.name)} title="Apri un terminale in {project.folder}">Apri terminale</button>
  </header>
  <p class="folder">{project.folder}</p>
  {#if project.displayName}<p class="compose-name">Nome compose: {project.name}</p>{/if}
  <p class="folder">{project.group ? `Gruppo: ${project.group}` : 'Nessun Gruppo'}</p>
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
    <button class="edit" onclick={startEditing}>Modifica nome e Gruppo</button>
  {/if}
  {#if project.composeMissing}
    <p class="missing">File compose mancante: i file del Progetto non esistono più. Puoi toglierlo dalla lista.</p>
  {/if}
  {#if profiles.length > 0}
    <fieldset class="profiles">
      <legend>Profili</legend>
      {#each profiles as profile (profile)}
        <label>
          <input
            type="checkbox"
            checked={activeProfiles.includes(profile)}
            onchange={(e) => toggleProfile(profile, e.currentTarget.checked)}
          />
          {profile}
        </label>
      {/each}
      <p class="hint">Vengono usati da Up, Stop, Restart, Pull e Build. Cambiarli non avvia né ferma nulla. Down agisce su tutti i Profili.</p>
    </fieldset>
  {/if}
  <div class="destructive">
    <button disabled={unavailable || busy} onclick={() => (pending = 'down')}>Down</button>
    <button class="danger" disabled={unavailable || busy} onclick={() => (pending = 'downVolumes')}>Down con volumi</button>
    <button
      class:highlight={project.composeMissing && removable}
      disabled={unavailable || busy || !removable}
      title={removable ? 'Toglie il Progetto dalla lista senza toccare Docker' : 'Disponibile solo per un Progetto Rimosso'}
      onclick={() => (pending = 'remove')}
    >
      Rimuovi dalla lista
    </button>
  </div>
  {#if project.activity}
    <p class="busy">
      ⏳ {actionLabels[project.activity.kind] ?? project.activity.kind}
      {project.activity.service ? `su ${project.activity.service} ` : ''}in corso
    </p>
  {/if}

  {#if logs && logs.project === project.name}
    <LogPanel
      {logs}
      services={services.map((s) => s.name)}
      onservice={(service) => onopenlogs(project.name, service)}
      onclose={oncloselogs}
    />
  {:else}
    <button class="open-logs" onclick={() => onopenlogs(project.name, '')}>Mostra log</button>
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
  header { display: flex; align-items: center; gap: 0.75rem; }
  h2 { flex: 1; font-size: 1.1rem; margin: 0; }
  h3 { font-size: 0.95rem; margin: 0; }
  .folder { color: var(--muted); font-size: 0.8rem; margin: 0.25rem 0 0.75rem; }
  .compose-name { color: var(--muted); font-size: 0.8rem; margin: 0; }
  .edit { margin-bottom: 0.75rem; }
  .settings { display: grid; gap: 0.5rem; max-width: 24rem; margin-bottom: 0.75rem; padding: 0.6rem 0.9rem; border-radius: 8px; background: var(--surface); border: 1px solid var(--border); }
  .settings label { display: grid; gap: 0.2rem; font-size: 0.85rem; }
  .settings input, .settings select { font: inherit; padding: 0.25rem 0.4rem; border-radius: 6px; border: 1px solid var(--border); background: var(--surface); color: var(--text); }
  .form-actions { display: flex; gap: 0.4rem; }
  .profiles { display: flex; flex-wrap: wrap; gap: 0.4rem 1rem; margin: 0 0 0.75rem; padding: 0.4rem 0.9rem; border-radius: 8px; border: 1px solid var(--border); background: var(--surface); }
  .profiles label { display: flex; align-items: center; gap: 0.35rem; font-size: 0.85rem; }
  .profiles .hint { flex-basis: 100%; margin: 0; color: var(--muted); font-size: 0.75rem; }
  .busy { color: var(--warn); font-size: 0.85rem; }
  .empty { color: var(--muted); }
  .missing { padding: 0.5rem 0.8rem; border-radius: 8px; background: var(--warn-bg); color: var(--warn); font-size: 0.85rem; }
  .open-logs { margin-bottom: 0.75rem; }
  .destructive { display: flex; gap: 0.4rem; margin-bottom: 0.75rem; }
  button.danger { background: var(--danger); border-color: var(--danger); color: #fff; }
  button.highlight { border-color: var(--warn); color: var(--warn); font-weight: 600; box-shadow: 0 0 0 2px var(--warn-bg); }
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
