# Indica — Cantegril trusted providers

Residents of condo **Cantegril** share trusted service-provider indications. This monorepo has:

- `/backend` — Go API
- `/frontend` — Vite/React SPA
- Root `Dockerfile` — production image (SPA + API, same origin)

**Production:** Google Cloud Run (`southamerica-east1`) + optional Firebase Hosting for the custom domain `indica-cantegril.com.br`.

## Prerequisites

- Go 1.26+
- Node 22+ (frontend)
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
| `COOKIE_SAMESITE` | `Lax` |
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

## 5. Production deploy (Cloud Run)

Preferred: **one Cloud Run service** serves API + SPA (same origin → first-party cookies on mobile).

| Piece | Where |
|---|---|
| Postgres | Supabase (Session pooler URI + `sslmode=require`) |
| App (API + SPA) | Cloud Run — Docker from **repo-root** `Dockerfile` |
| Custom domain (optional) | Firebase Hosting rewrite → Cloud Run |

### Cloud Run

- Region: `southamerica-east1` (or another region you choose)
- Build: continuous deploy from GitHub using root `Dockerfile`, or:
  ```bash
  gcloud run deploy SERVICE_NAME \
    --source . \
    --region southamerica-east1 \
    --allow-unauthenticated
  ```
- Container listens on `PORT` (Cloud Run sets this automatically)

**Cloud Run env (minimum):**

| Variable | Value |
|---|---|
| `APP_ENV` | `production` |
| `COOKIE_SECURE` | `true` |
| `COOKIE_SAMESITE` | `Lax` (same-origin; default) |
| `CORS_ORIGIN` | `https://YOUR-SERVICE-xxxxx.REGION.run.app` (no trailing slash), or your custom domain once live |
| `DATABASE_URL` | Supabase Session pooler URI |
| `INVITE_CODE` / `ADMIN_PASSWORD` | strong non-default values |
| `SEED_DEMO` | `false` |

`STATIC_DIR=/app/static` and `MIGRATIONS_DIR=/app/migrations` are set in the image.

Open `https://YOUR-SERVICE-….run.app/` for the app and `/api/health` for the API.

### Custom domain (Firebase Hosting)

Cloud Run **domain mapping is not available** in `southamerica-east1`. Use Firebase Hosting as a reverse proxy:

1. Add Firebase to the **same** GCP project as Cloud Run
2. Configure `firebase.json` (already in repo) — rewrite `**` → Cloud Run service `indica-02` in `southamerica-east1`
3. Deploy Hosting:
   ```bash
   npm i -g firebase-tools
   firebase login
   firebase use YOUR_GCP_PROJECT_ID
   firebase deploy --only hosting
   ```
4. Firebase Console → Hosting → **Add custom domain** (e.g. `indica-cantegril.com.br`)
5. Add the DNS records Firebase shows at your registrar (e.g. Registro.br)
6. When the domain is **Connected**, set Cloud Run:
   ```text
   CORS_ORIGIN=https://indica-cantegril.com.br
   ```

App deploys stay on Cloud Run (GitHub / `gcloud`). `firebase deploy --only hosting` only updates the Hosting proxy/domain config.

### Local Docker smoke test

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
| GET | `/api/config` | public (e.g. condominio fases) |
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
| POST | `/api/admin/providers/:id/remove` | admin |
| GET | `/api/admin/reviews` | admin |
| POST | `/api/admin/reviews/:id/approve` | admin |
| POST | `/api/admin/reviews/:id/reject` | admin |
| GET | `/api/admin/users` | admin |
| POST | `/api/admin/users/:id/reset-password` | admin |

Session cookie name: `session` (HTTP-only). CSRF cookie name: `csrf` (readable by JS).
