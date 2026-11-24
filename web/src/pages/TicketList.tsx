import { useMemo, useState } from 'react'
import { api, errorMessage } from '../api'
import { CategoryBadge, NotTriagedBadge, PriorityBadge, StatusBadge } from '../components/Badges'
import { ButtonLink, EmptyState, ErrorBanner, LoadingBlock } from '../components/ui'
import { categoryLabel, fullDate, priorityLabel, priorityRank, statusLabel, timeAgo } from '../format'
import { Link } from '../components/Link'
import { CATEGORIES, PRIORITIES, STATUSES, type TicketListItem } from '../types'
import { useLoad } from '../useLoad'

type Sort = 'newest' | 'priority'

export function TicketList() {
  const { data: tickets, error, loading, reload } = useLoad((signal) => api.listTickets(signal))

  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('active')
  const [priority, setPriority] = useState('')
  const [category, setCategory] = useState('')
  const [sort, setSort] = useState<Sort>('newest')

  const visible = useMemo(() => {
    if (!tickets) return []
    const q = search.trim().toLowerCase()
    const list = tickets.filter((t) => {
      if (status === 'active' && (t.status === 'resolved' || t.status === 'closed')) return false
      if (status && status !== 'active' && t.status !== status) return false
      if (priority === 'untriaged' && t.triage) return false
      if (priority && priority !== 'untriaged' && t.triage?.priority !== priority) return false
      if (category && t.triage?.category !== category) return false
      if (q && !`${t.title} ${t.description} ${t.requesterEmail} #${t.id}`.toLowerCase().includes(q)) return false
      return true
    })
    if (sort === 'priority') {
      // Stable sort: equal priorities keep the API's newest-first order.
      list.sort((a, b) => rank(b) - rank(a))
    }
    return list
  }, [tickets, search, status, priority, category, sort])

  const filtered = status !== '' || priority !== '' || category !== '' || search !== ''

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Tickets</h1>
          {tickets && (
            <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
              {visible.length} of {tickets.length} {tickets.length === 1 ? 'ticket' : 'tickets'}
              {countUntriaged(tickets) > 0 && ` · ${countUntriaged(tickets)} waiting for triage`}
            </p>
          )}
        </div>
      </div>

      <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-[1fr_repeat(4,minmax(0,10rem))]">
        <input
          type="search"
          className="control sm:col-span-2 lg:col-span-1"
          placeholder="Search title, description, email…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          aria-label="Search tickets"
        />
        <select className="control" value={status} onChange={(e) => setStatus(e.target.value)} aria-label="Filter by status">
          <option value="active">Active</option>
          <option value="">All statuses</option>
          {STATUSES.map((s) => (
            <option key={s} value={s}>
              {statusLabel[s]}
            </option>
          ))}
        </select>
        <select className="control" value={priority} onChange={(e) => setPriority(e.target.value)} aria-label="Filter by priority">
          <option value="">All priorities</option>
          {PRIORITIES.map((p) => (
            <option key={p} value={p}>
              {priorityLabel[p]}
            </option>
          ))}
          <option value="untriaged">Not triaged</option>
        </select>
        <select className="control" value={category} onChange={(e) => setCategory(e.target.value)} aria-label="Filter by category">
          <option value="">All categories</option>
          {CATEGORIES.map((c) => (
            <option key={c} value={c}>
              {categoryLabel[c]}
            </option>
          ))}
        </select>
        <select className="control" value={sort} onChange={(e) => setSort(e.target.value as Sort)} aria-label="Sort">
          <option value="newest">Newest first</option>
          <option value="priority">Highest priority</option>
        </select>
      </div>

      {error ? (
        <ErrorBanner onRetry={reload}>{errorMessage(error)}</ErrorBanner>
      ) : loading && !tickets ? (
        <LoadingBlock label="Loading tickets…" />
      ) : tickets && tickets.length === 0 ? (
        <EmptyState title="No tickets yet">
          <p className="text-sm text-slate-500 dark:text-slate-400">Create one and the assistant will triage it.</p>
          <ButtonLink to="/tickets/new" variant="primary">
            Create the first ticket
          </ButtonLink>
        </EmptyState>
      ) : visible.length === 0 ? (
        <EmptyState title="No tickets match these filters">
          {filtered && (
            <button
              type="button"
              className="text-sm font-medium text-indigo-600 hover:underline dark:text-indigo-400"
              onClick={() => {
                setSearch('')
                setStatus('')
                setPriority('')
                setCategory('')
              }}
            >
              Clear filters
            </button>
          )}
        </EmptyState>
      ) : (
        <ul className="card divide-y divide-slate-200 overflow-hidden dark:divide-slate-800">
          {visible.map((t) => (
            <TicketRow key={t.id} ticket={t} />
          ))}
        </ul>
      )}
    </div>
  )
}

function TicketRow({ ticket: t }: { ticket: TicketListItem }) {
  return (
    <li>
      <Link
        to={`/tickets/${t.id}`}
        className="flex flex-col gap-2 px-4 py-3.5 transition-colors hover:bg-slate-50 sm:flex-row sm:items-center sm:gap-4 dark:hover:bg-slate-800/50"
      >
        <div className="min-w-0 flex-1">
          <div className="flex items-baseline gap-2">
            <span className="shrink-0 text-xs tabular-nums text-slate-400">#{t.id}</span>
            <span className="truncate font-medium">{t.title}</span>
          </div>
          <p className="mt-0.5 truncate text-sm text-slate-500 dark:text-slate-400">{t.triage?.summary ?? t.description}</p>
        </div>
        <div className="flex flex-wrap items-center gap-1.5 sm:w-64 sm:justify-end">
          {t.triage ? (
            <>
              <PriorityBadge priority={t.triage.priority} />
              <CategoryBadge category={t.triage.category} />
            </>
          ) : (
            <NotTriagedBadge />
          )}
          <StatusBadge status={t.status} />
        </div>
        <time
          dateTime={t.createdAt}
          title={fullDate(t.createdAt)}
          className="shrink-0 text-xs text-slate-500 sm:w-24 sm:text-right dark:text-slate-400"
        >
          {timeAgo(t.createdAt)}
        </time>
      </Link>
    </li>
  )
}

function rank(t: TicketListItem): number {
  return t.triage ? priorityRank[t.triage.priority] : 0
}

function countUntriaged(tickets: TicketListItem[]): number {
  return tickets.filter((t) => !t.triage).length
}
