# Triage web

React 19 + TypeScript + Tailwind v4, built with Vite.

## Develop

Run the Go API and the Vite dev server side by side:

```sh
# terminal 1, repo root (no OPENAI_API_KEY = fake triager)
go run ./cmd/triage

# terminal 2
cd web
npm install
npm run dev        # http://localhost:5173
```

Vite proxies `/api/*` to `http://localhost:8080` (see `vite.config.ts`), so no CORS setup is needed.

## Layout

| Path | What it is |
|---|---|
| `src/types.ts` | Types mirroring the Go JSON (`store.Ticket`, `store.Triage`, …) |
| `src/api.ts` | Typed fetch client; non-2xx responses throw `ApiError` with the server's `error` message |
| `src/router.ts`, `components/Link.tsx` | Tiny path-based router (`/`, `/tickets/new`, `/tickets/:id`, `/kb`) |
| `src/useLoad.ts` | Load-on-mount hook with cancellation, `reload` and `setData` |
| `src/pages/` | Ticket list, new ticket, ticket detail (triage, status), knowledge base |

## Checks

```sh
npm run build   # type-check + production build into dist/
npm run lint    # oxlint
```
