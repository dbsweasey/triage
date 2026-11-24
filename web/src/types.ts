// These mirror the Go structs' JSON tags in internal/store and cmd/triage.

export const PRIORITIES = ['high', 'medium', 'low'] as const
export const CATEGORIES = ['account', 'email', 'hardware', 'software', 'network', 'security', 'other'] as const
export const STATUSES = ['open', 'in_progress', 'resolved', 'closed'] as const

export type Priority = (typeof PRIORITIES)[number]
export type Category = (typeof CATEGORIES)[number]
export type Status = (typeof STATUSES)[number]

export interface Ticket {
  id: number
  title: string
  description: string
  requesterEmail: string
  status: Status
  createdAt: string
}

export interface Triage {
  id: number
  ticketId: number
  priority: Priority
  category: Category
  summary: string
  steps: string[]
  articleIds: number[]
  createdAt: string
}

export interface Article {
  id: number
  title: string
  body: string
  tags: string[]
}

/** GET /api/tickets: each ticket with its latest triage, or null if never triaged. */
export interface TicketListItem extends Ticket {
  triage: Triage | null
}

/** GET/POST /api/tickets/{id}/triage: the triage plus the full suggested articles. */
export interface TriageWithArticles extends Triage {
  articles: Article[]
}

export interface NewTicket {
  title: string
  description: string
  requesterEmail: string
}
