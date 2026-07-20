# Indica — Cantegril trusted providers

Residents of condo **Cantegril** share trusted service-provider indications. This monorepo has:

- `/backend` — Go API (this doc)
- `/frontend` — Vite/React app (expects the API on `http://localhost:8080`)

## Prerequisites

- Go 1.22+
- Docker + Docker Compose

## 1. Environment

```bash
cp .env.example .env
```

Defaults:

| Variable | Default |
|---|---|
| `DATABASE_URL` | `postgres://indica:indica@localhost:5432/indica?sslmode=disable` |
| `SESSION_SECRET` | `dev-secret-change-me` |
| `INVITE_CODE` | `CANTEGRIL2026` |
| `CORS_ORIGIN` | `http://localhost:5173` |
| `PORT` | `8080` |
| `ADMIN_EMAIL` | `admin@cantegril.local` |
| `ADMIN_PASSWORD` | `admin123` |
| `ADMIN_DISPLAY_NAME` | `Admin` |

## 2. Start Postgres

From the repo root:

```bash
docker compose up -d
```

Wait until healthy (`docker compose ps`).

## 3. Run the API

Migrations and seed (Cantegril condo + admin user) run automatically on startup.

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

## Default credentials

- **Invite code (signup):** `CANTEGRIL2026`
- **Admin:** `admin@cantegril.local` / `admin123`

## API overview

| Method | Path | Auth |
|---|---|---|
| GET | `/api/health` | public |
| POST | `/api/auth/signup` | public (invite code) |
| POST | `/api/auth/login` | public |
| POST | `/api/auth/logout` | optional |
| GET | `/api/auth/me` | session |
| GET | `/api/providers` | public (approved only) |
| GET | `/api/providers/:id` | public (approved; admin can see others) |
| POST | `/api/providers` | session |
| POST | `/api/providers/:id/reviews` | session |
| GET | `/api/admin/providers` | admin |
| POST | `/api/admin/providers/:id/approve` | admin |
| POST | `/api/admin/providers/:id/reject` | admin |
| GET | `/api/admin/reviews` | admin |
| POST | `/api/admin/reviews/:id/approve` | admin |
| POST | `/api/admin/reviews/:id/reject` | admin |

Session cookie name: `session` (HTTP-only).
