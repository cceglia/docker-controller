<!-- The two-panel guscio: the column with the mark, the commands and the
     navigation, the canvas with the global Utilizzo risorse on top and the
     current screen below. It is presentational: App.svelte owns the store and
     passes the content in. Below 900px the column keeps only the mark and the
     icons; every command keeps an aria-label. -->
<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { UsageSample } from '../../bindings/dockercontroller/internal/compose/models'
  import Mark from './Mark.svelte'
  import UsageBar from './UsageBar.svelte'

  let {
    usage,
    onaddproject,
    onprojects,
    projectsCurrent,
    onsettings,
    navigation,
    drawer,
    children,
  }: {
    /** The latest global sample; null until the stream delivers one. */
    usage: UsageSample | null
    onaddproject: () => void
    /** Shows the summary: no Project open and the Impostazioni screen closed. */
    onprojects: () => void
    /** The canvas shows the summary. */
    projectsCurrent: boolean
    onsettings: () => void
    /** The Progetti of the column. */
    navigation: Snippet
    /** The bottom drawer: Activities and Log, under the scrolling content. */
    drawer: Snippet
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

    <button class="command" onclick={onaddproject} aria-label="Aggiungi progetto">
      <svg
        class="icon"
        viewBox="0 0 16 16"
        width="16"
        height="16"
        aria-hidden="true"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
      >
        <path d="M8 3v10M3 8h10" />
      </svg>
      <span class="command-label">Aggiungi progetto</span>
    </button>

    <!-- The way to the summary: always there, with no saved Progetto and on the
         Impostazioni screen too. It sits under the command and above the list. -->
    <button
      class="command"
      aria-current={projectsCurrent ? 'page' : undefined}
      aria-label="Progetti"
      onclick={onprojects}
    >
      <svg
        class="icon"
        viewBox="0 0 16 16"
        width="16"
        height="16"
        aria-hidden="true"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
      >
        <path d="M2.5 4h1M6 4h7.5M2.5 8h1M6 8h7.5M2.5 12h1M6 12h7.5" />
      </svg>
      <span class="command-label">Progetti</span>
    </button>

    <div class="nav">
      {@render navigation()}
    </div>

    <button class="command bottom" onclick={onsettings} aria-label="Impostazioni">
      <svg
        class="icon"
        viewBox="0 0 16 16"
        width="16"
        height="16"
        aria-hidden="true"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
      >
        <path d="M2.5 5h2.5M8.5 5h5M2.5 11h5M11 11h2.5" />
        <circle cx="6.5" cy="5" r="1.5" />
        <circle cx="9.5" cy="11" r="1.5" />
      </svg>
      <span class="command-label">Impostazioni</span>
    </button>
  </aside>

  <div class="canvas">
    <UsageBar {usage} />
    <div class="content">
      {@render children()}
    </div>
    {@render drawer()}
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

  .command {
    display: flex;
    flex: none;
    align-items: center;
    gap: 8px;
    width: 100%;
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
  .command.bottom {
    margin-top: 4px;
  }
  /* The Progetti entry of the summary: a tone step and a heavier name, the same
     grammar as the open row of the list, never colour alone. */
  .command[aria-current='page'] {
    background: var(--active);
    color: var(--text);
    font-weight: 600;
  }
  .icon {
    flex: none;
  }

  /* the Progetti take the room between the two commands and scroll on their own */
  .nav {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    margin: 4px 0;
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

  @media (max-width: 900px) {
    .shell {
      grid-template-columns: var(--sidebar-narrow-width) minmax(0, 1fr);
    }
    .sidebar {
      padding: 4px;
    }
    .brand {
      justify-content: center;
      padding: 0;
    }
    /* the wordmark leaves the row but stays the window's h1 */
    .wordmark {
      position: absolute;
      width: 1px;
      height: 1px;
      margin: -1px;
      overflow: hidden;
      clip-path: inset(50%);
      white-space: nowrap;
    }
    .command {
      justify-content: center;
      padding: 0;
    }
    .command-label {
      display: none;
    }
  }
</style>
