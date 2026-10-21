// The only module that talks to the Go bindings. Components stay
// presentational and read from here.
import { Events } from '@wailsio/runtime'
import { ProjectsService } from '../../bindings/dockercontroller'
import { Activity as ActivityBusEvent, Log as LogBusEvent, Snapshot as SnapshotEvent } from '../../bindings/events'
import { LogEventType } from '../../bindings/dockercontroller/models'
import type { LogEvent } from '../../bindings/dockercontroller/models'
import {
  ActionKind,
  ActivityEventType,
} from '../../bindings/dockercontroller/internal/compose/models'
import type {
  ActivityEvent,
  Project,
  Snapshot,
} from '../../bindings/dockercontroller/internal/compose/models'

export { ActionKind }

/** Italian label of an Action, as shown on buttons and indicators. */
export const actionLabels: Record<string, string> = {
  [ActionKind.ActionUp]: 'Up',
  [ActionKind.ActionStart]: 'Start',
  [ActionKind.ActionStop]: 'Stop',
  [ActionKind.ActionRestart]: 'Restart',
  [ActionKind.ActionPull]: 'Pull',
  [ActionKind.ActionBuild]: 'Build',
  [ActionKind.ActionDown]: 'Down',
  [ActionKind.ActionDownVolumes]: 'Down con volumi',
}

export interface OutputLine {
  text: string
  stderr: boolean
}

/** One Activity of this session, with its output and outcome. */
export interface ActivityRecord {
  id: string
  project: string
  /** The Service the Action targets, '' for the whole Project. */
  service: string
  kind: ActionKind
  state: 'running' | 'ok' | 'error'
  error: string
  lines: OutputLine[]
}

/** The most lines the log panel keeps; the oldest are dropped. */
export const LOG_LIMIT = 5000

/** One line of the log panel. `seq` is unique and increasing within the open stream. */
export interface LogEntry {
  seq: number
  /** The Service that wrote the line; '' for a compose message. */
  service: string
  /** RFC 3339 timestamp as Docker recorded it; '' when there is none. */
  timestamp: string
  text: string
  stderr: boolean
}

/** The log stream open for a Project (null when the panel is closed). */
export interface LogView {
  project: string
  /** The Service the stream is restricted to, '' for the whole Project. */
  service: string
  lines: LogEntry[]
  /** False once compose ended the stream by itself. */
  live: boolean
  /** Why the stream ended, '' when it did not fail. */
  error: string
}

export interface Toast {
  message: string
  /** The Activity whose output the toast links to, when there is one. */
  activityId: string | null
}

const empty: Snapshot = { dockerReachable: true, projects: [], groups: [] }

let snapshot = $state<Snapshot>(empty)
let loaded = $state(false)
let started = false
// Session history, newest first.
let activities = $state<ActivityRecord[]>([])
let selectedId = $state<string | null>(null)
let toast = $state<Toast | null>(null)
let selectedProject = $state<string | null>(null)
let adding = $state(false)

function applyActivity(e: ActivityEvent) {
  if (e.type === ActivityEventType.ActivityStarted) {
    activities.unshift({ id: e.id, project: e.project, service: e.service, kind: e.kind, state: 'running', error: '', lines: [] })
    selectedId = e.id
    return
  }
  const record = activities.find((a) => a.id === e.id)
  if (!record) return
  if (e.type === ActivityEventType.ActivityLine) {
    record.lines.push({ text: e.line, stderr: e.stderr })
    return
  }
  record.state = e.error ? 'error' : 'ok'
  record.error = e.error
  if (e.error) {
    toast = {
      message: `${actionLabels[e.kind] ?? e.kind} su ${e.service ? `${e.project}/${e.service}` : e.project} non riuscito`,
      activityId: e.id,
    }
  }
}

// Log stream. The id is chosen here and given to the binding, so events are
// recognised even before StartLogs returns; lines of a stream that was closed
// are ignored.
let logView = $state<LogView | null>(null)
let logId: string | null = null
let logSeq = 0
let logPending: LogEntry[] = []
let logTimer: ReturnType<typeof setTimeout> | null = null

/** Appends `incoming` to `lines`, keeping only the last `limit`. */
export function appendCapped<T>(lines: T[], incoming: T[], limit: number): T[] {
  const all = lines.length === 0 ? incoming : lines.concat(incoming)
  return all.length > limit ? all.slice(all.length - limit) : all
}

// Lines arrive in bursts (the first 500 at once): they are applied together
// so the panel re-renders once per burst, not once per line.
function flushLogs() {
  logTimer = null
  if (logView && logPending.length > 0) logView.lines = appendCapped(logView.lines, logPending, LOG_LIMIT)
  logPending = []
}

function applyLog(e: LogEvent) {
  if (!logView || e.id !== logId) return
  if (e.type === LogEventType.LogLineEvent) {
    logPending.push({ seq: ++logSeq, service: e.line.service, timestamp: e.line.timestamp, text: e.line.text, stderr: e.line.stderr })
    if (logTimer === null) logTimer = setTimeout(flushLogs, 50)
    return
  }
  flushLogs()
  logView.live = false
  logView.error = e.error
}

function apply(next: Snapshot) {
  snapshot = { ...next, projects: next.projects ?? [] }
  loaded = true
}

/** Subscribes to snapshot events and loads the initial one. Idempotent. */
export function start() {
  if (started) return
  started = true
  Events.On(SnapshotEvent, (event) => apply(event.data))
  Events.On(ActivityBusEvent, (event) => applyActivity(event.data))
  Events.On(LogBusEvent, (event) => applyLog(event.data))
  ProjectsService.Snapshot().then((s) => {
    // An event may already have delivered something newer.
    if (!loaded) apply(s)
  })
}

function fail(err: unknown) {
  toast = { message: err instanceof Error ? err.message : String(err), activityId: null }
}

/** Launches an Action on a Project, or on one of its Services; a rejection is reported as a toast. */
export async function run(project: string, kind: ActionKind, service = '') {
  try {
    await ProjectsService.Do({ project, service, kind })
  } catch (err) {
    fail(err)
  }
}

/** Removes a Rimosso Project from the saved list (Docker is not touched). */
export async function removeFromList(project: string) {
  try {
    await ProjectsService.RemoveFromList(project)
  } catch (err) {
    fail(err)
  }
}

/** Sets (or, with an empty string, clears) the Display name and/or Group of a Project; null leaves a field as it is. */
export async function updateSettings(project: string, displayName: string | null, group: string | null) {
  try {
    await ProjectsService.UpdateSettings(project, { displayName, group, profiles: null })
  } catch (err) {
    fail(err)
  }
}

/** Saves the active Profiles of a Project (the complete set). It only saves: no Action is launched. */
export async function setProfiles(project: string, profiles: string[]) {
  try {
    await ProjectsService.UpdateSettings(project, { displayName: null, group: null, profiles })
  } catch (err) {
    fail(err)
  }
}

/** Shows the native file picker; resolves to the chosen File compose, '' when cancelled. */
export function pickComposeFile(): Promise<string> {
  return ProjectsService.PickComposeFile()
}

/** The compose name proposed for a Project whose main File compose is `mainFile`. */
export function proposeName(mainFile: string): Promise<string> {
  return ProjectsService.ProposeName(mainFile)
}

/** Adds a Project by hand; resolves to the error message when it is refused, null when saved. */
export async function addProject(name: string, files: string[]): Promise<string | null> {
  try {
    await ProjectsService.AddProject(name, files)
    adding = false
    return null
  } catch (err) {
    return err instanceof Error ? err.message : String(err)
  }
}

/** Opens or closes the "Aggiungi Progetto" dialog. */
export function setAdding(value: boolean) {
  adding = value
}

/** Opens http://localhost:<port> in the browser. */
export async function openPort(port: number) {
  try {
    await ProjectsService.OpenPort(port)
  } catch (err) {
    fail(err)
  }
}

/** Opens the folder of a Project in the file manager. */
export async function openFolder(project: string) {
  try {
    await ProjectsService.OpenFolder(project)
  } catch (err) {
    fail(err)
  }
}

/** Opens the detail of a Project (null goes back to the list). The log stream of the previous one is closed. */
export function selectProject(name: string | null) {
  closeLogs()
  selectedProject = name
}

/**
 * Opens the log stream of a Project, or of one of its Services ('' for all);
 * a stream already open is closed first. The last 500 lines arrive, then the
 * live ones; the panel keeps at most LOG_LIMIT.
 */
export async function openLogs(project: string, service = '') {
  closeLogs()
  const id = crypto.randomUUID()
  logId = id
  logSeq = 0
  logView = { project, service, lines: [], live: true, error: '' }
  try {
    await ProjectsService.StartLogs(id, project, service)
  } catch (err) {
    if (logId !== id) return // closed in the meantime
    logView = null
    logId = null
    fail(err)
  }
}

/** Closes the log stream: the `compose logs` process is terminated. */
export function closeLogs() {
  if (logTimer !== null) clearTimeout(logTimer)
  logTimer = null
  logPending = []
  if (logId !== null) {
    const id = logId
    ProjectsService.StopLogs(id).catch(fail)
  }
  logId = null
  logView = null
}

const SERVICE_HUES = 16

/** A CSS colour for a Service: the same name always gets the same colour. */
export function serviceColor(service: string): string {
  if (service === '') return 'var(--muted)'
  let h = 0
  for (const ch of service) h = (h * 31 + ch.codePointAt(0)!) >>> 0
  return `hsl(${(h % SERVICE_HUES) * (360 / SERVICE_HUES)} 65% var(--log-lightness, 50%))`
}

/** Shows the output of an Activity in the panel. */
export function selectActivity(id: string | null) {
  selectedId = id
}

export function dismissToast() {
  toast = null
}

export const store = {
  get activities(): ActivityRecord[] {
    return activities
  },
  get selectedActivity(): ActivityRecord | null {
    return activities.find((a) => a.id === selectedId) ?? null
  },
  /** The "Aggiungi Progetto" dialog is open. */
  get adding(): boolean {
    return adding
  },
  /** The open log stream; null when the panel is closed. */
  get logs(): LogView | null {
    return logView
  },
  get toast(): Toast | null {
    return toast
  },
  /** The Project whose detail is open; null when it was removed from the list. */
  get selectedProject(): Project | null {
    return (snapshot.projects ?? []).find((p) => p.name === selectedProject) ?? null
  },
  /** The Groups in use, each once, sorted. */
  get groups(): string[] {
    return snapshot.groups ?? []
  },
  get projects(): Project[] {
    return snapshot.projects ?? []
  },
  get dockerReachable(): boolean {
    return snapshot.dockerReachable
  },
  get loaded(): boolean {
    return loaded
  },
}
