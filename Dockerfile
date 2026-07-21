# Production image: Vite SPA + Go API (same origin → first-party cookies on mobile).
# Render: Root Directory = repo root, Dockerfile Path = ./Dockerfile

FROM node:22-alpine AS frontend
WORKDIR /fe
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
# Same-origin: relative /api (do not set VITE_API_URL). Base path is /.
ENV VITE_API_URL=
ENV VITE_BASE_PATH=/
RUN npm run build

FROM golang:1.26-alpine AS backend
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates \
	&& adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=backend /server /app/server
COPY --from=backend /src/migrations /app/migrations
COPY --from=frontend /fe/dist /app/static
RUN chown -R appuser:appuser /app
USER appuser
ENV MIGRATIONS_DIR=/app/migrations
ENV STATIC_DIR=/app/static
EXPOSE 8080
CMD ["/app/server"]
