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
            <span class="path" title={file}>{file}</span>
            <span class="role">{i === 0 ? 'principale' : 'aggiuntivo'}</span>
            <button type="button" aria-label="Rimuovi {file}" onclick={() => remove(i)}>×</button>
          </li>
        {/each}
      </ol>
    {/if}
    <button type="button" onclick={pick}>
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
    <button type="button" onclick={oncancel}>Annulla</button>
    <button type="button" class="primary" disabled={!canAdd} onclick={add}>Aggiungi</button>
  </div>
</dialog>

<style>
  dialog {
    width: min(36rem, 90vw); border: 1px solid var(--border); border-radius: 10px;
    background: var(--surface); color: var(--text); padding: 1.1rem 1.25rem;
  }
  dialog::backdrop { background: rgb(0 0 0 / 0.5); }
  h3 { margin: 0 0 0.75rem; font-size: 1rem; }
  .files { display: grid; gap: 0.4rem; justify-items: start; margin-bottom: 0.75rem; }
  .label { font-size: 0.85rem; }
  .hint { margin: 0; color: var(--muted); font-size: 0.85rem; }
  ol { list-style: none; margin: 0; padding: 0; display: grid; gap: 0.25rem; width: 100%; }
  li { display: flex; align-items: center; gap: 0.5rem; font-size: 0.85rem; }
  .path { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .role { color: var(--muted); font-size: 0.75rem; }
  label { display: grid; gap: 0.3rem; font-size: 0.85rem; margin-bottom: 0.75rem; }
  input { font: inherit; padding: 0.3rem 0.5rem; border-radius: 6px; border: 1px solid var(--border); background: var(--bg); color: var(--text); }
  .error { color: var(--danger); font-size: 0.85rem; margin: 0 0 0.75rem; }
  .buttons { display: flex; justify-content: flex-end; gap: 0.5rem; }
  button {
    font: inherit; font-size: 0.85rem; padding: 0.3rem 0.8rem; border-radius: 6px;
    border: 1px solid var(--border); background: var(--surface); color: var(--text); cursor: pointer;
  }
  button.primary { border-color: var(--ok); color: var(--ok); }
  button:disabled, input:disabled { opacity: 0.45; cursor: not-allowed; }
</style>
