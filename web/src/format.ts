import type { Category, Priority, Status } from './types'

export const statusLabel: Record<Status, string> = {
  open: 'Open',
  in_progress: 'In progress',
  resolved: 'Resolved',
  closed: 'Closed',
}

export const priorityLabel: Record<Priority, string> = {
  high: 'High',
  medium: 'Medium',
  low: 'Low',
}

export const categoryLabel: Record<Category, string> = {
  account: 'Account',
  email: 'Email',
  hardware: 'Hardware',
  software: 'Software',
  network: 'Network',
  security: 'Security',
  other: 'Other',
}

/** Higher number = more urgent, for sorting. Untriaged tickets sort last. */
export const priorityRank: Record<Priority, number> = { high: 3, medium: 2, low: 1 }

const relative = new Intl.RelativeTimeFormat(undefined, { numeric: 'auto' })
const absolute = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' })

const units: [Intl.RelativeTimeFormatUnit, number][] = [
  ['year', 365 * 24 * 3600],
  ['month', 30 * 24 * 3600],
  ['week', 7 * 24 * 3600],
  ['day', 24 * 3600],
  ['hour', 3600],
  ['minute', 60],
]

/** "5 minutes ago", "yesterday", "just now". */
export function timeAgo(iso: string): string {
  const seconds = (new Date(iso).getTime() - Date.now()) / 1000
  for (const [unit, size] of units) {
    if (Math.abs(seconds) >= size) return relative.format(Math.round(seconds / size), unit)
  }
  return 'just now'
}

export function fullDate(iso: string): string {
  return absolute.format(new Date(iso))
}

/** Models often number their steps ("1. Restart"); we render an <ol>, so drop that prefix. */
export function stripStepNumber(step: string): string {
  return step.replace(/^\s*(?:step\s*)?\d+\s*[.):-]\s*/i, '')
}
