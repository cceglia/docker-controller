<script lang="ts">
  import { dismissToast, selectActivity } from './store.svelte'
  import type { Toast } from './store.svelte'

  let { toast }: { toast: Toast } = $props()

  function showOutput() {
    if (toast.activityId) selectActivity(toast.activityId)
    document.getElementById('attivita')?.scrollIntoView({ behavior: 'smooth' })
    dismissToast()
  }
</script>

<div class="toast" role="alert">
  <span>{toast.message}</span>
  {#if toast.activityId}
    <button class="link" onclick={showOutput}>Vedi output</button>
  {/if}
  <button class="close" aria-label="Chiudi" onclick={dismissToast}>×</button>
</div>

<style>
  .toast {
    position: fixed; right: 1rem; bottom: 1rem; max-width: 26rem; display: flex; align-items: center; gap: 0.75rem;
    padding: 0.7rem 0.9rem; border-radius: 8px; background: var(--surface); border: 1px solid var(--danger);
    color: var(--text); box-shadow: 0 4px 16px rgb(0 0 0 / 0.25);
  }
  button { font: inherit; background: none; border: none; color: inherit; cursor: pointer; }
  .link { color: var(--danger); text-decoration: underline; }
  .close { font-size: 1.1rem; color: var(--muted); }
</style>
