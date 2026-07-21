# Indica — Cantegril trusted providers

Residents of condo **Cantegril** share trusted service-provider indications. This monorepo has:

- `/backend` — Go API (this doc)
- `/frontend` — Vite/React app (expects the API on `http://localhost:8080`)

## Prerequisites

- Go 1.26+
- Docker + Docker Compose

## 1. Environment

```bash
cp .env.example .env
```

Defaults:

| Variable | Default |
|---|---|
| `DATABASE_URL` | `postgres://indica:indica@localhost:5432/indica?sslmode=disable` |
| `APP_ENV` | `development` (`production` enables fail-closed checks) |
| `COOKIE_SECURE` | unset → `true` when `APP_ENV=production`, else `false` |
| `INVITE_CODE` | `CANTEGRIL2026` |
| `CORS_ORIGIN` | `http://localhost:5173` |
| `PORT` | `8080` |
| `ADMIN_EMAIL` | `admin@cantegril.local` |
| `ADMIN_PASSWORD` | `admin123` |
| `ADMIN_DISPLAY_NAME` | `Admin` |
| `SEED_DEMO` | `false` (`true` loads sample providers + reviews) |
| `RESET_DB` | `false` (`true` wipes all data on startup, then runs seeders; refused in production) |

In production (`APP_ENV=production`), the API refuses weak defaults (`admin123`, `CANTEGRIL2026`), requires `COOKIE_SECURE=true`, and blocks `RESET_DB`.

## 2. Start Postgres

From the repo root:

```bash
docker compose up -d
```

Wait until healthy (`docker compose ps`).

## 3. Run the API

Migrations and seed (Cantegril condo + admin user) run automatically on startup. Seed is create-once: existing condo invite codes and admin passwords are not overwritten on restart.

```bash
cd backend
go run ./cmd/server
```

Or build first:

```bash
cd backend
go build -o bin/server ./cmd/server
./bin/server
```

The server loads `.env` from the current working directory or the parent (`../.env`).

Health check: [http://localhost:8080/api/health](http://localhost:8080/api/health) → `{"status":"ok"}`

## 4. Frontend

With the API running on `:8080`:

```bash
cd frontend
npm install
npm run dev
```

Open [http://localhost:5173](http://localhost:5173). Vite proxies `/api` to the Go server (cookie sessions).

Optional frontend env (see `frontend/.env.example`): `VITE_API_URL` (empty locally), `VITE_BASE_PATH` (default `/`).

## 5. Production deploy (pilot)

Preferred: **one Render Web Service** serves API + SPA (same origin → cookies work on iOS).

| Piece | Where |
|---|---|
| Postgres | Supabase (Session pooler URI + `sslmode=require`) |
| App (API + SPA) | Render Web Service — Docker from **repo-root** `Dockerfile` |

**Render settings**

- Root Directory: *(empty / repo root)*
- Dockerfile Path: `./Dockerfile`
- Docker build context: repo root

**Render env (minimum):**

| Variable | Value |
|---|---|
| `APP_ENV` | `production` |
| `COOKIE_SECURE` | `true` |
| `COOKIE_SAMESITE` | `Lax` (same-origin; default) |
| `CORS_ORIGIN` | `https://YOUR-SERVICE.onrender.com` (no trailing slash) |
| `DATABASE_URL` | Supabase Session pooler URI |
| `INVITE_CODE` / `ADMIN_PASSWORD` | strong non-default values |
| `SEED_DEMO` | `false` |

Open `https://YOUR-SERVICE.onrender.com/` for the app and `/api/health` for the API. `STATIC_DIR=/app/static` is set in the image.

Optional legacy: GitHub Pages + API-only `backend/Dockerfile` requires `COOKIE_SAMESITE=None` and `CORS_ORIGIN=https://guilhermemalfatti.github.io` (mobile Safari often still blocks those cookies).

Local Docker smoke test from repo root:

```bash
docker build -t indica .
docker run --rm -p 8080:8080 --env-file .env -e STATIC_DIR=/app/static indica
```


## CSRF

Mutating requests (`POST` / `PUT` / `PATCH` / `DELETE` under `/api/*`) require a CSRF token:

1. `GET /api/auth/csrf` → sets a non-HttpOnly `csrf` cookie and returns `{"csrf_token":"..."}`
2. Send the same value in the `X-CSRF-Token` header on mutating requests (with `credentials: 'include'`)

## Default credentials

- **Invite code (signup):** `CANTEGRIL2026`
- **Admin:** `admin@cantegril.local` / `admin123`

Admins can list users and reset passwords in the Admin UI. A reset issues a temporary password, revokes sessions, and forces a password change on next login (`POST /api/auth/change-password`).

### Fresh demo data

1. Set `RESET_DB=true` and `SEED_DEMO=true` in `.env`
2. Restart the API (`go run ./cmd/server`)
3. Set `RESET_DB=false` again so the next restart does not wipe data

`RESET_DB` truncates all tables, then runs the condo/admin seeder and the demo seeder (15 providers, many positive and negative reviews). Only allowed when `APP_ENV` is not `production`.

## API overview

| Method | Path | Auth |
|---|---|---|
| GET | `/api/health` | public |
| GET | `/api/auth/csrf` | public (issues CSRF cookie + token) |
| POST | `/api/auth/signup` | public (invite code) |
| POST | `/api/auth/login` | public |
| POST | `/api/auth/logout` | optional |
| GET | `/api/auth/me` | session |
| POST | `/api/auth/change-password` | session |
| GET | `/api/providers` | public (approved only; phone hidden unless logged in) |
| GET | `/api/providers/:id` | public (approved; admin can see others; phone hidden unless logged in) |
| POST | `/api/providers` | session |
| POST | `/api/providers/:id/reviews` | session |
| GET | `/api/admin/providers` | admin |
| POST | `/api/admin/providers/:id/approve` | admin |
| POST | `/api/admin/providers/:id/reject` | admin |
| GET | `/api/admin/reviews` | admin |
| POST | `/api/admin/reviews/:id/approve` | admin |
| POST | `/api/admin/reviews/:id/reject` | admin |
| GET | `/api/admin/users` | admin |
| POST | `/api/admin/users/:id/reset-password` | admin |

Session cookie name: `session` (HTTP-only). CSRF cookie name: `csrf` (readable by JS).
