<!-- The open Project. The identity block carries the name, the status as a dot
     plus text, the strip of its Container, its compose name and its folder, and
     the Project Actions: one primary that follows the status (Up while Fermo or
     Rimosso, Stop while Attivo, Restart while Parziale) and the others quiet.
     Down, Down con volumi and Rimuovi dalla lista sit behind the "…" menu with
     the same confirmations as before. Below: the cells of Cartella, Nome
     compose and Gruppo with their commands, then the Profili as boxes.
     Below: one flat block per Servizio, one row per Container, and the log
     panel inline until the drawer of ticket 06. -->
<script lang="ts">
  import type { Container, Project } from '../../bindings/dockercontroller/internal/compose/models'
  import { Status } from '../../bindings/dockercontroller/internal/compose/models'
  import ConfirmModal from './ConfirmModal.svelte'
  import ContainerStrip from './ContainerStrip.svelte'
  import LogPanel from './LogPanel.svelte'
  import { formatBytes, formatCores, formatPercent, formatRate } from './format'
  import Meter from './Meter.svelte'
  import { ActionKind, actionLabels } from './store.svelte'
  import type { LogView, ServiceUsageView } from './store.svelte'

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
    /** Utilizzo risorse by Service name: measures and rates. A Service without running Containers is absent. null until the first sample. */
    usage: Record<string, ServiceUsageView> | null
  } = $props()

  // The Actions of a Service stay on its own block; these are the Actions of the Project.
  const kinds = [ActionKind.ActionStart, ActionKind.ActionStop, ActionKind.ActionRestart]
  const projectKinds = [
    ActionKind.ActionUp,
    ActionKind.ActionStop,
    ActionKind.ActionRestart,
    ActionKind.ActionPull,
    ActionKind.ActionBuild,
  ]

  // Only these health states are a problem: they alone wear a pill.
  const isProblemHealth = (health: string) => health === 'unhealthy' || health === 'starting'

  // The dot of a Container follows the same grammar as the Project status.
  function dotTone(container: Container): 'positive' | 'warning' | 'negative' | 'stopped' {
    if (container.health === 'unhealthy' || container.state === 'dead') return 'negative'
    if (container.health === 'starting' || container.state === 'restarting' || container.state === 'paused') {
      return 'warning'
    }
    if (container.state === 'running') return 'positive'
    return 'stopped'
  }

  const statusLabels: Record<string, string> = {
    [Status.StatusActive]: 'Attivo',
    [Status.StatusPartial]: 'Parziale',
    [Status.StatusStopped]: 'Fermo',
    [Status.StatusRemoved]: 'Rimosso',
  }

  // Generated bindings type arrays as nullable; the backend never sends null.
  const services = $derived(project.services ?? [])
  const profiles = $derived(project.profiles ?? [])
  const activeProfiles = $derived(project.activeProfiles ?? [])
  const containers = $derived(services.flatMap((s) => s.containers ?? []))
  const busy = $derived(project.activity !== null)
  const removable = $derived(project.status === Status.StatusRemoved && project.total === 0)
  const statusLabel = $derived(statusLabels[project.status] ?? project.status)

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

  // The "…" menu: open state, its trigger, its panel and the box around both.
  let menuOpen = $state(false)
  let menuWrap = $state<HTMLElement | null>(null)
  let menuTrigger = $state<HTMLButtonElement | null>(null)
  let menuPanel = $state<HTMLElement | null>(null)

  function toggleMenu() {
    menuOpen = !menuOpen
  }

  // The menu items that can take the focus, in order.
  function menuItems(): HTMLButtonElement[] {
    return [...(menuPanel?.querySelectorAll<HTMLButtonElement>('[role="menuitem"]:not(:disabled)') ?? [])]
  }

  // Escape closes and returns the focus to the trigger; the arrows move between items.
  // It is bound to the trigger and to every item, so the focus is always on an
  // interactive element while the menu is open.
  function onMenuKeydown(event: KeyboardEvent) {
    if (!menuOpen) return
    if (event.key === 'Escape') {
      event.preventDefault()
      closeMenu()
      return
    }
    if (event.key !== 'ArrowDown' && event.key !== 'ArrowUp' && event.key !== 'Home' && event.key !== 'End') return
    const items = menuItems()
    if (items.length === 0) return
    const current = items.indexOf(document.activeElement as HTMLButtonElement)
    let next = 0
    if (event.key === 'ArrowDown') next = current < 0 ? 0 : (current + 1) % items.length
    else if (event.key === 'ArrowUp') next = current <= 0 ? items.length - 1 : current - 1
    else if (event.key === 'End') next = items.length - 1
    event.preventDefault()
    items[next].focus()
  }

  // ArrowDown on the closed trigger opens the menu on its first item.
  function onTriggerKeydown(event: KeyboardEvent) {
    if (event.key === 'ArrowDown' && !menuOpen) {
      event.preventDefault()
      menuOpen = true
      return
    }
    onMenuKeydown(event)
  }

  function closeMenu() {
    if (!menuOpen) return
    menuOpen = false
    menuTrigger?.focus()
  }

  // Choosing an item closes the menu; the confirmation, if any, then opens on its own.
  function choose(what: 'down' | 'downVolumes' | 'remove') {
    menuOpen = false
    menuTrigger?.focus()
    pending = what
  }

  // The focus that leaves the menu closes it: Tab then goes on from the trigger.
  function onMenuFocusOut(event: FocusEvent) {
    const next = event.relatedTarget as Node | null
    if (!next || menuWrap?.contains(next)) return
    menuOpen = false
  }

  // While the menu is open the first item takes the focus, and a click outside closes it.
  $effect(() => {
    if (!menuOpen) return
    menuItems()[0]?.focus()
    const outside = (event: MouseEvent) => {
      const target = event.target as Node | null
      if (target && menuWrap?.contains(target)) return
      menuOpen = false
    }
    document.addEventListener('click', outside)
    return () => document.removeEventListener('click', outside)
  })

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
        <span class="dot {project.status}" aria-hidden="true"></span>
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

      <div class="menu-wrap" bind:this={menuWrap} onfocusout={onMenuFocusOut}>
        <button
          class="menu-trigger"
          bind:this={menuTrigger}
          aria-haspopup="menu"
          aria-expanded={menuOpen}
          aria-controls={menuOpen ? 'project-menu' : undefined}
          aria-label="Altre azioni sul Progetto"
          title="Altre azioni sul Progetto"
          onclick={toggleMenu}
          onkeydown={onTriggerKeydown}
        >…</button>
        {#if menuOpen}
          <div class="menu" id="project-menu" role="menu" aria-label="Altre azioni sul Progetto" bind:this={menuPanel}>
            <button role="menuitem" class="menu-item" disabled={unavailable || busy} onkeydown={onMenuKeydown} onclick={() => choose('down')}>Down</button>
            <button
              role="menuitem"
              class="menu-item danger"
              disabled={unavailable || busy}
              onkeydown={onMenuKeydown}
              onclick={() => choose('downVolumes')}
            >
              Down con volumi
            </button>
            <button
              role="menuitem"
              class="menu-item"
              disabled={unavailable || busy || !removable}
              title={removable ? 'Toglie il Progetto dalla lista senza toccare Docker' : 'Disponibile solo per un Progetto Rimosso'}
              onkeydown={onMenuKeydown}
              onclick={() => choose('remove')}
            >
              Rimuovi dalla lista
            </button>
          </div>
        {/if}
      </div>
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
    {#if !(logs && logs.project === project.name)}
      <button class="quiet" onclick={() => onopenlogs(project.name, '')}>Mostra log</button>
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
    {@const serviceContainers = service.containers ?? []}
    <article class="service">
      <div class="service-head">
        <div class="service-title">
          <h4>{service.name}</h4>
          <ContainerStrip containers={serviceContainers} />
        </div>
        <div class="service-actions">
          {#each kinds as kind (kind)}
            <button class="quiet" disabled={unavailable || busy} onclick={() => onaction(project.name, kind, service.name)}>
              {actionLabels[kind]}
            </button>
          {/each}
          <button class="quiet danger" disabled={unavailable || busy} onclick={() => (pendingService = service.name)}>
            {actionLabels[ActionKind.ActionDownService]}
          </button>
        </div>
      </div>
      {#if usage?.[service.name]}
        {@const u = usage[service.name]}
        <div class="usage" aria-label="Utilizzo risorse di {service.name}">
          <div class="measure">
            <span class="measure-label">CPU</span>
            <span class="measure-value num">{formatPercent(u.measures.cpuPercent)}</span>
            <span class="measure-note num">{formatCores(u.measures.cpuPercent)}</span>
            <Meter value={u.measures.cpuPercent} label="CPU" tone="accent" wide />
          </div>
          <div class="measure">
            <span class="measure-label">Memoria</span>
            <span class="measure-value num">{formatBytes(u.measures.memoryBytes)}</span>
          </div>
          <div class="measure">
            <span class="measure-label">Rete</span>
            <span class="pair">
              <span class="reading"><span class="dir">in</span><span class="value num">{formatRate(u.rate.netInBytesPerSec)}</span></span>
              <span class="reading"><span class="dir">out</span><span class="value num">{formatRate(u.rate.netOutBytesPerSec)}</span></span>
            </span>
          </div>
          <div class="measure">
            <span class="measure-label">Disco</span>
            <span class="pair">
              <span class="reading"><span class="dir">lettura</span><span class="value num">{formatRate(u.rate.blockReadBytesPerSec)}</span></span>
              <span class="reading"><span class="dir">scrittura</span><span class="value num">{formatRate(u.rate.blockWriteBytesPerSec)}</span></span>
            </span>
          </div>
        </div>
      {:else if usage === null && serviceContainers.some((c) => c.state === 'running')}
        <p class="usage-none" aria-label="Utilizzo risorse di {service.name}" title="In attesa della prima misura">In attesa della prima misura…</p>
      {:else}
        <p class="usage-none" aria-label="Utilizzo risorse di {service.name}" title="Nessun Container in esecuzione">Utilizzo risorse: —</p>
      {/if}
      <ul class="containers">
        {#each serviceContainers as c (c.name)}
          <li class="container">
            <span class="dot {dotTone(c)}" aria-hidden="true"></span>
            <span class="cname mono" title={c.name}>{c.name}</span>
            <span class="image mono" title={c.image}>{c.image}</span>
            <span class="state">{c.state}</span>
            {#if c.health}
              {#if isProblemHealth(c.health)}
                <span class="pill {c.health === 'unhealthy' ? 'negative' : 'warning'}">{c.health}</span>
              {:else}
                <span class="health">{c.health}</span>
              {/if}
            {/if}
            {#if (c.ports ?? []).length > 0}
              <span class="ports">
                {#each (c.ports ?? []) as port (`${port.published}/${port.protocol}`)}
                  <button
                    class="port"
                    title="Apri http://localhost:{port.published}"
                    aria-label="Apri http://localhost:{port.published} nel browser: porta {port.target}/{port.protocol}"
                    onclick={() => onopenport(port.published)}
                  >{port.published}→{port.target}/{port.protocol}</button>
                {/each}
              </span>
            {/if}
          </li>
        {/each}
      </ul>
    </article>
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
  .dot {
    flex: none;
    box-sizing: border-box;
    width: var(--dot-size);
    height: var(--dot-size);
    border-radius: var(--dot-size);
  }
  .dot.active {
    background: var(--positive);
  }
  .dot.partial {
    background: var(--warning);
  }
  .dot.stopped {
    background: var(--muted);
  }
  .dot.removed {
    border: 1px solid var(--control);
    background: transparent;
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

  /* the "…" menu: flat, one tone step above the surface, closed by choice or by a click outside */
  .menu-wrap {
    position: relative;
    display: flex;
  }
  .menu-trigger {
    width: var(--control-height);
    height: var(--control-height);
    padding: 0;
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
  .menu-trigger:hover,
  .menu-trigger[aria-expanded='true'] {
    background: var(--raised);
    color: var(--text);
  }
  .menu {
    position: absolute;
    top: calc(var(--control-height) + 4px);
    right: 0;
    z-index: 10;
    display: grid;
    min-width: 12rem;
    padding: 4px;
    border: 1px solid var(--border);
    border-radius: var(--radius-card);
    background: var(--raised);
  }
  .menu-item {
    display: flex;
    align-items: center;
    min-height: var(--control-height);
    padding: 0 var(--space-1);
    border: 0;
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--text);
    font-family: var(--font-text);
    font-size: var(--fs-body);
    text-align: left;
    cursor: pointer;
  }
  .menu-item:not(:disabled):hover {
    background: var(--active);
  }
  .menu-item.danger {
    color: var(--negative);
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
  .service {
    display: grid;
    gap: 8px;
    padding: var(--space-1) 0 0;
    border-top: 1px solid var(--border);
  }
  .service-head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 4px var(--space-1);
  }
  .service-title {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px var(--space-1);
    min-width: 0;
  }
  .service-title h4 {
    font-size: var(--fs-section);
    font-weight: 600;
    overflow-wrap: anywhere;
  }
  .service-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }
  .service-actions .danger {
    color: var(--negative);
  }
  .service-actions .danger:not(:disabled):hover {
    color: var(--negative);
  }

  /* Utilizzo risorse: one cell per measure, an instant value or a rate. */
  .usage {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr));
    gap: var(--space-1);
    padding: var(--space-1) 0;
    border-top: 1px solid var(--rule);
  }
  .measure {
    display: grid;
    align-content: start;
    gap: 2px;
    min-width: 0;
    padding-left: var(--space-1);
    border-left: 1px solid var(--rule);
  }
  .measure:first-child {
    padding-left: 0;
    border-left: 0;
  }
  .measure-label {
    color: var(--muted);
    font-size: var(--fs-small);
    line-height: 1.3;
  }
  .measure-value,
  .measure-note,
  .reading .value {
    font-family: var(--font-mono);
    font-variant-numeric: tabular-nums;
    line-height: 1.4;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .measure-value {
    font-size: var(--fs-body);
    font-weight: 500;
  }
  .measure-note {
    color: var(--muted);
    font-size: var(--fs-small);
    line-height: 1.3;
  }
  .pair {
    display: grid;
    gap: 1px;
    min-width: 0;
  }
  .reading {
    display: flex;
    align-items: baseline;
    gap: 6px;
    min-width: 0;
  }
  .dir {
    flex: none;
    min-width: 5ch;
    color: var(--muted);
    font-size: var(--fs-small);
    line-height: 1.3;
  }
  .reading .value {
    font-size: var(--fs-small);
  }
  .usage-none {
    margin: 0;
    padding: var(--space-1) 0;
    border-top: 1px solid var(--rule);
    color: var(--muted);
    font-size: var(--fs-body);
  }

  /* Container rows: dot, name, image, state, health, published ports. */
  .containers {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
  }
  .container {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px var(--space-1);
    min-height: var(--row-height);
    padding: 6px 0;
    border-top: 1px solid var(--rule);
    font-size: var(--fs-small);
  }
  .cname {
    max-width: 100%;
    font-size: var(--fs-body);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .image {
    flex: 1 1 10rem;
    min-width: 0;
    color: var(--muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .state,
  .health {
    color: var(--muted);
  }
  /* A pill marks a problem only: unhealthy, or a healthcheck still starting. */
  .pill {
    flex: none;
    padding: 1px 4px;
    border-radius: var(--radius-control);
    font-size: var(--fs-small);
    font-weight: 500;
    line-height: 1.2;
    white-space: nowrap;
  }
  .pill.negative {
    background: var(--pill-negative-bg);
    color: var(--pill-negative-ink);
  }
  .pill.warning {
    background: var(--pill-warning-bg);
    color: var(--pill-warning-ink);
  }
  .ports {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    margin-left: auto;
  }
  .port {
    height: var(--control-height);
    padding: 0 8px;
    border: 1px solid var(--control);
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--text);
    font-family: var(--font-mono);
    font-size: var(--fs-small);
    font-variant-numeric: tabular-nums;
    line-height: 1;
    cursor: pointer;
    transition: background-color 120ms ease;
  }
  .port:hover {
    background: var(--raised);
  }
  .port:active {
    background: var(--active);
  }

  /* The dots of the Container rows share the grammar of the Project status. */
  .dot.positive {
    background: var(--positive);
  }
  .dot.warning {
    background: var(--warning);
  }
  .dot.negative {
    background: var(--negative);
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
