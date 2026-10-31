<!-- The bottom drawer of the canvas: the Activities and the Log of the open
     Project, one tab at a time, with underline tabs. It stands in the guscio,
     below the scrolling content, so the output never takes room from the
     Servizi. The Log tab exists only with a stream. Closing the drawer hides
     it; it does not stop the stream. -->
<script lang="ts">
  import ActivityPanel from './ActivityPanel.svelte'
  import LogPanel from './LogPanel.svelte'
  import type { ActivityRecord, DrawerTab, LogView } from './store.svelte'

  let {
    tab,
    activities,
    selected,
    logs,
    services,
    ontab,
    onclose,
    onservice,
    oncloselogs,
  }: {
    /** The tab the store asks for. */
    tab: DrawerTab
    activities: ActivityRecord[]
    selected: ActivityRecord | null
    /** The open stream, already restricted to the Project on screen; null when there is none. */
    logs: LogView | null
    /** The Services of the open Project, for the log filter. */
    services: string[]
    ontab: (tab: DrawerTab) => void
    /** Hides the drawer without stopping the stream. */
    onclose: () => void
    /** Restarts the stream restricted to a Service ('' for all of them). */
    onservice: (service: string) => void
    /** Ends the stream. */
    oncloselogs: () => void
  } = $props()

  // The Log tab exists only with a stream: without one the drawer shows the Activities.
  const shown = $derived(tab === 'log' && logs === null ? 'attivita' : tab)
</script>

<section class="drawer" aria-label="Attività e Log">
  <div class="head">
    <div class="tabs" role="tablist" aria-label="Attività e Log">
      <button
        class="tab"
        id="tab-attivita"
        role="tab"
        aria-selected={shown === 'attivita'}
        aria-controls="pannello-attivita"
        onclick={() => ontab('attivita')}
      >Attività</button>
      {#if logs}
        <button
          class="tab"
          id="tab-log"
          role="tab"
          aria-selected={shown === 'log'}
          aria-controls="pannello-log"
          onclick={() => ontab('log')}
        >Log</button>
      {/if}
    </div>
    <button class="close" type="button" onclick={onclose}>Chiudi</button>
  </div>

  <div
    class="panel"
    id="pannello-attivita"
    role="tabpanel"
    aria-labelledby="tab-attivita"
    hidden={shown !== 'attivita'}
  >
    {#if shown === 'attivita'}
      <ActivityPanel {activities} {selected} />
    {/if}
  </div>

  {#if logs}
    <div class="panel" id="pannello-log" role="tabpanel" aria-labelledby="tab-log" hidden={shown !== 'log'}>
      {#if shown === 'log'}
        <LogPanel {logs} {services} {onservice} onclose={oncloselogs} />
      {/if}
    </div>
  {/if}
</section>

<style>
  .drawer {
    display: flex;
    flex: none;
    flex-direction: column;
    height: var(--drawer-height);
    min-height: 0;
    border-top: 1px solid var(--border);
    background: var(--surface);
  }

  /* underline tabs: the active one is the only one that carries the accent */
  .head {
    display: flex;
    flex: none;
    align-items: stretch;
    height: var(--control-height);
    padding: 0 var(--space-2);
    border-bottom: 1px solid var(--border);
  }
  .tabs {
    display: flex;
    gap: var(--space-2);
    margin-bottom: -1px;
  }
  .tab {
    padding: 0;
    border: 0;
    border-bottom: 2px solid transparent;
    background: transparent;
    color: var(--muted);
    font-family: var(--font-text);
    font-size: var(--fs-body);
    line-height: 1;
    cursor: pointer;
    transition: color 120ms ease, border-color 120ms ease;
  }
  .tab:hover:not(:disabled) {
    color: var(--text);
  }
  .tab[aria-selected='true'] {
    color: var(--text);
    border-bottom-color: var(--accent);
  }

  /* hiding the drawer never stops the stream; ending the stream is "Chiudi log" */
  .close {
    align-self: center;
    height: var(--control-height);
    margin-left: auto;
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
  .close:hover {
    background: var(--raised);
    color: var(--text);
  }
  .close:active {
    background: var(--active);
  }

  /* the panel takes the room left by the tabs */
  .panel {
    display: grid;
    flex: 1;
    grid-template-rows: minmax(0, 1fr);
    min-height: 0;
    padding: var(--space-1) var(--space-2);
  }
  .panel[hidden] {
    display: none;
  }
</style>
