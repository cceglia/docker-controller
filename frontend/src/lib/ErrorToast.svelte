<script lang="ts">
  import { tick } from 'svelte'
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

  // The drawer opens on the Activity of the message; the element is then in
  // the page and can be brought into view. The scroll is smooth only when the
  // user did not ask for reduced motion.
  async function showOutput() {
    if (toast.activityId) selectActivity(toast.activityId)
    dismissToast()
    await tick()
    const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches
    document.getElementById('attivita')?.scrollIntoView({ behavior: reduce ? 'auto' : 'smooth' })
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
  <span class="dot" aria-hidden="true"></span>
  <span class="message">{toast.message}</span>
  <span class="actions">
    {#if toast.activityId}
      <button class="link" type="button" onclick={showOutput}>Vedi output</button>
    {/if}
    <button class="close" type="button" onclick={dismissToast}>Chiudi</button>
  </span>
</div>

<style>
  .toast {
    position: fixed;
    right: var(--space-2);
    bottom: var(--space-2);
    display: flex;
    align-items: flex-start;
    gap: var(--space-1);
    max-width: 26rem;
    padding: var(--space-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-card);
    background: var(--surface);
    color: var(--text);
  }
  .dot {
    flex: none;
    width: var(--dot-size);
    height: var(--dot-size);
    margin-top: 7px;
    border-radius: var(--dot-size);
    background: var(--negative);
  }
  .message {
    flex: 1;
    min-width: 0;
    overflow-wrap: anywhere;
    font-size: var(--fs-body);
    line-height: 1.4;
  }
  .actions {
    display: flex;
    flex: none;
    align-items: center;
    gap: 4px;
  }
  button {
    border: 0;
    background: none;
    font-size: var(--fs-small);
    cursor: pointer;
  }
  .link {
    min-height: var(--control-height);
    padding: 0 6px;
    border-radius: var(--radius-control);
    color: var(--text);
    text-decoration: underline;
  }
  .link:hover {
    background: var(--raised);
  }
  .close {
    height: var(--control-height);
    padding: 0 var(--space-1);
    border: 1px solid var(--control);
    border-radius: var(--radius-control);
    background: var(--surface);
    color: var(--text);
  }
  .close:hover {
    background: var(--raised);
  }
</style>
