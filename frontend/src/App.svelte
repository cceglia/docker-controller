<script lang="ts">
  import { onMount } from 'svelte'
  import ProjectList from './lib/ProjectList.svelte'
  import ProjectDetail from './lib/ProjectDetail.svelte'
  import ActivityPanel from './lib/ActivityPanel.svelte'
  import ErrorToast from './lib/ErrorToast.svelte'
  import AddProjectDialog from './lib/AddProjectDialog.svelte'
  import SettingsView from './lib/SettingsView.svelte'
  import CloseDialog from './lib/CloseDialog.svelte'
  import UsageBar from './lib/UsageBar.svelte'
  import {
    addProject,
    cancelClose,
    closeLogs,
    openFolder,
    openLogs,
    openPort,
    openTerminal,
    pickComposeFile,
    proposeName,
    removeFromList,
    resolveClose,
    run,
    saveSettings,
    selectProject,
    setAdding,
    setShowingSettings,
    start,
    store,
    updateSettings,
    setProfiles,
    usageOfProject,
  } from './lib/store.svelte'

  onMount(start)
</script>

<main>
  <UsageBar usage={store.usage} />
  <header>
    <h1>Docker Controller</h1>
    <div class="header-actions">
      <button onclick={() => setShowingSettings(true)}>Impostazioni</button>
      <button onclick={() => setAdding(true)}>Aggiungi Progetto</button>
    </div>
  </header>
  {#if !store.dockerReachable}
    <p class="banner">Docker non raggiungibile</p>
  {/if}
  {#if store.loaded}
    {#if store.showingSettings && store.settings}
      <SettingsView settings={store.settings} onsave={saveSettings} onback={() => setShowingSettings(false)} />
    {:else if store.selectedProject}
      <ProjectDetail
        project={store.selectedProject}
        onaction={run}
        onopenport={openPort}
        onopenfolder={openFolder}
        onopenterminal={openTerminal}
        onback={() => selectProject(null)}
        onremove={removeFromList}
        onsettings={updateSettings}
        onprofiles={setProfiles}
        groups={store.groups}
        logs={store.logs}
        onopenlogs={openLogs}
        oncloselogs={closeLogs}
        unavailable={!store.dockerReachable}
        usage={usageOfProject(store.selectedProject.name)}
      />
    {:else}
      <ProjectList projects={store.projects} onaction={run} onselect={selectProject} unavailable={!store.dockerReachable} />
    {/if}
    <ActivityPanel activities={store.activities} selected={store.selectedActivity} />
  {:else}
    <p class="loading">Caricamento…</p>
  {/if}
</main>

{#if store.adding}
  <AddProjectDialog onpick={pickComposeFile} onpropose={proposeName} onadd={addProject} oncancel={() => setAdding(false)} />
{/if}

{#if store.closeAsked}
  <CloseDialog onchoose={resolveClose} oncancel={cancelClose} />
{/if}

{#if store.toast}
  <ErrorToast toast={store.toast} />
{/if}

<style>
  main { max-width: 56rem; margin: 0 auto; padding: 1.25rem; }
  header { display: flex; align-items: center; justify-content: space-between; margin: 0 0 1rem; }
  h1 { font-size: 1.25rem; margin: 0; }
  .header-actions { display: flex; gap: 0.5rem; }
  header button {
    font: inherit; font-size: 0.85rem; padding: 0.3rem 0.8rem; border-radius: 6px;
    border: 1px solid var(--border); background: var(--surface); color: var(--text); cursor: pointer;
  }
  .banner { padding: 0.6rem 0.9rem; border-radius: 8px; background: var(--warn-bg); color: var(--warn); }
  .loading { color: var(--muted); }
</style>
