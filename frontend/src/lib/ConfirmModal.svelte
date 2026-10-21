<script lang="ts">
  let {
    title,
    message,
    confirmLabel,
    confirmText = '',
    danger = false,
    onconfirm,
    oncancel,
  }: {
    title: string
    message: string
    confirmLabel: string
    /** When not empty, the confirm button enables only once exactly this text is typed. */
    confirmText?: string
    danger?: boolean
    onconfirm: () => void
    oncancel: () => void
  } = $props()

  let typed = $state('')

  const confirmable = $derived(confirmText === '' || typed === confirmText)

  /** Opens the dialog as a modal as soon as it is mounted. */
  const showModal = (dialog: HTMLDialogElement) => {
    dialog.showModal()
  }
</script>

<dialog {@attach showModal} {oncancel} aria-labelledby="modal-title">
  <h3 id="modal-title">{title}</h3>
  <p>{message}</p>
  {#if confirmText}
    <label>
      Per confermare digita <code>{confirmText}</code>
      <input type="text" bind:value={typed} autocomplete="off" spellcheck="false" />
    </label>
  {/if}
  <div class="buttons">
    <button onclick={oncancel}>Annulla</button>
    <button class:danger disabled={!confirmable} onclick={onconfirm}>{confirmLabel}</button>
  </div>
</dialog>

<style>
  dialog {
    max-width: 28rem; border: 1px solid var(--border); border-radius: 10px;
    background: var(--surface); color: var(--text); padding: 1.1rem 1.25rem;
  }
  dialog::backdrop { background: rgb(0 0 0 / 0.5); }
  h3 { margin: 0 0 0.5rem; font-size: 1rem; }
  p { margin: 0 0 0.75rem; font-size: 0.9rem; }
  label { display: grid; gap: 0.3rem; font-size: 0.85rem; margin-bottom: 0.75rem; }
  input { font: inherit; padding: 0.3rem 0.5rem; border-radius: 6px; border: 1px solid var(--border); background: var(--bg); color: var(--text); }
  .buttons { display: flex; justify-content: flex-end; gap: 0.5rem; }
  button {
    font: inherit; font-size: 0.85rem; padding: 0.3rem 0.8rem; border-radius: 6px;
    border: 1px solid var(--border); background: var(--surface); color: var(--text); cursor: pointer;
  }
  button.danger { background: var(--danger); border-color: var(--danger); color: #fff; }
  button:disabled { opacity: 0.45; cursor: not-allowed; }
</style>
