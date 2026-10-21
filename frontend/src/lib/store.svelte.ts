// The only module that talks to the Go bindings. Components stay
// presentational and read from here.
import { Events } from '@wailsio/runtime'
import { ProjectsService } from '../../bindings/dockercontroller'
import { Snapshot as SnapshotEvent } from '../../bindings/events'
import type { Project, Snapshot } from '../../bindings/dockercontroller/internal/compose/models'

const empty: Snapshot = { dockerReachable: true, projects: [] }

let snapshot = $state<Snapshot>(empty)
let loaded = $state(false)
let started = false

function apply(next: Snapshot) {
  snapshot = { ...next, projects: next.projects ?? [] }
  loaded = true
}

/** Subscribes to snapshot events and loads the initial one. Idempotent. */
export function start() {
  if (started) return
  started = true
  Events.On(SnapshotEvent, (event) => apply(event.data))
  ProjectsService.Snapshot().then((s) => {
    // An event may already have delivered something newer.
    if (!loaded) apply(s)
  })
}

export const store = {
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
