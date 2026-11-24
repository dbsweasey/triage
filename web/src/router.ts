import { useSyncExternalStore } from 'react'

// A tiny path-based router: enough for four pages without adding a dependency.
// Links live in components/Link.tsx.

export type Route =
  | { name: 'tickets' }
  | { name: 'new' }
  | { name: 'ticket'; id: number }
  | { name: 'kb' }
  | { name: 'notFound' }

const NAV_EVENT = 'triage:navigate'

function parse(pathname: string): Route {
  if (pathname === '/' || pathname === '/tickets') return { name: 'tickets' }
  if (pathname === '/tickets/new') return { name: 'new' }
  if (pathname === '/kb') return { name: 'kb' }
  const m = pathname.match(/^\/tickets\/(\d+)$/)
  if (m) return { name: 'ticket', id: Number(m[1]) }
  return { name: 'notFound' }
}

function subscribe(onChange: () => void) {
  window.addEventListener('popstate', onChange)
  window.addEventListener(NAV_EVENT, onChange)
  return () => {
    window.removeEventListener('popstate', onChange)
    window.removeEventListener(NAV_EVENT, onChange)
  }
}

export function useRoute(): Route {
  const pathname = useSyncExternalStore(subscribe, () => window.location.pathname)
  return parse(pathname)
}

/** Navigate without a page reload. `state` is available afterwards as history.state. */
export function navigate(to: string, state?: unknown) {
  window.history.pushState(state ?? null, '', to)
  window.dispatchEvent(new Event(NAV_EVENT))
  window.scrollTo(0, 0)
}
