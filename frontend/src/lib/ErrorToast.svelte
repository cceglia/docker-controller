<script lang="ts">
  import { dismissToast, selectActivity } from './store.svelte'
  import type { Toast } from './store.svelte'

  let { toast }: { toast: Toast } = $props()

  /** How long the toast stays before going away on its own. */
  const VISIBLE_MS = 6000

  let timer: ReturnType<typeof setTimeout> | null = null
  let hovering = false
  let focused = false

  function stopTimer() {
    if (timer !== null) clearTimeout(timer)
    timer = null
  }

  function startTimer() {
    stopTimer()
    if (hovering || focused) return
    timer = setTimeout(dismissToast, VISIBLE_MS)
  }

  // A new message replaces the previous one and restarts the countdown. The
  // timer pauses while the toast is hovered or focused, so it cannot vanish
  // while it is being read or clicked.
  $effect(() => {
    // Depend on the whole toast: a new one, even with the same text, restarts
    // the countdown.
    void toast.message
    void toast.activityId
    startTimer()
    return stopTimer
  })

  function showOutput() {
    if (toast.activityId) selectActivity(toast.activityId)
    document.getElementById('attivita')?.scrollIntoView({ behavior: 'smooth' })
    dismissToast()
  }
</script>

<div
  class="toast"
  role="alert"
  onmouseenter={() => {
    hovering = true
    stopTimer()
  }}
  onmouseleave={() => {
    hovering = false
    startTimer()
  }}
  onfocusin={() => {
    focused = true
    stopTimer()
  }}
  onfocusout={() => {
    focused = false
    startTimer()
  }}
>
  <span class="message">{toast.message}</span>
  {#if toast.activityId}
    <button class="link" type="button" onclick={showOutput}>Vedi output</button>
  {/if}
  <button class="close" type="button" aria-label="Chiudi" onclick={dismissToast}>Chiudi</button>
</div>

<style>
  .toast {
    position: fixed; right: 1rem; bottom: 1rem; max-width: 26rem; display: flex; align-items: center; gap: 0.75rem;
    padding: 0.7rem 0.9rem; border-radius: 8px; background: var(--surface); border: 1px solid var(--danger);
    color: var(--text); box-shadow: 0 4px 16px rgb(0 0 0 / 0.25);
  }
  .message { flex: 1; min-width: 0; overflow-wrap: anywhere; }
  button { font: inherit; border: none; cursor: pointer; }
  .link { flex: none; padding: 0; background: none; color: var(--danger); text-decoration: underline; }
  .close {
    flex: none; padding: 0.25rem 0.6rem; border: 1px solid var(--border); border-radius: 6px;
    background: var(--bg); color: var(--text); font-size: 0.8rem;
  }
</style>
