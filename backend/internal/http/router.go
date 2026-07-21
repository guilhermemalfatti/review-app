package httpserver

import (
	"net/http"
	"time"

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
	authLimiter := middleware.NewIPRateLimiter(20, 15*time.Minute, handlers.WriteError)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger)
	r.Use(middleware.Recoverer(handlers.WriteError))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{d.CORSOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.CSRF(handlers.WriteError))

	r.Get("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		handlers.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/auth", func(r chi.Router) {
		r.Get("/csrf", middleware.CSRFTokenHandler(d.CookieSecure, handlers.WriteError))
		r.With(authLimiter.Middleware).Post("/signup", authH.Signup)
		r.With(authLimiter.Middleware).Post("/login", authH.Login)
		r.Post("/logout", authH.Logout)
		r.Get("/me", authH.Me)
		r.With(middleware.RequireAuth(d.Sessions, handlers.WriteError)).Post("/change-password", authH.ChangePassword)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.OptionalAuth(d.Sessions, handlers.WriteError))
		r.Get("/api/providers", providersH.List)
		r.Get("/api/providers/{id}", providersH.Get)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(d.Sessions, handlers.WriteError))
		r.Use(middleware.RequirePasswordChanged(handlers.WriteError))
		r.Post("/api/providers", providersH.Create)
		r.Get("/api/providers/{id}/my-review", providersH.MyReview)
		r.Post("/api/providers/{id}/reviews", providersH.CreateReview)
	})

	r.Route("/api/admin", func(r chi.Router) {
		r.Use(middleware.RequireAuth(d.Sessions, handlers.WriteError))
		r.Use(middleware.RequireAdmin(handlers.WriteError))
		r.Use(middleware.RequirePasswordChanged(handlers.WriteError))
		r.Get("/providers", adminH.ListProviders)
		r.Post("/providers/{id}/approve", adminH.ApproveProvider)
		r.Post("/providers/{id}/reject", adminH.RejectProvider)
		r.Get("/reviews", adminH.ListReviews)
		r.Post("/reviews/{id}/approve", adminH.ApproveReview)
		r.Post("/reviews/{id}/reject", adminH.RejectReview)
		r.Get("/users", adminH.ListUsers)
		r.Post("/users/{id}/reset-password", adminH.ResetPassword)
	})

	return r
}
