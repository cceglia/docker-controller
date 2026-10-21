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

<section class="panel" id="attivita">
  <h2>Attività</h2>
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
        <pre class="output" bind:this={output}>{#each selected.lines as line}<span class:err={line.stderr}>{line.text}
</span>{/each}{#if selected.state === 'error'}<span class="fail">{selected.error}
</span>{/if}</pre>
      {/if}
    </div>
  {/if}
</section>

<style>
  .panel { margin-top: 1.5rem; }
  h2 { font-size: 1rem; margin: 0 0 0.5rem; }
  .body { display: grid; grid-template-columns: 14rem 1fr; gap: 0.75rem; }
  .history { list-style: none; margin: 0; padding: 0; display: grid; gap: 0.25rem; align-content: start; max-height: 16rem; overflow: auto; }
  .history button {
    width: 100%; text-align: left; font: inherit; font-size: 0.8rem; padding: 0.3rem 0.5rem;
    border-radius: 6px; border: 1px solid var(--border); background: var(--surface); color: var(--text); cursor: pointer;
  }
  .history button.current { border-color: var(--ok); }
  .output {
    margin: 0; height: 16rem; overflow: auto; padding: 0.6rem; border-radius: 8px; font-size: 0.78rem;
    background: var(--surface); border: 1px solid var(--border); white-space: pre-wrap; word-break: break-all;
  }
  .err { color: var(--warn); }
  .fail { color: var(--danger); font-weight: 600; }
  .empty { color: var(--muted); }
</style>
