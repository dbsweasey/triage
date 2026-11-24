import { useEffect, useState } from 'react'
import { api, errorMessage } from '../api'
import { ArticleList } from '../components/ArticleList'
import { EmptyState, ErrorBanner, LoadingBlock, Spinner } from '../components/ui'
import { useLoad } from '../useLoad'

export function KnowledgeBase() {
  const [input, setInput] = useState('')
  const query = useDebounced(input.trim(), 250)
  const { data: articles, error, loading, reload } = useLoad((signal) => api.listArticles(query, signal), query)

  return (
    <div className="space-y-5">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Knowledge base</h1>
        <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
          The articles the assistant can suggest. Search uses the same full-text index as triage.
        </p>
      </div>

      <div className="relative">
        <input
          type="search"
          className="control pr-9"
          placeholder="Search articles, e.g. “vpn disconnecting” or “forgot password”"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          aria-label="Search articles"
        />
        {loading && articles && (
          <Spinner className="absolute top-1/2 right-3 size-4 -translate-y-1/2 text-slate-400" />
        )}
      </div>

      {error ? (
        <ErrorBanner onRetry={reload}>{errorMessage(error)}</ErrorBanner>
      ) : !articles ? (
        <LoadingBlock label="Loading articles…" />
      ) : articles.length === 0 ? (
        <EmptyState title={query ? `No articles match “${query}”` : 'The knowledge base is empty'} />
      ) : (
        <div className="space-y-2">
          <p className="text-sm text-slate-500 dark:text-slate-400">
            {query ? `${articles.length} ${articles.length === 1 ? "match" : "matches"}, most relevant first` : `${articles.length} articles`}
          </p>
          <div className="card overflow-hidden">
            <ArticleList articles={articles} />
          </div>
        </div>
      )}
    </div>
  )
}

function useDebounced<T>(value: T, ms: number): T {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), ms)
    return () => clearTimeout(id)
  }, [value, ms])
  return debounced
}
