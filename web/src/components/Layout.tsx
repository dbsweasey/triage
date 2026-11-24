import type { ReactNode } from 'react'
import { useRoute } from '../router'
import { Link } from './Link'
import { ButtonLink } from './ui'

export function Layout({ children }: { children: ReactNode }) {
  const route = useRoute()
  const onTickets = route.name === 'tickets' || route.name === 'ticket'

  return (
    <div className="min-h-screen">
      <header className="sticky top-0 z-10 border-b border-slate-200 bg-white/85 backdrop-blur dark:border-slate-800 dark:bg-slate-950/85">
        <div className="mx-auto flex h-14 max-w-6xl items-center gap-4 px-4 sm:gap-6">
          <Link to="/" className="flex items-center gap-2 font-semibold tracking-tight">
            <img src="/favicon.svg" alt="" className="size-7" />
            <span className="hidden sm:inline">Triage</span>
          </Link>
          <nav className="flex items-center gap-1 text-sm">
            <NavLink to="/" active={onTickets}>
              Tickets
            </NavLink>
            <NavLink to="/kb" active={route.name === 'kb'}>
              Knowledge base
            </NavLink>
          </nav>
          <div className="ml-auto">
            {route.name !== 'new' && (
              <ButtonLink to="/tickets/new" variant="primary">
                <span aria-hidden>+</span> New<span className="hidden sm:inline"> ticket</span>
              </ButtonLink>
            )}
          </div>
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-4 py-6 sm:py-8">{children}</main>
    </div>
  )
}

function NavLink({ to, active, children }: { to: string; active: boolean; children: ReactNode }) {
  return (
    <Link
      to={to}
      aria-current={active ? 'page' : undefined}
      className={
        'rounded-md px-2.5 py-1.5 font-medium transition-colors ' +
        (active
          ? 'bg-slate-100 text-slate-900 dark:bg-slate-800 dark:text-white'
          : 'text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white')
      }
    >
      {children}
    </Link>
  )
}
