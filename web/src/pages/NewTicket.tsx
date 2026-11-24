import { useState, type FormEvent } from 'react'
import { api, errorMessage } from '../api'
import { Button, ErrorBanner, Field, Spinner } from '../components/ui'
import { Link } from '../components/Link'
import { navigate } from '../router'

type Phase = 'idle' | 'creating' | 'triaging'

/** Passed to the detail page through history state when triage fails right after creating. */
export interface TicketPageState {
  triageError?: string
}

export function NewTicket() {
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [email, setEmail] = useState('')
  const [runTriage, setRunTriage] = useState(true)
  const [phase, setPhase] = useState<Phase>('idle')
  const [error, setError] = useState<string>()

  const busy = phase !== 'idle'

  async function submit(e: FormEvent) {
    e.preventDefault()
    setError(undefined)
    setPhase('creating')

    let id: number
    try {
      const ticket = await api.createTicket({ title: title.trim(), description: description.trim(), requesterEmail: email.trim() })
      id = ticket.id
    } catch (err) {
      setError(errorMessage(err))
      setPhase('idle')
      return
    }

    if (!runTriage) {
      navigate(`/tickets/${id}`)
      return
    }

    setPhase('triaging')
    try {
      await api.runTriage(id)
      navigate(`/tickets/${id}`)
    } catch (err) {
      // The ticket exists; show it with the triage error so the user can retry from there.
      navigate(`/tickets/${id}`, { triageError: errorMessage(err) } satisfies TicketPageState)
    }
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <Link to="/" className="text-sm text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white">
          ← Tickets
        </Link>
        <h1 className="mt-2 text-2xl font-semibold tracking-tight">New ticket</h1>
        <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
          Describe the problem. The assistant will set a priority and category, summarize it, and suggest resolution steps.
        </p>
      </div>

      <form onSubmit={submit} className="card space-y-5 p-5 sm:p-6">
        {error && <ErrorBanner>{error}</ErrorBanner>}

        <Field label="Title" htmlFor="title">
          <input
            id="title"
            className="control"
            required
            maxLength={200}
            placeholder="e.g. VPN keeps disconnecting"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            disabled={busy}
            autoFocus
          />
        </Field>

        <Field label="Description" htmlFor="description">
          <textarea
            id="description"
            className="control min-h-36 resize-y"
            required
            placeholder="What happened, when it started, what you've already tried, and how many people are affected."
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            disabled={busy}
          />
        </Field>

        <Field label="Requester email" hint="optional" htmlFor="email">
          <input
            id="email"
            type="email"
            className="control"
            placeholder="name@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            disabled={busy}
          />
        </Field>

        <label className="flex items-start gap-3 text-sm">
          <input
            type="checkbox"
            className="mt-0.5 size-4 accent-indigo-600"
            checked={runTriage}
            onChange={(e) => setRunTriage(e.target.checked)}
            disabled={busy}
          />
          <span>
            <span className="font-medium">Triage with AI after creating</span>
            <span className="block text-slate-500 dark:text-slate-400">You can also run or re-run triage from the ticket page.</span>
          </span>
        </label>

        <div className="flex flex-wrap items-center justify-end gap-3 border-t border-slate-200 pt-5 dark:border-slate-800">
          {phase === 'triaging' && (
            <span className="mr-auto text-sm text-slate-500 dark:text-slate-400">This can take up to 30 seconds.</span>
          )}
          <Button variant="ghost" onClick={() => navigate('/')} disabled={busy}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={busy || !title.trim() || !description.trim()}>
            {busy && <Spinner />}
            {phase === 'creating' ? 'Creating…' : phase === 'triaging' ? 'Triaging…' : 'Create ticket'}
          </Button>
        </div>
      </form>
    </div>
  )
}
