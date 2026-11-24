import type { Article } from '../types'

/** Articles as expandable rows. Uses <details> so it works with keyboard and without JS state. */
export function ArticleList({ articles, defaultOpen = false }: { articles: Article[]; defaultOpen?: boolean }) {
  return (
    <ul className="divide-y divide-slate-200 dark:divide-slate-800">
      {articles.map((a) => (
        <li key={a.id}>
          <details className="group" open={defaultOpen}>
            <summary className="flex cursor-pointer list-none items-start gap-3 px-4 py-3 hover:bg-slate-50 dark:hover:bg-slate-800/50 [&::-webkit-details-marker]:hidden">
              <svg
                className="mt-0.5 size-4 shrink-0 text-slate-400 transition-transform group-open:rotate-90"
                viewBox="0 0 20 20"
                fill="currentColor"
                aria-hidden
              >
                <path d="M7.2 4.3a1 1 0 0 1 1.4 0l5 5a1 1 0 0 1 0 1.4l-5 5a1 1 0 1 1-1.4-1.4L11.5 10 7.2 5.7a1 1 0 0 1 0-1.4Z" />
              </svg>
              <div className="min-w-0 flex-1">
                <p className="text-sm font-medium">{a.title}</p>
                {a.tags.length > 0 && (
                  <div className="mt-1.5 flex flex-wrap gap-1">
                    {a.tags.map((tag) => (
                      <span
                        key={tag}
                        className="rounded bg-slate-100 px-1.5 py-0.5 text-[11px] text-slate-600 dark:bg-slate-800 dark:text-slate-400"
                      >
                        {tag}
                      </span>
                    ))}
                  </div>
                )}
              </div>
              <span className="shrink-0 text-xs text-slate-400">KB-{a.id}</span>
            </summary>
            <p className="whitespace-pre-wrap px-4 pb-4 pl-11 text-sm leading-relaxed text-slate-600 dark:text-slate-300">
              {a.body}
            </p>
          </details>
        </li>
      ))}
    </ul>
  )
}
