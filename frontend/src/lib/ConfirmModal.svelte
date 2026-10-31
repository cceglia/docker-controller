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
      <span>Per confermare digita <code>{confirmText}</code></span>
      <input type="text" bind:value={typed} autocomplete="off" spellcheck="false" />
    </label>
  {/if}
  <div class="buttons">
    <button class="quiet" onclick={oncancel}>Annulla</button>
    <button class="primary" class:danger disabled={!confirmable} onclick={onconfirm}>{confirmLabel}</button>
  </div>
</dialog>

<style>
  dialog {
    width: min(28rem, 90vw);
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
  label {
    display: grid;
    gap: 6px;
    margin: 0 0 var(--space-1);
    color: var(--muted);
    font-size: var(--fs-small);
  }
  code {
    padding: 1px 4px;
    border-radius: var(--radius-control);
    background: var(--raised);
    color: var(--text);
    font-family: var(--font-mono);
    font-size: var(--fs-body);
  }
  input {
    height: var(--control-height);
    padding: 0 var(--space-1);
    border: 1px solid var(--control);
    border-radius: var(--radius-control);
    background: var(--surface);
    color: var(--text);
    font-size: var(--fs-body);
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
  .primary {
    background: var(--primary);
    color: var(--primary-ink);
    font-weight: 500;
  }
  .primary:not(:disabled):hover {
    opacity: 0.9;
  }
  .primary:not(:disabled):active {
    opacity: 0.8;
  }
  .primary.danger {
    background: var(--negative);
  }
  .quiet {
    color: var(--muted);
  }
  .quiet:hover {
    background: var(--raised);
    color: var(--text);
  }
  .quiet:active {
    background: var(--active);
  }
  button:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
</style>
