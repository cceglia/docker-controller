<script lang="ts">
  import { onMount } from 'svelte'
  import AppShell from './lib/AppShell.svelte'
  import ProjectList from './lib/ProjectList.svelte'
  import ProjectDetail from './lib/ProjectDetail.svelte'
  import ProjectSummary from './lib/ProjectSummary.svelte'
  import ActivityPanel from './lib/ActivityPanel.svelte'
  import ErrorToast from './lib/ErrorToast.svelte'
  import AddProjectDialog from './lib/AddProjectDialog.svelte'
  import SettingsView from './lib/SettingsView.svelte'
  import CloseDialog from './lib/CloseDialog.svelte'
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

  // Whatever hides the log panel (Impostazioni, the Project gone from the
  // list) must also stop the stream.
  $effect(() => {
    if (store.logs !== null && (store.showingSettings || store.selectedProject === null)) closeLogs()
  })
</script>

{#snippet navigation()}
  {#if store.loaded}
    <ProjectList
      projects={store.projects}
      selected={store.selectedProject?.name ?? null}
      onselect={selectProject}
    />
  {/if}
{/snippet}

<AppShell
  usage={store.usage}
  onaddproject={() => setAdding(true)}
  onsettings={() => setShowingSettings(true)}
  {navigation}
>
  {#if store.loaded && !store.dockerReachable}
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
      <ProjectSummary projects={store.projects} />
    {/if}
    <ActivityPanel activities={store.activities} selected={store.selectedActivity} />
  {:else}
    <p class="loading">Caricamento…</p>
  {/if}
</AppShell>

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
  .banner {
    margin: 0 0 var(--space-2);
    padding: var(--space-1);
    border-radius: var(--radius-card);
    background: var(--pill-warning-bg);
    color: var(--pill-warning-ink);
    font-size: var(--fs-body);
  }
  .loading {
    color: var(--muted);
  }
</style>
