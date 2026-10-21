// The only module that talks to the Go bindings. Components stay
// presentational and read from here.
import { Events } from '@wailsio/runtime'
import { ProjectsService } from '../../bindings/dockercontroller'
import { Activity as ActivityBusEvent, Snapshot as SnapshotEvent } from '../../bindings/events'
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
  [ActionKind.ActionStop]: 'Stop',
  [ActionKind.ActionRestart]: 'Restart',
  [ActionKind.ActionPull]: 'Pull',
  [ActionKind.ActionBuild]: 'Build',
}

export interface OutputLine {
  text: string
  stderr: boolean
}

/** One Activity of this session, with its output and outcome. */
export interface ActivityRecord {
  id: string
  project: string
  kind: ActionKind
  state: 'running' | 'ok' | 'error'
  error: string
  lines: OutputLine[]
}

export interface Toast {
  message: string
  /** The Activity whose output the toast links to, when there is one. */
  activityId: string | null
}

const empty: Snapshot = { dockerReachable: true, projects: [] }

let snapshot = $state<Snapshot>(empty)
let loaded = $state(false)
let started = false
// Session history, newest first.
let activities = $state<ActivityRecord[]>([])
let selectedId = $state<string | null>(null)
let toast = $state<Toast | null>(null)

function applyActivity(e: ActivityEvent) {
  if (e.type === ActivityEventType.ActivityStarted) {
    activities.unshift({ id: e.id, project: e.project, kind: e.kind, state: 'running', error: '', lines: [] })
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
      message: `${actionLabels[e.kind] ?? e.kind} su ${e.project} non riuscito`,
      activityId: e.id,
    }
  }
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
  ProjectsService.Snapshot().then((s) => {
    // An event may already have delivered something newer.
    if (!loaded) apply(s)
  })
}

/** Launches an Action on a Project; a rejection is reported as a toast. */
export async function run(project: string, kind: ActionKind) {
  try {
    await ProjectsService.Do({ project, kind })
  } catch (err) {
    toast = { message: err instanceof Error ? err.message : String(err), activityId: null }
  }
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
  get toast(): Toast | null {
    return toast
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
