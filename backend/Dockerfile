# API-only image (no SPA). Prefer the repo-root Dockerfile for production
# so the Vite build is served from the same origin as /api (mobile cookies).
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates \
	&& adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=build /server /app/server
COPY migrations /app/migrations
RUN chown -R appuser:appuser /app
USER appuser
ENV MIGRATIONS_DIR=/app/migrations
EXPOSE 8080
CMD ["/app/server"]
