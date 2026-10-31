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
    <button class="quiet" onclick={oncancel}>Annulla</button>
    <button class="primary" onclick={() => onchoose(CloseBehavior.CloseHide, remember)}>Nascondi nel tray</button>
    <button class="quiet danger" onclick={() => onchoose(CloseBehavior.CloseQuit, remember)}>Esci</button>
  </div>
</dialog>

<style>
  dialog {
    width: min(30rem, 90vw);
    padding: var(--space-2);
    border: 1px solid var(--border);
    border-radius: var(--radius-card);
    background: var(--surface);
    color: var(--text);
  }
  dialog::backdrop {
    background: var(--backdrop);
  }
  h3 {
    margin: 0 0 var(--space-1);
    font-size: var(--fs-section);
    font-weight: 600;
    line-height: 1.3;
  }
  p {
    margin: 0 0 var(--space-1);
    color: var(--muted);
    font-size: var(--fs-body);
    line-height: 1.45;
  }
  .remember {
    display: flex;
    align-items: center;
    gap: 8px;
    min-height: var(--control-height);
    margin: 0 0 var(--space-1);
    font-size: var(--fs-body);
    cursor: pointer;
  }
  .buttons {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }
  button {
    height: var(--control-height);
    padding: 0 var(--space-1);
    border: 0;
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--text);
    font-size: var(--fs-body);
    line-height: 1;
    cursor: pointer;
  }
  button.primary {
    background: var(--primary);
    color: var(--primary-ink);
    font-weight: 500;
  }
  button.primary:hover {
    opacity: 0.9;
  }
  button.primary:active {
    opacity: 0.8;
  }
  button.quiet {
    color: var(--muted);
  }
  button.quiet:hover {
    background: var(--raised);
    color: var(--text);
  }
  button.quiet:active {
    background: var(--active);
  }
  button.danger {
    color: var(--negative);
  }
  button.danger:hover {
    background: var(--raised);
    color: var(--negative);
  }
</style>
