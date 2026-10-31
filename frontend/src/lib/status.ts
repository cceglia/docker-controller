// The one place for the label and the tone of a state: the four Project
// states, the dot of a Container and the tone of its health. Components read
// these rules and never repeat them.

import { Status } from '../../bindings/dockercontroller/internal/compose/models'
import type { Container } from '../../bindings/dockercontroller/internal/compose/models'

/** The colour role of one status dot; StatusDot draws it. */
export type Tone = 'positive' | 'warning' | 'negative' | 'muted' | 'outline'

const PROJECT_TONES: Record<string, Tone> = {
  [Status.StatusActive]: 'positive',
  [Status.StatusPartial]: 'warning',
  [Status.StatusStopped]: 'muted',
  [Status.StatusRemoved]: 'outline',
}

/** The tone of a Project status; a status outside the four reads as stopped. */
export function projectTone(status: string): Tone {
  return PROJECT_TONES[status] ?? 'muted'
}

const PROJECT_LABELS: Record<string, string> = {
  [Status.StatusActive]: 'Attivo',
  [Status.StatusPartial]: 'Parziale',
  [Status.StatusStopped]: 'Fermo',
  [Status.StatusRemoved]: 'Rimosso',
}

/** The Italian label of a Project status; an unknown status is shown as it is. */
export function projectLabel(status: string): string {
  return PROJECT_LABELS[status] ?? status
}

/**
 * The tone of a Container: negative while Docker reports it unhealthy or dead,
 * warning while it starts, restarts or is paused, positive while it runs,
 * muted otherwise.
 */
export function containerTone(container: Container): Tone {
  if (container.health === 'unhealthy' || container.state === 'dead') return 'negative'
  if (container.health === 'starting' || container.state === 'restarting' || container.state === 'paused') {
    return 'warning'
  }
  if (container.state === 'running') return 'positive'
  return 'muted'
}

/** A Container is at fault when its tone is the error one: unhealthy or dead. */
export function containerFaulty(container: Container): boolean {
  return containerTone(container) === 'negative'
}

/**
 * The tone of the health pill: negative while unhealthy, warning while the
 * healthcheck is starting, null when the health is no problem and needs no pill.
 */
export function healthTone(health: string): 'negative' | 'warning' | null {
  if (health === 'unhealthy') return 'negative'
  if (health === 'starting') return 'warning'
  return null
}
