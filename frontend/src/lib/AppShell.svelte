<!-- The two-panel guscio: the column with the mark and the two commands, the
     canvas with the global Utilizzo risorse on top and the current screen below.
     It is presentational: App.svelte owns the store and passes the content in. -->
<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { UsageSample } from '../../bindings/dockercontroller/internal/compose/models'
  import Mark from './Mark.svelte'
  import UsageBar from './UsageBar.svelte'

  let {
    usage,
    onaddproject,
    onsettings,
    children,
  }: {
    /** The latest global sample; null until the stream delivers one. */
    usage: UsageSample | null
    onaddproject: () => void
    onsettings: () => void
    /** The screen shown in the canvas. */
    children: Snippet
  } = $props()
</script>

<div class="shell">
  <aside class="sidebar" aria-label="Docker Controller">
    <div class="brand">
      <Mark size={20} />
      <h1 class="wordmark">Docker Controller</h1>
    </div>

    <div class="commands">
      <button class="command" onclick={onaddproject}>Aggiungi progetto</button>
      <button class="command" onclick={onsettings}>Impostazioni</button>
    </div>
  </aside>

  <div class="canvas">
    <UsageBar {usage} />
    <div class="content">
      {@render children()}
    </div>
  </div>
</div>

<style>
  .shell {
    display: grid;
    grid-template-columns: var(--sidebar-width) minmax(0, 1fr);
    height: 100%;
    min-height: 0;
  }

  .sidebar {
    display: flex;
    flex-direction: column;
    min-height: 0;
    padding: var(--space-1);
    border-right: 1px solid var(--border);
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: var(--control-height);
    padding: 0 4px;
    color: var(--text);
  }
  .wordmark {
    font-family: var(--font-text);
    font-size: var(--fs-section);
    font-weight: 600;
    line-height: 1.3;
    letter-spacing: -0.01em;
    white-space: nowrap;
  }

  /* the two commands keep their final place at the foot of the column */
  .commands {
    display: grid;
    gap: 2px;
    margin-top: auto;
  }
  .command {
    display: flex;
    align-items: center;
    height: var(--control-height);
    padding: 0 var(--space-1);
    border: 0;
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--muted);
    font-family: var(--font-text);
    font-size: var(--fs-body);
    line-height: 1.4;
    text-align: left;
    cursor: pointer;
    transition: background-color 120ms ease, color 120ms ease;
  }
  .command:hover {
    background: var(--raised);
    color: var(--text);
  }
  .command:active {
    background: var(--active);
  }

  .canvas {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
    background: var(--ground);
  }
  .content {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: var(--space-2);
  }
</style>
