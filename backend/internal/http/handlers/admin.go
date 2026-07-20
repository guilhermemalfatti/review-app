package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminHandler struct {
	pool    *pgxpool.Pool
	condoID uuid.UUID
}

func NewAdminHandler(pool *pgxpool.Pool, condoID uuid.UUID) *AdminHandler {
	return &AdminHandler{pool: pool, condoID: condoID}
}

type AdminProviderItem struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	Category           string    `json:"category"`
	Phone              string    `json:"phone"`
	Notes              string    `json:"notes"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	CreatorEmail       string    `json:"creator_email"`
	CreatorDisplayName string    `json:"creator_display_name"`
}

type AdminReviewItem struct {
	ID            uuid.UUID `json:"id"`
	ProviderID    uuid.UUID `json:"provider_id"`
	ProviderName  string    `json:"provider_name"`
	AuthorEmail   string    `json:"author_email"`
	AuthorDisplay string    `json:"author_display_name"`
	IsAnonymous   bool      `json:"is_anonymous"`
	Recommend     bool      `json:"recommend"`
	ScorePrice    *int      `json:"score_price"`
	ScoreQuality  *int      `json:"score_quality"`
	ScoreDeadline *int      `json:"score_deadline"`
	Comment       string    `json:"comment"`
	ServiceDate   *string   `json:"service_date"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (h *AdminHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status == "" {
		status = "pending"
	}
	if status != "pending" && status != "approved" && status != "rejected" {
		WriteError(w, http.StatusBadRequest, "invalid status")
		return
	}

	rows, err := h.pool.Query(r.Context(), `
		SELECT p.id, p.name, p.category, p.phone, p.notes, p.status, p.created_at, p.updated_at,
			u.email, u.display_name
		FROM providers p
		JOIN users u ON u.id = p.created_by
		WHERE p.condo_id = $1 AND p.status = $2
		ORDER BY p.created_at ASC
	`, h.condoID, status)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list providers")
		return
	}
	defer rows.Close()

	items := make([]AdminProviderItem, 0)
	for rows.Next() {
		var item AdminProviderItem
		if err := rows.Scan(
			&item.ID, &item.Name, &item.Category, &item.Phone, &item.Notes, &item.Status,
			&item.CreatedAt, &item.UpdatedAt, &item.CreatorEmail, &item.CreatorDisplayName,
		); err != nil {
			WriteError(w, http.StatusInternalServerError, "failed to scan provider")
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list providers")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *AdminHandler) setProviderStatus(w http.ResponseWriter, r *http.Request, status string) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid provider id")
		return
	}

	var item AdminProviderItem
	err = h.pool.QueryRow(r.Context(), `
		UPDATE providers p
		SET status = $1, updated_at = now()
		FROM users u
		WHERE p.id = $2 AND p.condo_id = $3 AND u.id = p.created_by
		RETURNING p.id, p.name, p.category, p.phone, p.notes, p.status, p.created_at, p.updated_at,
			u.email, u.display_name
	`, status, id, h.condoID).Scan(
		&item.ID, &item.Name, &item.Category, &item.Phone, &item.Notes, &item.Status,
		&item.CreatedAt, &item.UpdatedAt, &item.CreatorEmail, &item.CreatorDisplayName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "provider not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to update provider")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *AdminHandler) ApproveProvider(w http.ResponseWriter, r *http.Request) {
	h.setProviderStatus(w, r, "approved")
}

func (h *AdminHandler) RejectProvider(w http.ResponseWriter, r *http.Request) {
	h.setProviderStatus(w, r, "rejected")
}

func (h *AdminHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status == "" {
		status = "pending"
	}
	if status != "pending" && status != "approved" && status != "rejected" {
		WriteError(w, http.StatusBadRequest, "invalid status")
		return
	}

	rows, err := h.pool.Query(r.Context(), `
		SELECT r.id, r.provider_id, p.name, u.email, u.display_name,
			r.is_anonymous, r.recommend, r.score_price, r.score_quality, r.score_deadline,
			r.comment, r.service_date::text, r.status, r.created_at, r.updated_at
		FROM reviews r
		JOIN providers p ON p.id = r.provider_id
		JOIN users u ON u.id = r.user_id
		WHERE p.condo_id = $1 AND r.status = $2
		ORDER BY r.created_at ASC
	`, h.condoID, status)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list reviews")
		return
	}
	defer rows.Close()

	items := make([]AdminReviewItem, 0)
	for rows.Next() {
		var item AdminReviewItem
		if err := rows.Scan(
			&item.ID, &item.ProviderID, &item.ProviderName, &item.AuthorEmail, &item.AuthorDisplay,
			&item.IsAnonymous, &item.Recommend, &item.ScorePrice, &item.ScoreQuality, &item.ScoreDeadline,
			&item.Comment, &item.ServiceDate, &item.Status, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			WriteError(w, http.StatusInternalServerError, "failed to scan review")
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list reviews")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *AdminHandler) setReviewStatus(w http.ResponseWriter, r *http.Request, status string) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid review id")
		return
	}

	var item AdminReviewItem
	err = h.pool.QueryRow(r.Context(), `
		UPDATE reviews r
		SET status = $1, updated_at = now()
		FROM providers p, users u
		WHERE r.id = $2 AND p.id = r.provider_id AND p.condo_id = $3 AND u.id = r.user_id
		RETURNING r.id, r.provider_id, p.name, u.email, u.display_name,
			r.is_anonymous, r.recommend, r.score_price, r.score_quality, r.score_deadline,
			r.comment, r.service_date::text, r.status, r.created_at, r.updated_at
	`, status, id, h.condoID).Scan(
		&item.ID, &item.ProviderID, &item.ProviderName, &item.AuthorEmail, &item.AuthorDisplay,
		&item.IsAnonymous, &item.Recommend, &item.ScorePrice, &item.ScoreQuality, &item.ScoreDeadline,
		&item.Comment, &item.ServiceDate, &item.Status, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "review not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to update review")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *AdminHandler) ApproveReview(w http.ResponseWriter, r *http.Request) {
	h.setReviewStatus(w, r, "approved")
}

func (h *AdminHandler) RejectReview(w http.ResponseWriter, r *http.Request) {
	h.setReviewStatus(w, r, "rejected")
}
