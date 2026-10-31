<script lang="ts">
  import { untrack } from 'svelte'
  import { CloseBehavior } from './store.svelte'
  import type { Settings } from './store.svelte'

  let {
    settings,
    onsave,
  }: {
    settings: Settings
    /** Saves the settings; resolves to an error message, null on success. */
    onsave: (settings: Settings) => Promise<string | null>
  } = $props()

  const closeOptions = [
    { value: CloseBehavior.CloseHide, label: 'Nascondi nel tray' },
    { value: CloseBehavior.CloseQuit, label: 'Esci' },
    { value: CloseBehavior.CloseAsk, label: 'Chiedi ogni volta' },
  ]

  // Edited values: the form starts from the saved settings and is then its own
  // state (the screen is mounted only once they are loaded).
  let closeBehavior = $state(untrack(() => settings.closeBehavior))
  let terminalCommand = $state(untrack(() => settings.terminalCommand))
  let error = $state('')
  let saved = $state(false)

  async function save(e: SubmitEvent) {
    e.preventDefault()
    saved = false
    error = (await onsave({ closeBehavior, terminalCommand })) ?? ''
    saved = error === ''
  }
</script>

<section class="settings">
  <h2 class="head">Impostazioni</h2>
  <form onsubmit={save}>
    <div class="row" role="radiogroup" aria-labelledby="close-title">
      <div class="label">
        <span class="title" id="close-title">Alla chiusura della finestra</span>
        <p class="description">Scegli cosa fa l'app quando chiudi la finestra.</p>
      </div>
      <div class="panel">
        {#each closeOptions as option (option.value)}
          <label class="radio">
            <input type="radio" name="close" value={option.value} bind:group={closeBehavior} />
            {option.label}
          </label>
        {/each}
      </div>
    </div>

    <div class="row">
      <div class="label">
        <label class="title" for="terminal-command">Comando del terminale</label>
        <p class="description">{'{dir}'} viene sostituito con la cartella del Progetto. Vuoto per il comando predefinito.</p>
      </div>
      <div class="panel">
        <input id="terminal-command" bind:value={terminalCommand} spellcheck="false" autocomplete="off" />
      </div>
    </div>

    <div class="foot">
      <span class="status">
        {#if error}<span class="error" role="alert">{error}</span>{/if}
        {#if saved}<span class="saved" role="status">Impostazioni salvate.</span>{/if}
      </span>
      <button type="submit" class="primary">Salva</button>
    </div>
  </form>
</section>

<style>
  .settings {
    display: grid;
    gap: var(--space-2);
  }
  .head {
    font-size: var(--fs-title);
    font-weight: 600;
    letter-spacing: -0.01em;
  }

  /* one row per setting: label and explanation left, control in a panel right */
  form {
    display: grid;
  }
  .row {
    display: grid;
    grid-template-columns: minmax(0, 2fr) minmax(0, 3fr);
    gap: var(--space-2);
    padding: var(--space-2) 0;
    border-top: 1px solid var(--rule);
  }
  .row:first-of-type {
    padding-top: var(--space-1);
    border-top: 0;
  }
  .label {
    display: grid;
    align-content: start;
    gap: 6px;
  }
  .title {
    font-size: var(--fs-section);
    font-weight: 500;
    line-height: 1.3;
  }
  .description {
    color: var(--muted);
    font-size: var(--fs-body);
    line-height: 1.45;
  }

  .panel {
    display: grid;
    align-content: start;
    gap: 4px;
    padding: var(--space-1);
    border-radius: var(--radius-card);
    background: var(--raised);
  }
  .radio {
    display: flex;
    align-items: center;
    gap: 8px;
    min-height: var(--control-height);
    font-size: var(--fs-body);
    cursor: pointer;
  }
  .panel input:not([type='radio']) {
    width: 100%;
    height: var(--control-height);
    padding: 0 var(--space-1);
    border: 1px solid var(--control);
    border-radius: var(--radius-control);
    background: var(--surface);
    color: var(--text);
    font-size: var(--fs-body);
  }

  /* one save model for the view: the outcome sits beside the Save that produces it */
  .foot {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: var(--space-1);
    padding-top: var(--space-2);
    border-top: 1px solid var(--rule);
  }
  .status {
    flex: 1;
    min-width: 0;
  }
  .error {
    color: var(--negative);
    font-size: var(--fs-body);
  }
  .saved {
    color: var(--positive);
    font-size: var(--fs-body);
  }
  .primary {
    height: var(--control-height);
    padding: 0 var(--space-2);
    border: 0;
    border-radius: var(--radius-control);
    background: var(--primary);
    color: var(--primary-ink);
    font-size: var(--fs-body);
    font-weight: 500;
    cursor: pointer;
  }
  .primary:hover {
    opacity: 0.9;
  }
  .primary:active {
    opacity: 0.8;
  }

  @media (max-width: 640px) {
    .row {
      grid-template-columns: minmax(0, 1fr);
      gap: var(--space-1);
    }
  }
</style>
