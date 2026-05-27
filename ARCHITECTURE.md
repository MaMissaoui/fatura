# Architecture

Fatura is a desktop application built on [Wails v2](https://wails.io), which embeds a React web app inside a native OS window backed by a Go process. There is no server, no network, and no cloud dependency — everything runs locally.

---

## High-level overview

```
┌─────────────────────────────────────────────────────┐
│                   OS Window (Wails)                  │
│                                                      │
│  ┌──────────────────────┐   ┌─────────────────────┐ │
│  │   React frontend     │   │    Go backend        │ │
│  │   (WebView)          │◄──►  (same process)      │ │
│  │                      │   │                      │ │
│  │  Jotai atoms         │   │  app.go methods      │ │
│  │  ↕ wailsjs bindings  │   │  ↕                   │ │
│  │                      │   │  db/ (SQLx + SQLite) │ │
│  └──────────────────────┘   └─────────────────────┘ │
└─────────────────────────────────────────────────────┘
                                      │
                               sqlite.db (disk)
```

The **Go process** owns the window and the database. The **React frontend** runs inside a WebView embedded in that window. They communicate through auto-generated function bindings — the frontend calls a Go method directly, as if it were a local async function, with no HTTP, no sockets, and no serialisation boilerplate.

---

## Technology stack

| Layer | Technology | Role |
|-------|-----------|------|
| Desktop shell | [Wails v2](https://wails.io) | Creates the OS window, bridges Go ↔ WebView |
| Backend language | Go 1.26 | Business logic, file I/O, database access |
| Database | SQLite via `modernc.org/sqlite` | Local persistent storage (CGO-free) |
| DB query layer | `jmoiron/sqlx` | Struct scanning, positional queries |
| DB migrations | `golang-migrate/migrate` | Versioned schema upgrades on startup |
| Frontend framework | React 19 | UI rendering |
| Build tool | Vite 8 (rolldown) | Dev server, bundler |
| State management | [Jotai](https://jotai.org) | Atomic reactive state; each atom owns one slice of data |
| UI components | [Ant Design 5](https://ant.design) | Component library |
| Routing | React Router 7 | Client-side routing inside the WebView |
| i18n | [LinguiJS 6](https://lingui.dev) | Translation extraction + runtime activation |
| PDF generation | `@react-pdf/renderer` | Renders invoice PDFs inside the browser context |
| Error tracking | Sentry (frontend JS only) | Crash reports; opt-in via `VITE_SENTRY_ENABLED` |

---

## Application startup sequence

### 1. Go entry point — `main.go`

```
main.go
  └── NewApp()           creates the App struct
  └── wails.Run(...)     starts Wails runtime:
        ├── opens OS window
        ├── starts embedded WebView
        ├── calls app.startup(ctx)   ← Go setup
        └── serves dist/ (prod) or proxies to Vite (dev)
```

`//go:embed all:dist` bundles the compiled frontend into the Go binary at build time. In development, Wails proxies the WebView to `http://localhost:5173` instead.

### 2. Backend setup — `app.go → startup()`

```
app.startup(ctx)
  ├── resolves platform config dir  (e.g. ~/Library/Application Support/Fatura/)
  ├── creates directory if missing
  └── db.NewDatabase(dbPath)
        ├── opens SQLite file (creates if absent)
        ├── sets PRAGMA journal_mode=WAL and foreign_keys=ON
        └── golang-migrate: runs all pending *.up.sql migrations
```

After `startup` returns, the `App` struct (with its live `*db.Database`) is bound to the WebView. Every public method on `App` becomes a callable async function in JavaScript.

### 3. HTML entry point — `index.html`

The WebView loads `index.html`, which contains a single `<div id="root">` and a `<script>` tag pointing to `src/main.tsx`.

### 4. React bootstrap — `src/main.tsx`

```
src/main.tsx
  └── ReactDOM.createRoot('#root').render(<App />)
        └── src/app.tsx
```

Applies a polyfill for `Promise.withResolvers` before mounting React.

### 5. Application shell — `src/app.tsx`

`app.tsx` does all one-time setup before rendering the route tree:

```
src/app.tsx  (App component)
  ├── src/styles/base.scss              global styles
  ├── src/utils/sentry.ts               Sentry init (calls GetVersion() → Go)
  ├── src/utils/lingui.tsx              i18n: loads default locale messages
  │
  └── AppContent component
        ├── reads localeAtom            (persisted in localStorage)
        ├── calls setOrganizationsAtom  → GetOrganizations() → Go → SQLite
        ├── activates LinguiJS locale
        ├── configures Ant Design locale
        │
        └── <MemoryRouter>
              └── <Routes>
                    ├── /                     → src/routes/index.tsx
                    ├── /organizations/new    → src/routes/organizations/new.tsx
                    └── /* (with BaseLayout)  → src/layouts/base.tsx
                          ├── /invoices/**
                          ├── /clients
                          ├── /projects
                          ├── /time-tracking/**
                          └── /settings/**
```

### 6. First-run redirect — `src/routes/index.tsx`

The `/` route acts as a smart redirect. It reads the loaded organizations and the persisted `organizationId` from localStorage, then decides:

| State | Action |
|-------|--------|
| No organizations exist | Navigate to `/organizations/new` |
| `organizationId` is set and valid | Navigate to `/invoices` |
| `organizationId` is stale (org deleted) | Clear it, stay on `/` |
| Organizations exist but none selected | Auto-select the first, navigate to `/invoices` |

### 7. Main layout — `src/layouts/base.tsx`

All authenticated routes render inside `BaseLayout`, which provides:
- The collapsible sidebar (Ant Design `Sider`) with navigation links
- The top header bar with the organization switcher and running-timer badge
- An `<Outlet />` where child routes render their content

---

## IPC: frontend ↔ backend

Wails generates a JavaScript module at `wailsjs/go/main/App.js` that contains one exported async function per public Go method on `App`. The frontend imports these directly:

```
src/atoms/*.ts
  └── import { GetClients, CreateClient, ... } from "wailsjs/go/main/App"
          │
          │  (Wails IPC bridge — synchronous from JS perspective)
          ▼
app.go  →  db/client.go  →  SQLite
```

The call flow for a data read (example: loading clients):

```
src/app.tsx
  └── setOrganizations()        [on mount]

src/layouts/base.tsx (any page that needs clients)
  └── useSetAtom(setClientsAtom)()

src/atoms/client.ts — setClientsAtom
  └── GetClients(organizationId)          ← imported from wailsjs/go/main/App
        │
        ▼ (Wails bridge)
app.go — GetClients(organizationID)
  └── db.GetClients(organizationID)

db/client.go — GetClients()
  └── sqlx.Select(&clients, "SELECT * FROM clients WHERE organizationId = ? ...")
        │
        ▼
  sqlite.db

        ▼ (return path)
[]db.Client  →  JSON  →  JavaScript array  →  set(clientsAtom, response)
```

---

## State management

Jotai atoms are the single source of truth for all application data. Each domain has its own file under `src/atoms/`:

| File | Atoms | Responsibility |
|------|-------|---------------|
| `generic.ts` | `localeAtom`, `siderAtom` | UI preferences (localStorage) |
| `organization.ts` | `organizationsAtom`, `organizationIdAtom`, `organizationAtom` | Multi-org support; `organizationId` persists across sessions |
| `invoice.ts` | `invoicesAtom`, `invoiceAtom`, `invoiceIdAtom` | Invoice list + selected invoice + CRUD |
| `client.ts` | `clientsAtom`, `clientAtom`, `clientIdAtom` | Client list + selected client + CRUD |
| `project.ts` | `projectsAtom`, `projectsByIdAtom` | Projects list + archive/unarchive |
| `tax-rate.ts` | `taxRatesAtom`, `taxRateAtom` | Tax rate list + CRUD |
| `time-tracking.ts` | `tagsAtom`, `timeEntriesAtom`, `runningTimerAtom` | Tags, time entries, active timer (localStorage) |

**Pattern:** every domain follows the same shape:
- A list atom (`clientsAtom`) holds the cached array
- A setter atom (`setClientsAtom`) fetches from Go and writes into the list atom
- A selected-ID atom (`clientIdAtom`) tracks which record is open
- A combined read/write atom (`clientAtom`) fetches one record and handles create/update

---

## Database layer

```
db/
├── db.go            Database struct, NewDatabase(), migration runner
├── migrations/      0001_base_models.up.sql … 0018_add_date_format_preference.up.sql
├── client.go        GetClients / GetClient / CreateClient / UpdateClient / DeleteClient / GetClientInvoiceCount
├── organization.go  GetOrganizations / GetOrganization / CreateOrganization / UpdateOrganization / DeleteOrganization
├── invoice.go       GetInvoices / GetInvoice / GetInvoiceLineItems / CreateInvoice / UpdateInvoice / UpdateInvoiceState / DeleteInvoice
├── project.go       GetProjects / GetProject / CreateProject / UpdateProject
├── tax_rate.go      GetTaxRates / GetTaxRate / CreateTaxRate / UpdateTaxRate / DeleteTaxRate
└── time_tracking.go GetTags / GetTag / CreateTag / UpdateTag / DeleteTag
                     GetTimeEntries / GetTimeEntry / CreateTimeEntry / UpdateTimeEntry / DeleteTimeEntry
```

Notable schema conventions:
- All primary keys are 21-character nanoid strings
- Monetary values (invoice totals, unit prices) are stored as **integer cents**
- Dates are stored as **Unix timestamps in milliseconds** (integer)
- Organization logo is stored as a **BLOB** (raw bytes of a data-URL string)
- Time entry tags are stored as a **JSON string** (array of tag names)

Transactions are used for operations that touch multiple tables: invoice create/update/delete (invoice + line items + org counter), and tax rate default toggling.

---

## Routing map

```
/                           → Index (redirect logic)
/organizations/new          → NewOrganization

[BaseLayout]
  /invoices                 → Invoices (list)
  /invoices/:id             → InvoiceDetails
  /invoices/:id/pdf         → InvoiceDetails (PDF preview mode)
  /clients                  → Clients (list + ClientForm modal)
  /projects                 → Projects (list)
  /time-tracking            → TimeTracking (timer + entries)
  /time-tracking/reports    → TimeTrackingReports
  /settings/organization    → SettingsOrganization
  /settings/invoice         → SettingsInvoice
  /settings/tax-rates       → SettingsTaxRates
  /settings/tax-rates/new   → TaxRateForm (nested modal route)
  /settings/tax-rates/:id   → TaxRateForm (nested modal route)
  /settings/backup          → SettingsBackup
```

---

## Internationalisation

```
lingui.config.ts          locales list, source paths
src/utils/lingui.tsx      loads default locale at module init; dynamicActivate(locale) on change
src/locales/*.po          one .po file per locale (source of truth for translators)
src/locales/*.js          compiled message catalogs (git-ignored, generated by pnpm extract)
```

String extraction: `pnpm extract` scans all `t\`...\`` and `<Trans>` macros and writes/updates the `.po` files. The LinguiJS Vite plugin compiles `.po` → JS at build time.

---

## Build artifacts

| Command | Output |
|---------|--------|
| `pnpm build` | `dist/` — compiled frontend (HTML + JS + CSS) |
| `wails build` | `build/bin/fatura` — single self-contained binary that embeds `dist/` |

The production binary contains the Go runtime, the WebView bridge, the SQLite driver, and the entire compiled React app. No external runtime or installer is required beyond the binary itself.
