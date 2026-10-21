<script lang="ts">
  import { untrack } from 'svelte'
  import { CloseBehavior } from './store.svelte'
  import type { Settings } from './store.svelte'

  let {
    settings,
    onsave,
    onback,
  }: {
    settings: Settings
    /** Saves the settings; resolves to an error message, null on success. */
    onsave: (settings: Settings) => Promise<string | null>
    onback: () => void
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
  <header>
    <button onclick={onback}>← Progetti</button>
    <h2>Impostazioni</h2>
  </header>
  <form onsubmit={save}>
    <fieldset>
      <legend>Alla chiusura della finestra</legend>
      {#each closeOptions as option (option.value)}
        <label class="radio">
          <input type="radio" name="close" value={option.value} bind:group={closeBehavior} />
          {option.label}
        </label>
      {/each}
    </fieldset>
    <label>
      Comando del terminale
      <input bind:value={terminalCommand} spellcheck="false" autocomplete="off" />
      <span class="hint">{'{dir}'} viene sostituito con la cartella del Progetto. Vuoto per il comando predefinito.</span>
    </label>
    {#if error}<p class="error">{error}</p>{/if}
    {#if saved}<p class="saved">Impostazioni salvate.</p>{/if}
    <div class="buttons"><button type="submit">Salva</button></div>
  </form>
</section>

<style>
  header { display: flex; align-items: center; gap: 0.75rem; margin: 0 0 1rem; }
  h2 { font-size: 1.1rem; margin: 0; }
  form { display: grid; gap: 1rem; max-width: 36rem; }
  fieldset { border: 1px solid var(--border); border-radius: 8px; display: grid; gap: 0.4rem; margin: 0; }
  legend { font-size: 0.85rem; padding: 0 0.3rem; }
  label { display: grid; gap: 0.3rem; font-size: 0.85rem; }
  label.radio { display: flex; align-items: center; gap: 0.4rem; }
  input:not([type='radio']) { font: inherit; padding: 0.3rem 0.5rem; border-radius: 6px; border: 1px solid var(--border); background: var(--bg); color: var(--text); }
  .hint { color: var(--muted); font-size: 0.8rem; }
  .error { color: var(--danger); font-size: 0.85rem; margin: 0; }
  .saved { color: var(--ok); font-size: 0.85rem; margin: 0; }
  .buttons { display: flex; gap: 0.5rem; }
  button {
    font: inherit; font-size: 0.85rem; padding: 0.3rem 0.8rem; border-radius: 6px;
    border: 1px solid var(--border); background: var(--surface); color: var(--text); cursor: pointer;
  }
</style>
