import { useEffect, useState } from 'react'
import { api, ApiError, errorMessage } from '../api'
import { ArticleList } from '../components/ArticleList'
import { CategoryBadge, PriorityBadge, StatusBadge } from '../components/Badges'
import { Button, ErrorBanner, LoadingBlock, Spinner } from '../components/ui'
import { fullDate, statusLabel, stripStepNumber, timeAgo } from '../format'
import { Link } from '../components/Link'
import { STATUSES, type Status, type Ticket, type TriageWithArticles } from '../types'
import { useLoad } from '../useLoad'
import type { TicketPageState } from './NewTicket'

interface Loaded {
  ticket: Ticket
  triage: TriageWithArticles | null
}

export function TicketDetail({ id }: { id: number }) {
  const { data, error, loading, reload, setData } = useLoad<Loaded>(
    async (signal) => {
      const [ticket, triage] = await Promise.all([api.getTicket(id, signal), api.getTriage(id, signal)])
      return { ticket, triage }
    },
    id,
  )

  if (error) {
    const missing = error instanceof ApiError && error.status === 404
    return (
      <div className="space-y-4">
        <BackLink />
        {missing ? (
          <ErrorBanner>Ticket #{id} doesn't exist.</ErrorBanner>
        ) : (
          <ErrorBanner onRetry={reload}>{errorMessage(error)}</ErrorBanner>
        )}
      </div>
    )
  }
  if (loading && !data) return <LoadingBlock label="Loading ticket…" />
  if (!data) return null

  return (
    <TicketView
      ticket={data.ticket}
      triage={data.triage}
      onTicket={(ticket) => setData({ ...data, ticket })}
      onTriage={(triage) => setData({ ...data, triage })}
    />
  )
}

function TicketView({
  ticket,
  triage,
  onTicket,
  onTriage,
}: {
  ticket: Ticket
  triage: TriageWithArticles | null
  onTicket: (t: Ticket) => void
  onTriage: (t: TriageWithArticles) => void
}) {
  return (
    <div className="space-y-6">
      <div>
        <BackLink />
        <div className="mt-2 flex flex-wrap items-start gap-x-3 gap-y-2">
          <h1 className="text-2xl font-semibold tracking-tight">
            <span className="mr-2 font-normal text-slate-400">#{ticket.id}</span>
            {ticket.title}
          </h1>
          <div className="flex items-center gap-1.5 pt-1.5">
            <StatusBadge status={ticket.status} />
            {triage && <PriorityBadge priority={triage.priority} />}
          </div>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_18rem]">
        <div className="space-y-6">
          <section className="card p-5">
            <h2 className="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">Description</h2>
            <p className="mt-2 whitespace-pre-wrap text-sm leading-relaxed">{ticket.description}</p>
          </section>
          <TriagePanel ticketId={ticket.id} triage={triage} onTriage={onTriage} />
        </div>

        <aside className="space-y-6">
          <DetailsCard ticket={ticket} onTicket={onTicket} />
        </aside>
      </div>
    </div>
  )
}

function TriagePanel({
  ticketId,
  triage,
  onTriage,
}: {
  ticketId: number
  triage: TriageWithArticles | null
  onTriage: (t: TriageWithArticles) => void
}) {
  const [running, setRunning] = useState(false)
  // A triage error from the New Ticket page arrives through history state.
  const [error, setError] = useState<string | undefined>(() => (window.history.state as TicketPageState | null)?.triageError)
  useEffect(() => {
    // Show that error once; don't bring it back on refresh.
    if ((window.history.state as TicketPageState | null)?.triageError) window.history.replaceState(null, '')
  }, [])

  async function run() {
    setRunning(true)
    setError(undefined)
    try {
      onTriage(await api.runTriage(ticketId))
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setRunning(false)
    }
  }

  const header = (
    <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-5 py-3.5 dark:border-slate-800">
      <div className="flex items-center gap-2">
        <SparkIcon />
        <h2 className="font-semibold">AI triage</h2>
        {triage && !running && (
          <time dateTime={triage.createdAt} title={fullDate(triage.createdAt)} className="text-xs text-slate-500 dark:text-slate-400">
            {timeAgo(triage.createdAt)}
          </time>
        )}
      </div>
      {triage && (
        <Button onClick={run} disabled={running}>
          {running && <Spinner />}
          {running ? 'Re-triaging…' : 'Re-triage'}
        </Button>
      )}
    </div>
  )

  return (
    <section className="card overflow-hidden">
      {header}
      {error && (
        <div className="px-5 pt-4">
          <ErrorBanner onRetry={running ? undefined : run}>Triage failed: {error}</ErrorBanner>
        </div>
      )}

      {!triage ? (
        <div className="flex flex-col items-center gap-3 px-6 py-12 text-center">
          <p className="font-medium">{running ? 'Triaging this ticket…' : 'This ticket hasn’t been triaged yet'}</p>
          <p className="max-w-sm text-sm text-slate-500 dark:text-slate-400">
            {running
              ? 'Classifying priority and category, and matching knowledge base articles. This can take up to 30 seconds.'
              : 'The assistant will classify priority and category, summarize the issue, and suggest steps and articles.'}
          </p>
          <Button variant="primary" onClick={run} disabled={running}>
            {running && <Spinner />}
            {running ? 'Triaging…' : 'Run triage'}
          </Button>
        </div>
      ) : (
        <div className={`space-y-6 p-5 transition-opacity ${running ? 'opacity-50' : ''}`}>
          <dl className="flex flex-wrap gap-x-8 gap-y-3">
            <div>
              <dt className="text-xs text-slate-500 dark:text-slate-400">Priority</dt>
              <dd className="mt-1">
                <PriorityBadge priority={triage.priority} />
              </dd>
            </div>
            <div>
              <dt className="text-xs text-slate-500 dark:text-slate-400">Category</dt>
              <dd className="mt-1">
                <CategoryBadge category={triage.category} />
              </dd>
            </div>
          </dl>

          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">Summary</h3>
            <p className="mt-2 text-sm leading-relaxed">{triage.summary}</p>
          </div>

          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">Resolution steps</h3>
            <ol className="mt-3 space-y-2.5">
              {triage.steps.map((step, i) => (
                <li key={i} className="flex gap-3 text-sm leading-relaxed">
                  <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-indigo-50 text-xs font-semibold text-indigo-700 dark:bg-indigo-500/15 dark:text-indigo-300">
                    {i + 1}
                  </span>
                  <span className="pt-0.5">{stripStepNumber(step)}</span>
                </li>
              ))}
            </ol>
          </div>

          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">
              Suggested articles
            </h3>
            {triage.articles.length === 0 ? (
              <p className="mt-2 text-sm text-slate-500 dark:text-slate-400">No knowledge base articles matched this ticket.</p>
            ) : (
              <div className="mt-3 overflow-hidden rounded-lg border border-slate-200 dark:border-slate-800">
                <ArticleList articles={triage.articles} />
              </div>
            )}
          </div>
        </div>
      )}
    </section>
  )
}

function DetailsCard({ ticket, onTicket }: { ticket: Ticket; onTicket: (t: Ticket) => void }) {
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string>()

  async function changeStatus(status: Status) {
    setSaving(true)
    setError(undefined)
    try {
      onTicket(await api.updateStatus(ticket.id, status))
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setSaving(false)
    }
  }

  return (
    <section className="card space-y-4 p-5">
      <div className="space-y-1.5">
        <label htmlFor="status" className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">
          Status {saving && <Spinner className="size-3" />}
        </label>
        <select
          id="status"
          className="control"
          value={ticket.status}
          onChange={(e) => changeStatus(e.target.value as Status)}
          disabled={saving}
        >
          {STATUSES.map((s) => (
            <option key={s} value={s}>
              {statusLabel[s]}
            </option>
          ))}
        </select>
        {error && <p className="text-xs text-red-600 dark:text-red-400">{error}</p>}
      </div>

      <dl className="space-y-3 text-sm">
        <div>
          <dt className="text-xs text-slate-500 dark:text-slate-400">Requester</dt>
          <dd className="mt-0.5 break-all">
            {ticket.requesterEmail ? (
              <a href={`mailto:${ticket.requesterEmail}`} className="text-indigo-600 hover:underline dark:text-indigo-400">
                {ticket.requesterEmail}
              </a>
            ) : (
              <span className="text-slate-400">Not provided</span>
            )}
          </dd>
        </div>
        <div>
          <dt className="text-xs text-slate-500 dark:text-slate-400">Created</dt>
          <dd className="mt-0.5">
            <time dateTime={ticket.createdAt}>{fullDate(ticket.createdAt)}</time>
          </dd>
        </div>
      </dl>
    </section>
  )
}

function BackLink() {
  return (
    <Link to="/" className="text-sm text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white">
      ← Tickets
    </Link>
  )
}

function SparkIcon() {
  return (
    <svg className="size-4 text-indigo-500" viewBox="0 0 20 20" fill="currentColor" aria-hidden>
      <path d="M10 1.5l1.9 4.6 4.6 1.9-4.6 1.9L10 14.5l-1.9-4.6L3.5 8l4.6-1.9L10 1.5Zm5.5 10.5.9 2.1 2.1.9-2.1.9-.9 2.1-.9-2.1-2.1-.9 2.1-.9.9-2.1Z" />
    </svg>
  )
}
