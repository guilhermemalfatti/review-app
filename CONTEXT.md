# Indica — Project Context (for future sessions)

> **Purpose:** Hand this file to a new chat so the agent has full product, architecture, and operational context without rediscovering the repo.
>
> **Last updated:** 2026-07-21  
> **Repo path:** `/Users/gmalfatti/projects-personal/review-app`  
> **Working name:** Indica (condo-scoped trusted service-provider network)

---

## 1. What this product is

**Indica** is a web app for condominium residents to share **trusted service-provider indications** (electrician, plumber, painter, etc.). It is intentionally *not* a full condo app (contrast with Indicaz / Vizis). Launch community: **Cantegril** (Brazil).

### Product principles (from discovery)

- Trust comes from **neighbor indications**, not Google Maps–style anonymous reviews alone.
- Example trust line: “42 hired · 38 recommend · 4 don’t · avg 4.8 · last service 2 weeks ago.”
- **Public browse** of approved providers/reviews; **login required** to write (review / suggest provider).
- Soft gate: signup needs a shared **invite code** (not a hard email allowlist).
- **Admin moderation** for new providers and every new/updated review (language / quality control).
- Reviewer can choose **show name** or **anonymous**; identity always stored for uniqueness/moderation; public sees `author_label` (`display_name` or `Anônimo`).
- One review version per submission (append-only); resubmit **creates a new pending row** and leaves the previous **approved** review published until the new one is approved (then old → `superseded`) or rejected (old stays approved). Rows are never deleted.
- Scores: optional 1–5 on **price / quality / deadline** + recommend yes/no + comment + service date.
- UI language: **Portuguese (Brazil)** for residents; code/identifiers in English.
- Audience includes **older adults** → keep admin and scores simple (tabs, large buttons, stars + “X de 5 — Bom”).

### Out of scope (v1)

- Review photos
- Magic link / OAuth
- Email verification / forgot-password email flow
- Multi-condo picker UI (schema is multi-condo ready; runtime is single seeded condo)
- Sidecar / API gateway auth
- Payments, WhatsApp bots, push notifications

### Planned deploy

| Piece | Target |
|-------|--------|
| Postgres | Supabase (DB only — **not** Supabase Auth) |
| Go API + SPA | **Render** Docker from **repo-root** `Dockerfile` (same origin) |

Use Supabase **Session** pooler (`sslmode=require`, prefer port `5432` on `*.pooler.supabase.com`) — transaction pooler (`6543`) breaks goose/pgx prepared statements. Append `default_query_exec_mode=simple_protocol` if needed.

Same-origin build leaves `VITE_API_URL` empty and `VITE_BASE_PATH=/`. Set Render `CORS_ORIGIN` to the Render service URL (e.g. `https://review-app-y7fl.onrender.com`). Use `COOKIE_SAMESITE=Lax` (default). GitHub Pages is optional/legacy and needs `COOKIE_SAMESITE=None` (often broken on iOS).

Local first; production secrets live only in Render / GitHub — never in `.env.example`.

---

## 2. Monorepo layout

```
review-app/
├── backend/                 # Go module: github.com/gmalfatti/indica/backend
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── auth/            # sessions, bcrypt, temp passwords
│   │   ├── config/
│   │   ├── db/              # connect, migrate (goose), seed, seed_demo, reset
│   │   └── http/
│   │       ├── router.go
│   │       ├── handlers/    # auth, providers, admin, helpers
│   │       └── middleware/  # auth, csrf, ratelimit
│   ├── migrations/          # goose SQL
│   └── Dockerfile           # non-root appuser
├── frontend/                # Vite + React + TypeScript
│   └── src/
│       ├── api/             # client + types (CSRF-aware fetch)
│       ├── auth/            # AuthContext
│       ├── components/
│       ├── config.ts        # COMMUNITY_NAME, CATEGORIES
│       ├── lib/format.ts
│       ├── pages/
│       └── styles/          # plain CSS (no Tailwind)
├── docker-compose.yml       # Postgres 16
├── .env.example
├── README.md                # runbook
└── CONTEXT.md               # this file
```

---

## 3. Stack

| Layer | Choice |
|-------|--------|
| API | Go 1.26+, chi, pgx, goose, bcrypt |
| DB | Postgres (Docker locally) |
| FE | React + Vite + TypeScript (`strict: true`), React Router |
| CSS | Plain CSS; Fraunces + Nunito Sans via `@fontsource` (self-hosted) |
| Auth | DIY email/password; **HTTP-only session cookie** + server-side `sessions` table (not JWT) |
| CSRF | Double-submit: cookie `csrf` + header `X-CSRF-Token` |
| Design | Coastal/residential (sand/stone/leaf greens); avoid generic AI purple/cream/terracotta looks |

---

## 4. Domain model

### Tables (high level)

- **condos** — `id`, `name`, `slug` (unique), `invite_code`, `created_at`
- **users** — `condo_id`, `email`, `password_hash`, `display_name`, `role` (`resident`|`admin`), `must_change_password`, `created_at`
  - Unique: **`(condo_id, email)`** (migration `00003`)
- **sessions** — `user_id`, `token_hash` (SHA-256 of raw cookie token), `expires_at`
- **providers** — `condo_id`, `name`, `category`, `phone`, `notes`, `status` (`pending`|`approved`|`rejected`), `created_by`, `reviewed_by`, `reviewed_at`
  - Unique: **`(condo_id, name)`** (migration `00003`)
- **reviews** — `provider_id`, `user_id`, `is_anonymous`, `recommend`, scores, `comment`, `service_date`, `status` (`pending`|`approved`|`rejected`|`superseded`), `reviewed_by`, `reviewed_at`
  - Append-only versions (migration `00005`); partial unique: one `pending` and one `approved` per `(user_id, provider_id)`
- **audit_events** — append-only activity log: `condo_id`, `actor_user_id`, `action`, `entity_type`, `entity_id`, `payload` (JSONB), `created_at`
  - Actions include `provider.created|approved|rejected`, `review.created|approved|rejected|superseded`, `user.password_reset`
  - Written in the **same transaction** as the state change
  - `reviewed_*` columns hold the **latest** moderator; full history lives in `audit_events`

### Aggregates (approved reviews only)

`hired_count`, `recommend_count`, `not_recommend_count`, `avg_price`, `avg_quality`, `avg_deadline`, `avg_overall`, `last_service_date`

### Runtime condo scope

API boots with seeded condo **Cantegril** (`slug=cantegril`) and scopes queries to that `condo_id`. Invite **check for signup uses env `INVITE_CODE`**, not a live DB lookup every time. Condo row `invite_code` is set on create-once seed only (not overwritten every boot).

---

## 5. Auth & security (important)

### Session model

1. Login/signup validates password → creates random token → stores **hash** in `sessions` → `Set-Cookie: session=<token>` (**HttpOnly**).
2. Browser sends cookie automatically (`credentials: 'include'`).
3. Middleware loads user (incl. **role**) from DB via session — role is **not** in a JWT.
4. Login/signup: **single active session** (delete others, then create).
5. Logout: delete session row + clear cookie.
6. Hourly cleanup of expired sessions in the API process.

### Roles

- `resident` — suggest providers, submit reviews
- `admin` — moderation queues + reset passwords

Frontend learns role from `/api/auth/me` (AuthContext). **Authorization is always enforced on the Go server.**

### CSRF

- `GET /api/auth/csrf` → sets non-HttpOnly `csrf` cookie + returns `{csrf_token}`
- All mutating `/api/*` methods require matching `X-CSRF-Token`
- Frontend `api/client.ts` caches token, attaches header, retries once on CSRF 403

### Other hardening already in place

- Password min length **8** (FE + BE)
- Dummy bcrypt hash on unknown email (timing / enumeration mitigation)
- Constant-time invite compare
- Login/signup IP rate limit (~20 / 15 min)
- JSON body max 1MiB
- Phones **hidden** on public provider list/detail until logged in
- `must_change_password` blocks writes/admin until `POST /api/auth/change-password`
- Admin password reset: temp password + revoke sessions + force change; transactional
- Change password: revoke all sessions, issue fresh session
- Production fail-closed (`APP_ENV=production`): no weak invite/admin password, `COOKIE_SECURE` required, `RESET_DB` refused
- `SESSION_SECRET` was removed (unused); do not reintroduce without a use

### CookieSecure

From `COOKIE_SECURE` env, or default `true` when `APP_ENV=production`.

---

## 6. API contract (summary)

Base: `/api`. Cookie: `session`. CSRF header on mutations.

| Method | Path | Who |
|--------|------|-----|
| GET | `/health` | public |
| GET | `/auth/csrf` | public |
| POST | `/auth/signup` | public + invite |
| POST | `/auth/login` | public |
| POST | `/auth/logout` | optional |
| GET | `/auth/me` | session |
| POST | `/auth/change-password` | session |
| GET | `/providers` | public (approved; phone if authed) |
| GET | `/providers/:id` | public (approved; admin can see non-approved) |
| GET | `/providers/:id/my-review` | session |
| POST | `/providers` | session → pending |
| POST | `/providers/:id/reviews` | session → pending (new row; prior approved stays live) |
| GET/POST | `/admin/providers…` | admin |
| GET/POST | `/admin/reviews…` | admin |
| GET | `/admin/users` | admin |
| POST | `/admin/users/:id/reset-password` | admin → `{temporary_password, user}` |

List endpoints return **JSON arrays** (not wrapped objects).

Errors: `{ "error": "message" }`.

---

## 7. Frontend routes & UX

| Route | Page |
|-------|------|
| `/` | Provider list (search + category chips) |
| `/providers/:id` | Detail: score board (stars + hints) + timeline of reviews |
| `/providers/:id/review` | Submit/update indication; warn if replacing existing |
| `/providers/new` | Suggest provider |
| `/login`, `/signup` | Auth (invite on signup) |
| `/change-password` | Forced after admin reset |
| `/admin` | Tabs: Novos prestadores / Novas indicações / Moradores (simple PT copy) |

Config: `frontend/src/config.ts` → `COMMUNITY_NAME = 'Cantegril'`, `CATEGORIES` (FE chips; backend category is free text).

Vite proxy: `/api` → `VITE_API_PROXY` or `http://localhost:8080`.

---

## 8. Seeders & demo data

| Mechanism | Role |
|-----------|------|
| `Seed` | Create-once Cantegril condo + bootstrap admin from env (**does not** overwrite existing admin password) |
| `SeedDemo` | When `SEED_DEMO=true` (or after `RESET_DB`): ~15 approved providers, many +/- reviews, demo residents |
| `ResetDB` | Truncates all app tables; then Seed + SeedDemo. **Dev only** |

Demo residents password (when seeded): historically `demo12345` — confirm in `seed_demo.go` if needed. Demo seed logs **counts only**, not passwords.

Fresh wipe workflow:

1. `RESET_DB=true`, `SEED_DEMO=true`
2. Restart API
3. Set `RESET_DB=false` again

---

## 9. Local run

```bash
cp .env.example .env
docker compose up -d
cd backend && go run ./cmd/server
# other terminal
cd frontend && npm install && npm run dev
```

- API: http://localhost:8080  
- UI: http://localhost:5173  
- Default invite: `CANTEGRIL2026`  
- Default admin (from `.env.example`): `admin@cantegril.local` / `admin123`  
  (local `.env` may differ — check file)

Go toolchain: **1.26+** (`go.mod` / Dockerfile aligned).

---

## 10. Decisions already made (don’t re-litigate unless asked)

1. Monorepo (not split FE/BE repos for MVP).
2. React **without** Next.js.
3. Email+password DIY auth (not magic link, not Auth0).
4. Public read / auth write; invite code soft gate.
5. Admin approval for providers **and** reviews.
6. No photos in v1.
7. Auth gate in **Go middleware**, not a sidecar.
8. Cookie sessions + DB, not JWT.
9. CSRF double-submit for credentialed cross-origin readiness.
10. Admin UI optimized for older users (one tab at a time, plain language).
11. Scores shown as stars + “X de 5” + word (Ruim…Excelente) + question hints.

---

## 11. Open / deferred (architecture)

- Pagination on providers/reviews
- Invite code single source of truth cleanup (env vs DB column)
- Multi-condo UI / host-based condo resolution
- Separate `cmd/seed` CLI vs boot seed
- Move `RESET_DB` / `SEED_DEMO` out of API binary into ops tool
- React Query (or similar) for FE data layer
- Approval audit trail — done (`reviewed_by`/`reviewed_at` + `audit_events`)
- Forgot-password / email verification
- ESLint + tests
- Production deploy wiring (Supabase + Render + GitHub Pages) — in progress

Historical review checklist lived in `REVIEW-TODO.md` (P0/P1 largely done as of 2026-07-21). Prefer this `CONTEXT.md` for ongoing orientation.

---

## 12. Conventions for agents working here

- Prefer **editing existing files**; don’t create docs unless asked (this file was requested).
- CSS: **no comments** in `.css` files (project preference).
- FE copy in **pt-BR**; keep older-user clarity in admin/scores.
- Never trust the SPA for authz; fix gates in Go.
- After auth/migration changes: remind user to **restart** `go run ./cmd/server`.
- Don’t commit secrets; `.env` is local.
- Commits/PRs only when the user asks.

---

## 13. Quick “where is X?”

| Concern | Location |
|---------|----------|
| Routes | `backend/internal/http/router.go` |
| Session cookie | `backend/internal/auth/auth.go`, handlers `auth.go` |
| CSRF | `backend/internal/http/middleware/csrf.go` + FE `api/client.ts` |
| Seed / demo / reset | `backend/internal/db/seed.go`, `seed_demo.go`, `reset.go` |
| Config / prod gates | `backend/internal/config/config.go` |
| Score UI | `frontend/src/components/ScoreDisplay.tsx`, `lib/format.ts` |
| Admin UX | `frontend/src/pages/AdminPage.tsx` |
| Categories / community name | `frontend/src/config.ts` |

---

*End of context dump. Prefer updating this file when major product or architecture decisions change.*
