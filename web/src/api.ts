import type { Article, NewTicket, Status, Ticket, TicketListItem, TriageWithArticles } from './types'

/** An error response from the Go API: the HTTP status plus its {"error": "..."} message. */
export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  let res: Response
  try {
    res = await fetch(`/api${path}`, {
      ...init,
      headers: { 'Content-Type': 'application/json', ...init.headers },
    })
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') throw err
    throw new ApiError(0, 'Could not reach the server. Is the Go backend running?')
  }

  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`
    try {
      const body = (await res.json()) as { error?: string }
      if (body.error) message = body.error
    } catch {
      // Not JSON (e.g. the Vite proxy's own error page); keep the status text.
    }
    throw new ApiError(res.status, message)
  }
  return (await res.json()) as T
}

export const api = {
  listTickets: (signal?: AbortSignal) => request<TicketListItem[]>('/tickets', { signal }),

  getTicket: (id: number, signal?: AbortSignal) => request<Ticket>(`/tickets/${id}`, { signal }),

  createTicket: (input: NewTicket) =>
    request<Ticket>('/tickets', { method: 'POST', body: JSON.stringify(input) }),

  updateStatus: (id: number, status: Status) =>
    request<Ticket>(`/tickets/${id}/status`, { method: 'PATCH', body: JSON.stringify({ status }) }),

  /** The latest triage, or null when the ticket hasn't been triaged yet. */
  getTriage: async (id: number, signal?: AbortSignal) => {
    try {
      return await request<TriageWithArticles>(`/tickets/${id}/triage`, { signal })
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) return null
      throw err
    }
  },

  runTriage: (id: number) => request<TriageWithArticles>(`/tickets/${id}/triage`, { method: 'POST' }),

  listArticles: (query: string, signal?: AbortSignal) => {
    const q = query.trim()
    return request<Article[]>(q ? `/kb?q=${encodeURIComponent(q)}&limit=20` : '/kb', { signal })
  },
}

export function errorMessage(err: unknown): string {
  if (err instanceof Error) return err.message
  return 'Something went wrong.'
}
