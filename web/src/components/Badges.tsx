import { categoryLabel, priorityLabel, statusLabel } from '../format'
import type { Category, Priority, Status } from '../types'

const base = 'inline-flex items-center gap-1.5 whitespace-nowrap rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset'

const priorityStyle: Record<Priority, string> = {
  high: 'bg-red-50 text-red-700 ring-red-600/20 dark:bg-red-500/10 dark:text-red-300 dark:ring-red-400/30',
  medium: 'bg-amber-50 text-amber-800 ring-amber-600/20 dark:bg-amber-500/10 dark:text-amber-300 dark:ring-amber-400/30',
  low: 'bg-emerald-50 text-emerald-700 ring-emerald-600/20 dark:bg-emerald-500/10 dark:text-emerald-300 dark:ring-emerald-400/30',
}

const priorityDot: Record<Priority, string> = {
  high: 'bg-red-500',
  medium: 'bg-amber-500',
  low: 'bg-emerald-500',
}

export function PriorityBadge({ priority }: { priority: Priority }) {
  return (
    <span className={`${base} ${priorityStyle[priority]}`}>
      <span className={`size-1.5 rounded-full ${priorityDot[priority]}`} aria-hidden />
      {priorityLabel[priority]}
    </span>
  )
}

export function CategoryBadge({ category }: { category: Category }) {
  return (
    <span className={`${base} bg-slate-100 text-slate-700 ring-slate-500/20 dark:bg-slate-800 dark:text-slate-300 dark:ring-slate-600/50`}>
      {categoryLabel[category]}
    </span>
  )
}

const statusStyle: Record<Status, string> = {
  open: 'bg-sky-50 text-sky-700 ring-sky-600/20 dark:bg-sky-500/10 dark:text-sky-300 dark:ring-sky-400/30',
  in_progress: 'bg-violet-50 text-violet-700 ring-violet-600/20 dark:bg-violet-500/10 dark:text-violet-300 dark:ring-violet-400/30',
  resolved: 'bg-emerald-50 text-emerald-700 ring-emerald-600/20 dark:bg-emerald-500/10 dark:text-emerald-300 dark:ring-emerald-400/30',
  closed: 'bg-slate-100 text-slate-600 ring-slate-500/20 dark:bg-slate-800 dark:text-slate-400 dark:ring-slate-600/50',
}

export function StatusBadge({ status }: { status: Status }) {
  return <span className={`${base} ${statusStyle[status] ?? statusStyle.open}`}>{statusLabel[status] ?? status}</span>
}

export function NotTriagedBadge() {
  return (
    <span className={`${base} bg-transparent text-slate-500 ring-slate-300 ring-dashed dark:text-slate-400 dark:ring-slate-700`}>
      Not triaged
    </span>
  )
}
