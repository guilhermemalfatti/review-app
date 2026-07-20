package httpserver

import (
	"net/http"

	"github.com/gmalfatti/indica/backend/internal/auth"
	"github.com/gmalfatti/indica/backend/internal/http/handlers"
	"github.com/gmalfatti/indica/backend/internal/http/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	Pool         *pgxpool.Pool
	Sessions     *auth.SessionStore
	CondoID      uuid.UUID
	InviteCode   string
	CORSOrigin   string
	CookieSecure bool
}

func NewRouter(d Deps) http.Handler {
	authH := handlers.NewAuthHandler(d.Pool, d.Sessions, d.CondoID, d.InviteCode, d.CookieSecure)
	providersH := handlers.NewProvidersHandler(d.Pool, d.CondoID)
	adminH := handlers.NewAdminHandler(d.Pool, d.CondoID)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{d.CORSOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		handlers.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/signup", authH.Signup)
		r.Post("/login", authH.Login)
		r.Post("/logout", authH.Logout)
		r.Get("/me", authH.Me)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.OptionalAuth(d.Sessions))
		r.Get("/api/providers", providersH.List)
		r.Get("/api/providers/{id}", providersH.Get)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(d.Sessions, handlers.WriteError))
		r.Post("/api/providers", providersH.Create)
		r.Post("/api/providers/{id}/reviews", providersH.CreateReview)
	})

	r.Route("/api/admin", func(r chi.Router) {
		r.Use(middleware.RequireAuth(d.Sessions, handlers.WriteError))
		r.Use(middleware.RequireAdmin(handlers.WriteError))
		r.Get("/providers", adminH.ListProviders)
		r.Post("/providers/{id}/approve", adminH.ApproveProvider)
		r.Post("/providers/{id}/reject", adminH.RejectProvider)
		r.Get("/reviews", adminH.ListReviews)
		r.Post("/reviews/{id}/approve", adminH.ApproveReview)
		r.Post("/reviews/{id}/reject", adminH.RejectReview)
	})

	return r
}
