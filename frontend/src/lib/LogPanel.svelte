<script lang="ts">
  import { serviceColor } from './store.svelte'
  import type { LogView } from './store.svelte'

  let {
    logs,
    services,
    onservice,
    onclose,
  }: {
    logs: LogView
    /** The Services of the Project, for the filter. */
    services: string[]
    /** Restarts the stream restricted to a Service ('' for all of them). */
    onservice: (service: string) => void
    onclose: () => void
  } = $props()

  // View state: it does not change what is streamed.
  let paused = $state(false)
  let search = $state('')
  let showTimestamps = $state(true)

  const term = $derived(search.trim().toLowerCase())
  const visible = $derived(term === '' ? logs.lines : logs.lines.filter((l) => l.text.toLowerCase().includes(term)))

  // Attachment: follows the newest line unless the scrolling is paused. It
  // runs again whenever the visible lines or the pause change.
  function follow(node: HTMLElement) {
    void visible.length
    if (!paused) node.scrollTop = node.scrollHeight
  }

  function time(timestamp: string): string {
    const d = new Date(timestamp)
    return Number.isNaN(d.getTime()) ? '' : d.toLocaleTimeString('it-IT')
  }
</script>

<section class="logs">
  <div class="bar">
    <h3>Log</h3>
    <label>
      Servizio
      <select value={logs.service} onchange={(e) => onservice(e.currentTarget.value)}>
        <option value="">Tutti</option>
        {#each services as s (s)}
          <option value={s}>{s}</option>
        {/each}
      </select>
    </label>
    <input type="search" placeholder="Cerca nel testo" bind:value={search} />
    <label><input type="checkbox" bind:checked={showTimestamps} /> Timestamp</label>
    <button class:on={paused} onclick={() => (paused = !paused)}>{paused ? 'Riprendi scorrimento' : 'Pausa scorrimento'}</button>
    <button onclick={onclose}>Chiudi log</button>
  </div>
  <div class="body" {@attach follow}>
    {#each visible as line (line.seq)}
      <div class="line" class:err={line.stderr}>
        {#if showTimestamps && line.timestamp}<span class="ts">{time(line.timestamp)}</span>{/if}
        {#if line.service}<span class="svc" style:color={serviceColor(line.service)}>{line.service}</span>{/if}
        <span class="text">{line.text}</span>
      </div>
    {/each}
    {#if logs.lines.length === 0 && logs.live}
      <p class="empty">In attesa di righe…</p>
    {:else if term !== '' && visible.length === 0}
      <p class="empty">Nessuna riga corrisponde alla ricerca.</p>
    {/if}
    {#if !logs.live}
      <p class="ended">{logs.error ? `Flusso interrotto: ${logs.error}` : 'Il flusso dei log è terminato.'}</p>
    {/if}
  </div>
  <p class="count">{visible.length} righe{term !== '' ? ` su ${logs.lines.length}` : ''} (massimo 5000)</p>
</section>

<style>
  .logs { margin-top: 0.75rem; padding: 0.6rem 0.9rem; border-radius: 8px; background: var(--surface); border: 1px solid var(--border); }
  .bar { display: flex; flex-wrap: wrap; align-items: center; gap: 0.5rem 0.75rem; margin-bottom: 0.5rem; font-size: 0.85rem; }
  h3 { font-size: 0.95rem; margin: 0; flex: 1; }
  .bar label { display: flex; align-items: center; gap: 0.3rem; }
  .bar input[type='search'], .bar select { font: inherit; padding: 0.2rem 0.4rem; border-radius: 6px; border: 1px solid var(--border); background: var(--surface); color: var(--text); }
  button {
    font: inherit; font-size: 0.8rem; padding: 0.2rem 0.55rem; border-radius: 6px;
    border: 1px solid var(--border); background: var(--surface); color: var(--text); cursor: pointer;
  }
  button.on { border-color: var(--warn); color: var(--warn); }
  .body { height: 20rem; overflow: auto; padding: 0.4rem; border-radius: 6px; border: 1px solid var(--border); font: 0.76rem/1.35 ui-monospace, monospace; }
  .line { display: flex; gap: 0.6rem; white-space: pre-wrap; word-break: break-all; }
  .ts { color: var(--muted); flex: none; }
  .svc { flex: none; font-weight: 600; }
  .text { flex: 1; }
  .err .text { color: var(--warn); }
  .empty, .count { color: var(--muted); font-size: 0.78rem; margin: 0.3rem 0 0; }
  .ended { color: var(--warn); font-weight: 600; }
</style>
