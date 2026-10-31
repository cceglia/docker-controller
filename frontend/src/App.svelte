<script lang="ts">
  import { onMount } from 'svelte'
  import AppShell from './lib/AppShell.svelte'
  import ProjectList from './lib/ProjectList.svelte'
  import ProjectDetail from './lib/ProjectDetail.svelte'
  import ProjectSummary from './lib/ProjectSummary.svelte'
  import ActivityDrawer from './lib/ActivityDrawer.svelte'
  import ErrorToast from './lib/ErrorToast.svelte'
  import AddProjectDialog from './lib/AddProjectDialog.svelte'
  import SettingsView from './lib/SettingsView.svelte'
  import CloseDialog from './lib/CloseDialog.svelte'
  import {
    addProject,
    cancelClose,
    closeDrawer,
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
    setDrawerTab,
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

  // The column's Progetti entry is current when the canvas shows the summary.
  const projectsCurrent = $derived(!store.showingSettings && store.selectedProject === null)
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

<!-- The drawer belongs to the guscio: it sits under the content of the canvas.
     Lines of a stream of another Project are never shown. -->
{#snippet drawer()}
  {#if store.drawerOpen}
    {@const logs = store.selectedProject && store.logs?.project === store.selectedProject.name ? store.logs : null}
    <ActivityDrawer
      tab={store.drawerTab}
      activities={store.activities}
      selected={store.selectedActivity}
      {logs}
      services={store.selectedProject?.services?.map((s) => s.name) ?? []}
      ontab={setDrawerTab}
      onclose={closeDrawer}
      onservice={(service) => {
        const project = store.selectedProject
        if (project) openLogs(project.name, service)
      }}
      oncloselogs={closeLogs}
    />
  {/if}
{/snippet}

<AppShell
  usage={store.usage}
  onaddproject={() => setAdding(true)}
  onprojects={() => selectProject(null)}
  projectsCurrent={projectsCurrent}
  onsettings={() => setShowingSettings(true)}
  {navigation}
  {drawer}
>
  {#if store.loaded && !store.dockerReachable}
    <p class="banner">Docker non raggiungibile</p>
  {/if}
  {#if store.loaded}
    {#if store.showingSettings && store.settings}
      <SettingsView settings={store.settings} onsave={saveSettings} />
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
        onopenlogs={openLogs}
        unavailable={!store.dockerReachable}
        usage={usageOfProject(store.selectedProject.name)}
      />
    {:else}
      <ProjectSummary projects={store.projects} />
    {/if}
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
