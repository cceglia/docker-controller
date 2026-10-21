// Number formatting for the Utilizzo risorse (Italian: decimal comma).

const BINARY = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB']
const DECIMAL = ['B', 'kB', 'MB', 'GB', 'TB', 'PB']

function number(value: number, digits: number): string {
  return value.toLocaleString('it-IT', { minimumFractionDigits: digits, maximumFractionDigits: digits })
}

/**
 * A size with its unit. Memory is shown with binary units (MiB, GiB), network
 * and disk with decimal ones (MB, GB), like `docker stats` does.
 */
export function formatBytes(bytes: number, base: 'binary' | 'decimal' = 'binary'): string {
  const step = base === 'binary' ? 1024 : 1000
  const units = base === 'binary' ? BINARY : DECIMAL
  let value = bytes
  let i = 0
  while (value >= step && i < units.length - 1) {
    value /= step
    i++
  }
  const digits = i === 0 || value >= 100 ? 0 : value >= 10 ? 1 : 2
  return `${number(value, digits)} ${units[i]}`
}

/** A percentage with one decimal. */
export function formatPercent(value: number): string {
  return `${number(value, 1)}%`
}

/** CPU in cores: docker reports 100 for one fully used core. */
export function formatCores(cpuPercent: number): string {
  return `${number(cpuPercent / 100, 2)} core`
}
