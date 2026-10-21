<script lang="ts">
  import { CloseBehavior } from './store.svelte'

  let {
    onchoose,
    oncancel,
  }: {
    /** The user chose what the closing does; `remember` saves the choice as the setting. */
    onchoose: (choice: CloseBehavior, remember: boolean) => void
    /** The dialog was dismissed: the window stays open. */
    oncancel: () => void
  } = $props()

  let remember = $state(false)

  /** Opens the dialog as a modal as soon as it is mounted. */
  const showModal = (dialog: HTMLDialogElement) => {
    dialog.showModal()
  }
</script>

<dialog {@attach showModal} {oncancel} aria-labelledby="close-title">
  <h3 id="close-title">Chiudere la finestra?</h3>
  <p>Puoi nascondere la finestra nel tray (i Progetti restano sotto controllo) oppure chiudere l'app.</p>
  <label class="remember">
    <input type="checkbox" bind:checked={remember} />
    Ricorda la scelta
  </label>
  <div class="buttons">
    <button onclick={oncancel}>Annulla</button>
    <button onclick={() => onchoose(CloseBehavior.CloseHide, remember)}>Nascondi nel tray</button>
    <button class="danger" onclick={() => onchoose(CloseBehavior.CloseQuit, remember)}>Esci</button>
  </div>
</dialog>

<style>
  dialog {
    max-width: 30rem; border: 1px solid var(--border); border-radius: 10px;
    background: var(--surface); color: var(--text); padding: 1.1rem 1.25rem;
  }
  dialog::backdrop { background: rgb(0 0 0 / 0.5); }
  h3 { margin: 0 0 0.5rem; font-size: 1rem; }
  p { margin: 0 0 0.75rem; font-size: 0.9rem; }
  .remember { display: flex; align-items: center; gap: 0.4rem; font-size: 0.85rem; margin-bottom: 0.75rem; }
  .buttons { display: flex; justify-content: flex-end; gap: 0.5rem; }
  button {
    font: inherit; font-size: 0.85rem; padding: 0.3rem 0.8rem; border-radius: 6px;
    border: 1px solid var(--border); background: var(--surface); color: var(--text); cursor: pointer;
  }
  button.danger { background: var(--danger); border-color: var(--danger); color: #fff; }
</style>
