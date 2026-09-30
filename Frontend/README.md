# PulseGraph Frontend

React + TypeScript dashboard for **PulseGraph** - the AI social media
intelligence platform (Smart India Hackathon 2026, Problem Statement 26152).

The app ships a deterministic mock-data engine, so the whole product - 3
projects, 5 monitored topics, 1,000+ simulated posts, sentiment, trends,
audience, network graph and AI reports - is fully demonstrable without a
backend.

## Stack

| Concern | Choice |
| --- | --- |
| Framework | React 18 + TypeScript (strict) |
| Build | Vite 5 |
| Styling | Tailwind CSS 3 (dark-first design system) |
| UI primitives | Hand-written shadcn/ui-style components |
| State | Zustand (session + UI) |
| Data fetching | TanStack React Query v5 |
| Routing | React Router DOM v6 |
| Charts | Recharts |
| Network graph | @xyflow/react (React Flow) |
| Animation | Framer Motion |
| Forms | React Hook Form + Zod |
| Icons | Lucide React |
| HTTP | Axios (typed API clients per domain) |

## Getting started

```bash
npm install
npm run dev        # http://localhost:5173
npm run build      # production build -> dist/
npm run preview    # serve the production build locally
```

Demo credentials (mock auth): `analyst@pulsegraph.dev` / `PulseGraph@2026`.

## Environment

Copy `.env.example` to `.env` (or set build-time variables in CI):

| Variable | Purpose |
| --- | --- |
| `VITE_API_URL` | Base URL of the PulseGraph Go backend, e.g. `http://localhost:8080/api/v1`. Empty = mock mode. |
| `VITE_USE_MOCKS` | `true` forces the mock engine even when `VITE_API_URL` is set. |
| `VITE_WS_URL` | Optional live WebSocket endpoint; without it a deterministic simulator drives the live feed. |

## Mock mode (default)

`src/services/mock/engine.ts` generates a coherent, seeded dataset:

- 3 projects (flood rumour watch, city events, product radar)
- 5 monitored topics with 1,000+ posts across three narrative arcs
- sentiment arcs, trend series, keyword clusters, influence graphs and
  aggregated audience distributions
- grounded AI reports computed from the same aggregates

`src/services/websocket.ts` simulates a live stream (new posts every few
seconds), so the dashboard feed animates without any backend.

## Structure

```
src/
├── app/          router, providers, constants
├── components/   ui primitives, charts, graphs, dashboard widgets, panels
├── features/     auth, projects, topics, timeline, sentiment, demographics,
│                 trends, network, reports (typed API clients + components)
├── hooks/        useIntelligence (query hooks), useLiveFeed, useDebounce
├── layouts/      DashboardLayout, Sidebar, AuthLayout
├── pages/        Landing, Login, Register, Projects, Dashboard, Analytics,
│                 Timeline, Sentiment, Audience, Trends, Network, Reports, 404
├── services/     api (axios + errors), endpoints (typed clients), websocket,
│                 mock/ (deterministic data engine)
├── stores/       Zustand UI store (+ persisted auth store in features/auth)
├── types/        shared domain types
└── utils/        cn (class merge), format helpers
```

## Self-contained build

The production build is designed for a strict-CSP static host: no Google
Fonts, no CDN scripts, no remote images and no external API calls in the
default (mock) build. Fonts use a system stack; every asset is bundled by
Vite.

## Connecting the Go backend

Run the backend (`../Backend`) and build with:

```bash
VITE_API_URL=http://localhost:8080/api/v1 npm run build
```

`src/services/endpoints.ts` contains best-effort adapters that map the
backend REST shapes onto the frontend types; endpoints not yet wired fall
back to error states with retry, never to fabricated data.
