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

<section class="logs" aria-label="Log">
  <div class="bar">
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
    <p class="count">{visible.length} righe{term !== '' ? ` su ${logs.lines.length}` : ''} (massimo 5000)</p>
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
</section>

<style>
  /* The drawer gives the room: the bar takes what it needs, the lines fill the rest. */
  .logs {
    display: grid;
    grid-template-rows: auto minmax(0, 1fr);
    gap: 8px;
    height: 100%;
    min-height: 0;
  }
  .bar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px var(--space-1);
    font-size: var(--fs-small);
  }
  .bar label {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .bar input[type='search'],
  .bar select {
    min-height: var(--control-height);
    padding: 0 8px;
    border: 1px solid var(--control);
    border-radius: var(--radius-control);
    background: var(--surface);
    color: var(--text);
    font-family: var(--font-text);
    font-size: var(--fs-body);
  }
  .bar input[type='search'] {
    flex: 1 1 12rem;
    min-width: 0;
  }
  button {
    height: var(--control-height);
    padding: 0 8px;
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    background: var(--surface);
    color: var(--text);
    font-family: var(--font-text);
    font-size: var(--fs-body);
    line-height: 1;
    cursor: pointer;
  }
  button:hover {
    background: var(--raised);
  }
  button.on {
    border-color: var(--warning);
    color: var(--warning);
  }
  .count {
    margin-left: auto;
    color: var(--muted);
    font-size: var(--fs-small);
  }
  .body {
    min-height: 0;
    padding: 8px;
    border: 1px solid var(--rule);
    border-radius: var(--radius-control);
    background: var(--ground);
    font-family: var(--font-mono);
    font-size: var(--fs-small);
    line-height: 1.4;
    overflow: auto;
  }
  .line {
    display: flex;
    gap: var(--space-1);
    white-space: pre-wrap;
    word-break: break-all;
  }
  .ts {
    flex: none;
    color: var(--muted);
    font-variant-numeric: tabular-nums;
  }
  .svc {
    flex: none;
    font-weight: 600;
  }
  .text {
    flex: 1;
  }
  .err .text {
    color: var(--warning);
  }
  .empty,
  .ended {
    margin: 0;
    font-family: var(--font-text);
  }
  .empty {
    color: var(--muted);
  }
  .ended {
    color: var(--warning);
    font-weight: 600;
  }
</style>
