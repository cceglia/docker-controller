<script lang="ts">
  import type { Measures, Project } from '../../bindings/dockercontroller/internal/compose/models'
  import { Status } from '../../bindings/dockercontroller/internal/compose/models'
  import ConfirmModal from './ConfirmModal.svelte'
  import LogPanel from './LogPanel.svelte'
  import { formatBytes, formatCores, formatPercent } from './format'
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
    usage,
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
    /** Utilizzo risorse by Service name; a Service without running Containers is absent. null until the first sample. */
    usage: Record<string, Measures> | null
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
  // The Service whose Down is waiting for a confirmation, if any.
  let pendingService: string | null = $state(null)
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
  <header>
    <button onclick={onback}>← Progetti</button>
    <h2>{name}</h2>
    <span class="count" title="Container in esecuzione sul totale">{project.running}/{project.total} in esecuzione</span>
    <div class="header-actions">
      <button onclick={() => onopenfolder(project.name)} title={project.folder}>Apri cartella</button>
      <button onclick={() => onopenterminal(project.name)} title="Apri un terminale in {project.folder}">Apri terminale</button>
    </div>
  </header>

  {#if project.composeMissing}
    <p class="missing">File compose mancante: i file del Progetto non esistono più. Puoi toglierlo dalla lista.</p>
  {/if}
  {#if project.activity}
    <p class="busy">
      ⏳ {actionLabels[project.activity.kind] ?? project.activity.kind}
      {project.activity.service ? `su ${project.activity.service} ` : ''}in corso
    </p>
  {/if}

  <div class="card info">
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
      <dl>
        <dt>Cartella</dt><dd class="path">{project.folder}</dd>
        <dt>Nome compose</dt><dd>{project.name}</dd>
        <dt>Gruppo</dt><dd>{project.group || 'Nessun Gruppo'}</dd>
      </dl>
      <button class="edit" onclick={startEditing}>Modifica nome e Gruppo</button>
    {/if}
  </div>

  {#if profiles.length > 0}
    <fieldset class="card profiles">
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

  <div class="section-head">
    <h3 class="section-title">Servizi <span class="muted">{services.length}</span></h3>
    {#if !(logs && logs.project === project.name)}
      <button onclick={() => onopenlogs(project.name, '')}>Mostra log</button>
    {/if}
  </div>

  {#if logs && logs.project === project.name}
    <LogPanel
      {logs}
      services={services.map((s) => s.name)}
      onservice={(service) => onopenlogs(project.name, service)}
      onclose={oncloselogs}
    />
  {/if}

  {#if services.length === 0}
    <p class="empty">Nessun Container: il Progetto non ha Servizi da mostrare.</p>
  {/if}
  {#each services as service (service.name)}
    <article class="card service">
      <div class="head">
        <h3>{service.name}</h3>
        <div class="actions">
          {#each kinds as kind (kind)}
            <button disabled={unavailable || busy} onclick={() => onaction(project.name, kind, service.name)}>
              {actionLabels[kind]}
            </button>
          {/each}
          <button disabled={unavailable || busy} onclick={() => (pendingService = service.name)}>
            {actionLabels[ActionKind.ActionDownService]}
          </button>
        </div>
      </div>
      {#if usage?.[service.name]}
        {@const u = usage[service.name]}
        <div class="usage" aria-label="Utilizzo risorse di {service.name}">
          <span><i>CPU</i><b>{formatPercent(u.cpuPercent)}</b> <small>{formatCores(u.cpuPercent)}</small></span>
          <span><i>Memoria</i><b>{formatBytes(u.memoryBytes)}</b></span>
          <span title="Totale dall'avvio dei Container"><i>Rete</i><b>↓ {formatBytes(u.netInBytes, 'decimal')}</b> <b>↑ {formatBytes(u.netOutBytes, 'decimal')}</b></span>
          <span title="Totale dall'avvio dei Container"><i>Disco</i><b>R {formatBytes(u.blockReadBytes, 'decimal')}</b> <b>W {formatBytes(u.blockWriteBytes, 'decimal')}</b></span>
        </div>
      {:else if usage === null && (service.containers ?? []).some((c) => c.state === 'running')}
        <p class="usage-none" aria-label="Utilizzo risorse di {service.name}" title="In attesa della prima misura">Utilizzo risorse: caricamento…</p>
      {:else}
        <p class="usage-none" aria-label="Utilizzo risorse di {service.name}" title="Nessun Container in esecuzione">Utilizzo risorse: —</p>
      {/if}
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

  <div class="card destructive">
    <div>
      <h3 class="section-title">Azioni sul Progetto</h3>
      <p class="hint">Chiedono sempre conferma.</p>
    </div>
    <div class="buttons">
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
  </div>
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
  .detail { display: grid; gap: 0.9rem; }
  header { display: flex; align-items: center; gap: 0.75rem; }
  h2 { font-size: 1.15rem; margin: 0; }
  .count { color: var(--muted); font-size: 0.85rem; flex: 1; }
  .header-actions { display: flex; gap: 0.4rem; }
  h3 { font-size: 0.95rem; margin: 0; }
  .card { margin: 0; padding: 0.8rem 1rem; border-radius: 10px; background: var(--surface); border: 1px solid var(--border); box-shadow: var(--shadow); }
  .info { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; }
  dl { display: grid; grid-template-columns: max-content 1fr; gap: 0.3rem 1rem; margin: 0; font-size: 0.85rem; min-width: 0; }
  dt { color: var(--muted); }
  dd { margin: 0; min-width: 0; overflow-wrap: anywhere; }
  .path { font-family: ui-monospace, monospace; font-size: 0.8rem; }
  .settings { display: grid; gap: 0.5rem; width: min(24rem, 100%); }
  .settings label { display: grid; gap: 0.2rem; font-size: 0.85rem; }
  .settings input, .settings select { font: inherit; padding: 0.25rem 0.4rem; border-radius: 6px; border: 1px solid var(--border); background: var(--surface); color: var(--text); }
  .form-actions { display: flex; gap: 0.4rem; }
  .profiles { display: flex; flex-wrap: wrap; gap: 0.4rem 1rem; }
  .profiles legend { padding: 0 0.3rem; font-size: 0.75rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.08em; color: var(--muted); }
  .profiles label { display: flex; align-items: center; gap: 0.35rem; font-size: 0.85rem; }
  .hint { flex-basis: 100%; margin: 0; color: var(--muted); font-size: 0.75rem; }
  .busy { margin: 0; color: var(--warn); font-size: 0.85rem; }
  .empty { color: var(--muted); margin: 0; }
  .missing { margin: 0; padding: 0.5rem 0.8rem; border-radius: 8px; background: var(--warn-bg); color: var(--warn); font-size: 0.85rem; }
  .section-head { display: flex; align-items: center; justify-content: space-between; margin-top: 0.4rem; }
  .section-title { font-size: 0.75rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.08em; color: var(--muted); }
  .muted { font-weight: 400; }
  .service { display: grid; gap: 0.6rem; }
  .head { display: flex; align-items: center; justify-content: space-between; gap: 0.75rem; flex-wrap: wrap; }
  .actions { display: flex; gap: 0.3rem; }
  .usage { display: grid; grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr)); gap: 0.4rem 1rem; padding: 0.5rem 0.7rem; border-radius: 8px; background: var(--bg); font-size: 0.82rem; }
  .usage span { display: flex; flex-direction: column; gap: 0.1rem; }
  .usage i { font-style: normal; font-size: 0.68rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.08em; color: var(--muted); }
  .usage b { font-weight: 600; font-variant-numeric: tabular-nums; }
  .usage small { color: var(--muted); }
  .usage-none { margin: 0; color: var(--muted); font-size: 0.8rem; }
  .containers { list-style: none; margin: 0; padding: 0; display: grid; gap: 0.3rem; }
  .containers li { display: flex; flex-wrap: wrap; align-items: center; gap: 0.6rem; padding-top: 0.4rem; border-top: 1px solid var(--border); font-size: 0.82rem; }
  .cname { font-weight: 600; }
  .image { color: var(--muted); }
  .state.running, .health.healthy { color: var(--ok); }
  .state.exited, .state.dead, .health.unhealthy { color: var(--danger); }
  .health.starting, .state.restarting, .state.paused { color: var(--warn); }
  .ports { display: flex; flex-wrap: wrap; gap: 0.25rem; margin-left: auto; }
  .destructive { display: flex; align-items: center; justify-content: space-between; gap: 1rem; flex-wrap: wrap; }
  .destructive .hint { margin-top: 0.15rem; }
  .buttons { display: flex; gap: 0.4rem; flex-wrap: wrap; }
  button {
    font: inherit; font-size: 0.8rem; padding: 0.2rem 0.55rem; border-radius: 6px;
    border: 1px solid var(--border); background: var(--surface); color: var(--text); cursor: pointer;
  }
  button:disabled { opacity: 0.45; cursor: not-allowed; }
  .port { font-variant-numeric: tabular-nums; }
  button.danger { background: var(--danger); border-color: var(--danger); color: #fff; }
  button.danger:not(:disabled):hover { background: var(--danger); filter: brightness(1.1); }
  button.highlight { border-color: var(--warn); color: var(--warn); font-weight: 600; box-shadow: 0 0 0 2px var(--warn-bg); }
</style>
