<script lang="ts">
  let {
    onpick,
    onpropose,
    onadd,
    oncancel,
  }: {
    /** Shows the file picker; resolves to the chosen path, '' when cancelled. */
    onpick: () => Promise<string>
    /** The compose name proposed for a main File compose. */
    onpropose: (mainFile: string) => Promise<string>
    /** Saves the Project; resolves to an error message, null on success. */
    onadd: (name: string, files: string[]) => Promise<string | null>
    oncancel: () => void
  } = $props()

  let files = $state<string[]>([])
  let name = $state('')
  // Once the user edits the name, a different main file no longer replaces it.
  let nameEdited = $state(false)
  let error = $state<string | null>(null)
  let busy = $state(false)

  const canAdd = $derived(files.length > 0 && name.trim() !== '' && !busy)

  /** Opens the dialog as a modal as soon as it is mounted. */
  const showModal = (dialog: HTMLDialogElement) => {
    dialog.showModal()
  }

  async function proposeFor(mainFile: string) {
    if (nameEdited) return
    try {
      name = await onpropose(mainFile)
    } catch (err) {
      error = err instanceof Error ? err.message : String(err)
    }
  }

  async function pick() {
    error = null
    try {
      const path = await onpick()
      if (!path) return
      if (files.includes(path)) return
      files.push(path)
      if (files.length === 1) await proposeFor(path)
    } catch (err) {
      error = err instanceof Error ? err.message : String(err)
    }
  }

  async function remove(index: number) {
    files.splice(index, 1)
    if (index === 0) {
      if (files.length > 0) await proposeFor(files[0])
      else if (!nameEdited) name = ''
    }
  }

  async function add() {
    busy = true
    error = null
    try {
      error = await onadd(name.trim(), $state.snapshot(files))
    } finally {
      busy = false
    }
  }
</script>

<dialog {@attach showModal} {oncancel} aria-labelledby="add-title">
  <h3 id="add-title">Aggiungi Progetto</h3>

  <div class="files">
    <span class="label">File compose</span>
    {#if files.length === 0}
      <p class="hint">Scegli il File compose principale del Progetto.</p>
    {:else}
      <ol>
        {#each files as file, i (file)}
          <li>
            <span class="path mono" title={file}>{file}</span>
            <span class="role">{i === 0 ? 'principale' : 'aggiuntivo'}</span>
            <button type="button" class="remove" aria-label="Rimuovi {file}" onclick={() => remove(i)}>×</button>
          </li>
        {/each}
      </ol>
    {/if}
    <button type="button" class="secondary" onclick={pick}>
      {files.length === 0 ? 'Scegli il File compose…' : 'Aggiungi un altro File compose…'}
    </button>
  </div>

  <label>
    Nome del Progetto
    <input
      type="text"
      bind:value={name}
      oninput={() => (nameEdited = true)}
      autocomplete="off"
      spellcheck="false"
      disabled={files.length === 0}
    />
  </label>

  {#if error}<p class="error" role="alert">{error}</p>{/if}

  <div class="buttons">
    <button type="button" class="secondary" onclick={oncancel}>Annulla</button>
    <button type="button" class="primary" disabled={!canAdd} onclick={add}>Aggiungi</button>
  </div>
</dialog>

<style>
  dialog {
    width: min(36rem, 90vw);
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

  .files {
    display: grid;
    justify-items: start;
    gap: 4px;
    margin: 0 0 var(--space-1);
  }
  .label {
    color: var(--muted);
    font-size: var(--fs-small);
  }
  .hint {
    color: var(--muted);
    font-size: var(--fs-body);
  }
  ol {
    display: grid;
    width: 100%;
    margin: 0;
    padding: 0;
    gap: 2px;
    list-style: none;
  }
  li {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: var(--fs-body);
  }
  .path {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .role {
    flex: none;
    color: var(--muted);
    font-size: var(--fs-small);
  }

  label {
    display: grid;
    gap: 6px;
    margin: 0 0 var(--space-1);
    color: var(--muted);
    font-size: var(--fs-small);
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

  .error {
    margin: 0 0 var(--space-1);
    color: var(--negative);
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
    border: 1px solid var(--control);
    border-radius: var(--radius-control);
    background: var(--surface);
    color: var(--text);
    font-size: var(--fs-body);
    cursor: pointer;
  }
  button:hover:not(:disabled) {
    background: var(--raised);
  }
  .secondary {
    justify-self: start;
  }
  .remove {
    flex: none;
    width: var(--control-height);
    padding: 0;
    border: 0;
    background: transparent;
    color: var(--muted);
    font-size: var(--fs-section);
    line-height: 1;
  }
  .remove:hover {
    background: var(--raised);
    color: var(--text);
  }
  .primary {
    border-color: var(--primary);
    background: var(--primary);
    color: var(--primary-ink);
    font-weight: 500;
  }
  .primary:hover:not(:disabled) {
    background: var(--primary);
    opacity: 0.9;
  }
  button:disabled,
  input:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
</style>
