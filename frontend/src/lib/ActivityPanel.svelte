<script lang="ts">
  import { actionLabels, selectActivity } from './store.svelte'
  import type { ActivityRecord } from './store.svelte'

  let { activities, selected }: { activities: ActivityRecord[]; selected: ActivityRecord | null } = $props()

  let output: HTMLPreElement | undefined = $state()

  // Follow the live output.
  $effect(() => {
    void selected?.lines.length
    if (output) output.scrollTop = output.scrollHeight
  })

  const stateLabel = { running: '⏳', ok: '✓', error: '✗' }
</script>

<section class="panel" id="attivita" aria-label="Attività">
  {#if activities.length === 0}
    <p class="empty">Nessuna Attività in questa sessione.</p>
  {:else}
    <div class="body">
      <ul class="history">
        {#each activities as a (a.id)}
          <li>
            <button class:current={selected?.id === a.id} onclick={() => selectActivity(a.id)}>
              {stateLabel[a.state]} {actionLabels[a.kind] ?? a.kind} · {a.project}
            </button>
          </li>
        {/each}
      </ul>
      {#if selected}
        <pre class="output" bind:this={output}>{#each selected.lines as line (line)}<span class:err={line.stderr}>{line.text}
</span>{/each}{#if selected.state === 'error'}<span class="fail">{selected.error}
</span>{/if}</pre>
      {/if}
    </div>
  {/if}
</section>

<style>
  /* The drawer gives the room: the history keeps its column, the output takes the rest. */
  .panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }
  .body {
    display: grid;
    flex: 1;
    grid-template-columns: 14rem minmax(0, 1fr);
    grid-template-rows: minmax(0, 1fr);
    gap: var(--space-1);
    min-height: 0;
  }
  .history {
    display: grid;
    gap: 4px;
    align-content: start;
    min-height: 0;
    margin: 0;
    padding: 0;
    overflow: auto;
    list-style: none;
  }
  .history button {
    width: 100%;
    padding: 4px 8px;
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    background: var(--surface);
    color: var(--text);
    font-family: var(--font-text);
    font-size: var(--fs-small);
    line-height: 1.4;
    text-align: left;
    cursor: pointer;
  }
  .history button:hover {
    background: var(--raised);
  }
  .history button.current {
    border-color: var(--positive);
  }
  .output {
    min-height: 0;
    margin: 0;
    padding: 8px;
    border: 1px solid var(--rule);
    border-radius: var(--radius-control);
    background: var(--ground);
    font-family: var(--font-mono);
    font-size: var(--fs-small);
    line-height: 1.4;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-all;
  }
  .err {
    color: var(--warning);
  }
  .fail {
    color: var(--negative);
    font-weight: 600;
  }
  .empty {
    color: var(--muted);
    font-size: var(--fs-body);
  }
</style>
