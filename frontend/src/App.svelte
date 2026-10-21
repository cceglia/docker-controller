<script lang="ts">
  import { onMount } from 'svelte'
  import ProjectList from './lib/ProjectList.svelte'
  import ActivityPanel from './lib/ActivityPanel.svelte'
  import ErrorToast from './lib/ErrorToast.svelte'
  import { run, start, store } from './lib/store.svelte'

  onMount(start)
</script>

<main>
  <h1>Docker Controller</h1>
  {#if !store.dockerReachable}
    <p class="banner">Docker non raggiungibile</p>
  {/if}
  {#if store.loaded}
    <ProjectList projects={store.projects} onaction={run} />
    <ActivityPanel activities={store.activities} selected={store.selectedActivity} />
  {:else}
    <p class="loading">Caricamento…</p>
  {/if}
</main>

{#if store.toast}
  <ErrorToast toast={store.toast} />
{/if}

<style>
  main { max-width: 56rem; margin: 0 auto; padding: 1.25rem; }
  h1 { font-size: 1.25rem; margin: 0 0 1rem; }
  .banner { padding: 0.6rem 0.9rem; border-radius: 8px; background: var(--warn-bg); color: var(--warn); }
  .loading { color: var(--muted); }
</style>
