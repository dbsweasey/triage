import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { Link } from './Link'

type Variant = 'primary' | 'secondary' | 'ghost'

const buttonBase =
  'inline-flex items-center justify-center gap-2 rounded-lg px-3.5 py-2 text-sm font-medium transition-colors ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500 disabled:cursor-not-allowed disabled:opacity-60'

const variants: Record<Variant, string> = {
  primary: 'bg-indigo-600 text-white shadow-xs hover:bg-indigo-500',
  secondary:
    'border border-slate-300 bg-white text-slate-700 shadow-xs hover:bg-slate-50 ' +
    'dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200 dark:hover:bg-slate-800',
  ghost: 'text-slate-600 hover:bg-slate-100 hover:text-slate-900 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-100',
}

export function Button({
  variant = 'secondary',
  className = '',
  ...rest
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: Variant }) {
  return <button type="button" className={`${buttonBase} ${variants[variant]} ${className}`} {...rest} />
}

export function ButtonLink({ to, variant = 'secondary', children }: { to: string; variant?: Variant; children: ReactNode }) {
  return (
    <Link to={to} className={`${buttonBase} ${variants[variant]}`}>
      {children}
    </Link>
  )
}

export function Spinner({ className = 'size-4' }: { className?: string }) {
  return (
    <svg className={`animate-spin ${className}`} viewBox="0 0 24 24" fill="none" aria-hidden>
      <circle cx="12" cy="12" r="9" stroke="currentColor" strokeOpacity="0.25" strokeWidth="3" />
      <path d="M21 12a9 9 0 0 0-9-9" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
    </svg>
  )
}

export function ErrorBanner({ children, onRetry }: { children: ReactNode; onRetry?: () => void }) {
  return (
    <div
      role="alert"
      className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800 dark:border-red-900/60 dark:bg-red-950/40 dark:text-red-200"
    >
      <span>{children}</span>
      {onRetry && (
        <button type="button" onClick={onRetry} className="font-medium underline underline-offset-2 hover:no-underline">
          Try again
        </button>
      )}
    </div>
  )
}

export function LoadingBlock({ label = 'Loading…' }: { label?: string }) {
  return (
    <div className="flex items-center justify-center gap-2 py-16 text-sm text-slate-500 dark:text-slate-400">
      <Spinner /> {label}
    </div>
  )
}

export function EmptyState({ title, children }: { title: string; children?: ReactNode }) {
  return (
    <div className="card flex flex-col items-center gap-3 px-6 py-14 text-center">
      <p className="font-medium">{title}</p>
      {children}
    </div>
  )
}

export function Field({ label, hint, htmlFor, children }: { label: string; hint?: string; htmlFor: string; children: ReactNode }) {
  return (
    <div className="space-y-1.5">
      <label htmlFor={htmlFor} className="block text-sm font-medium">
        {label}
        {hint && <span className="ml-1.5 font-normal text-slate-500 dark:text-slate-400">{hint}</span>}
      </label>
      {children}
    </div>
  )
}
