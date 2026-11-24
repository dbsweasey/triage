import type { AnchorHTMLAttributes, MouseEvent } from 'react'
import { navigate } from '../router'

type LinkProps = AnchorHTMLAttributes<HTMLAnchorElement> & { to: string }

/** An <a> that navigates without a page reload, but still opens new tabs normally. */
export function Link({ to, onClick, ...rest }: LinkProps) {
  function handleClick(e: MouseEvent<HTMLAnchorElement>) {
    onClick?.(e)
    if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return
    e.preventDefault()
    navigate(to)
  }
  return <a href={to} onClick={handleClick} {...rest} />
}
