package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gmalfatti/indica/backend/internal/audit"
	"github.com/gmalfatti/indica/backend/internal/auth"
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
	ID                   uuid.UUID  `json:"id"`
	Name                 string     `json:"name"`
	Category             string     `json:"category"`
	Phone                string     `json:"phone"`
	Notes                string     `json:"notes"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	CreatorEmail         string     `json:"creator_email"`
	CreatorDisplayName   string     `json:"creator_display_name"`
	ReviewedBy           *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt           *time.Time `json:"reviewed_at,omitempty"`
	ReviewerEmail        *string    `json:"reviewer_email,omitempty"`
	ReviewerDisplayName  *string    `json:"reviewer_display_name,omitempty"`
}

type AdminReviewItem struct {
	ID                  uuid.UUID  `json:"id"`
	ProviderID          uuid.UUID  `json:"provider_id"`
	ProviderName        string     `json:"provider_name"`
	AuthorEmail         string     `json:"author_email"`
	AuthorDisplay       string     `json:"author_display_name"`
	IsAnonymous         bool       `json:"is_anonymous"`
	Recommend           bool       `json:"recommend"`
	ScorePrice          *int       `json:"score_price"`
	ScoreQuality        *int       `json:"score_quality"`
	ScoreDeadline       *int       `json:"score_deadline"`
	Comment             string     `json:"comment"`
	ServiceDate         *string    `json:"service_date"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	ReviewedBy          *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt          *time.Time `json:"reviewed_at,omitempty"`
	ReviewerEmail       *string    `json:"reviewer_email,omitempty"`
	ReviewerDisplayName *string    `json:"reviewer_display_name,omitempty"`
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
			u.email, u.display_name,
			p.reviewed_by, p.reviewed_at, rv.email, rv.display_name
		FROM providers p
		JOIN users u ON u.id = p.created_by
		LEFT JOIN users rv ON rv.id = p.reviewed_by
		WHERE p.condo_id = $1 AND p.status = $2
		ORDER BY p.created_at ASC
	`, h.condoID, status)
	if err != nil {
		WriteServerError(w, r, "failed to list providers", err)
		return
	}
	defer rows.Close()

	items := make([]AdminProviderItem, 0)
	for rows.Next() {
		var item AdminProviderItem
		if err := rows.Scan(
			&item.ID, &item.Name, &item.Category, &item.Phone, &item.Notes, &item.Status,
			&item.CreatedAt, &item.UpdatedAt, &item.CreatorEmail, &item.CreatorDisplayName,
			&item.ReviewedBy, &item.ReviewedAt, &item.ReviewerEmail, &item.ReviewerDisplayName,
		); err != nil {
			WriteServerError(w, r, "failed to scan provider", err)
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		WriteServerError(w, r, "failed to list providers", err)
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *AdminHandler) setProviderStatus(w http.ResponseWriter, r *http.Request, status string) {
	actor := auth.UserFromContext(r.Context())
	if actor == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid provider id")
		return
	}

	action := audit.ActionProviderApproved
	if status == "rejected" {
		action = audit.ActionProviderRejected
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		WriteServerError(w, r, "failed to begin transaction", err)
		return
	}
	defer tx.Rollback(r.Context())

	var previousStatus string
	err = tx.QueryRow(r.Context(), `
		SELECT status FROM providers WHERE id = $1 AND condo_id = $2
	`, id, h.condoID).Scan(&previousStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "provider not found")
			return
		}
		WriteServerError(w, r, "failed to lookup provider", err)
		return
	}

	var item AdminProviderItem
	err = tx.QueryRow(r.Context(), `
		UPDATE providers p
		SET status = $1, reviewed_by = $2, reviewed_at = now(), updated_at = now()
		FROM users u
		WHERE p.id = $3 AND p.condo_id = $4 AND u.id = p.created_by
		RETURNING p.id, p.name, p.category, p.phone, p.notes, p.status, p.created_at, p.updated_at,
			u.email, u.display_name, p.reviewed_by, p.reviewed_at
	`, status, actor.ID, id, h.condoID).Scan(
		&item.ID, &item.Name, &item.Category, &item.Phone, &item.Notes, &item.Status,
		&item.CreatedAt, &item.UpdatedAt, &item.CreatorEmail, &item.CreatorDisplayName,
		&item.ReviewedBy, &item.ReviewedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "provider not found")
			return
		}
		WriteServerError(w, r, "failed to update provider", err)
		return
	}

	item.ReviewerEmail = &actor.Email
	item.ReviewerDisplayName = &actor.DisplayName

	if err := audit.Insert(r.Context(), tx, audit.Event{
		CondoID:     h.condoID,
		ActorUserID: audit.Ptr(actor.ID),
		Action:      action,
		EntityType:  audit.EntityProvider,
		EntityID:    id,
		Payload: map[string]any{
			"from_status": previousStatus,
			"to_status":   status,
			"name":        item.Name,
		},
	}); err != nil {
		WriteServerError(w, r, "failed to record audit event", err)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		WriteServerError(w, r, "failed to commit", err)
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
	if status != "pending" && status != "approved" && status != "rejected" && status != "superseded" {
		WriteError(w, http.StatusBadRequest, "invalid status")
		return
	}

	rows, err := h.pool.Query(r.Context(), `
		SELECT r.id, r.provider_id, p.name, u.email, u.display_name,
			r.is_anonymous, r.recommend, r.score_price, r.score_quality, r.score_deadline,
			r.comment, r.service_date::text, r.status, r.created_at, r.updated_at,
			r.reviewed_by, r.reviewed_at, rv.email, rv.display_name
		FROM reviews r
		JOIN providers p ON p.id = r.provider_id
		JOIN users u ON u.id = r.user_id
		LEFT JOIN users rv ON rv.id = r.reviewed_by
		WHERE p.condo_id = $1 AND r.status = $2
		ORDER BY r.created_at ASC
	`, h.condoID, status)
	if err != nil {
		WriteServerError(w, r, "failed to list reviews", err)
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
			&item.ReviewedBy, &item.ReviewedAt, &item.ReviewerEmail, &item.ReviewerDisplayName,
		); err != nil {
			WriteServerError(w, r, "failed to scan review", err)
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		WriteServerError(w, r, "failed to list reviews", err)
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *AdminHandler) setReviewStatus(w http.ResponseWriter, r *http.Request, status string) {
	actor := auth.UserFromContext(r.Context())
	if actor == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid review id")
		return
	}

	action := audit.ActionReviewApproved
	if status == "rejected" {
		action = audit.ActionReviewRejected
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		WriteServerError(w, r, "failed to begin transaction", err)
		return
	}
	defer tx.Rollback(r.Context())

	var previousStatus string
	var userID, providerID uuid.UUID
	err = tx.QueryRow(r.Context(), `
		SELECT r.status, r.user_id, r.provider_id
		FROM reviews r
		JOIN providers p ON p.id = r.provider_id
		WHERE r.id = $1 AND p.condo_id = $2
	`, id, h.condoID).Scan(&previousStatus, &userID, &providerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "review not found")
			return
		}
		WriteServerError(w, r, "failed to lookup review", err)
		return
	}
	if previousStatus != "pending" {
		WriteError(w, http.StatusConflict, "only pending reviews can be moderated")
		return
	}

	if status == "approved" {
		// Keep history: mark the previously published version as superseded (not deleted).
		superRows, err := tx.Query(r.Context(), `
			UPDATE reviews
			SET status = 'superseded', updated_at = now()
			WHERE user_id = $1 AND provider_id = $2 AND status = 'approved' AND id <> $3
			RETURNING id
		`, userID, providerID, id)
		if err != nil {
			WriteServerError(w, r, "failed to supersede prior review", err)
			return
		}
		var supersededIDs []uuid.UUID
		for superRows.Next() {
			var oldID uuid.UUID
			if err := superRows.Scan(&oldID); err != nil {
				superRows.Close()
				WriteServerError(w, r, "failed to scan superseded review", err)
				return
			}
			supersededIDs = append(supersededIDs, oldID)
		}
		superRows.Close()
		if err := superRows.Err(); err != nil {
			WriteServerError(w, r, "failed to supersede prior review", err)
			return
		}
		for _, oldID := range supersededIDs {
			if err := audit.Insert(r.Context(), tx, audit.Event{
				CondoID:     h.condoID,
				ActorUserID: audit.Ptr(actor.ID),
				Action:      audit.ActionReviewSuperseded,
				EntityType:  audit.EntityReview,
				EntityID:    oldID,
				Payload: map[string]any{
					"provider_id":      providerID,
					"replaced_by":      id,
					"reason":           "newer_review_approved",
					"from_status":      "approved",
					"to_status":        "superseded",
				},
			}); err != nil {
				WriteServerError(w, r, "failed to record audit event", err)
				return
			}
		}
	}

	var item AdminReviewItem
	err = tx.QueryRow(r.Context(), `
		UPDATE reviews r
		SET status = $1, reviewed_by = $2, reviewed_at = now(), updated_at = now()
		FROM providers p, users u
		WHERE r.id = $3 AND p.id = r.provider_id AND p.condo_id = $4 AND u.id = r.user_id
		RETURNING r.id, r.provider_id, p.name, u.email, u.display_name,
			r.is_anonymous, r.recommend, r.score_price, r.score_quality, r.score_deadline,
			r.comment, r.service_date::text, r.status, r.created_at, r.updated_at,
			r.reviewed_by, r.reviewed_at
	`, status, actor.ID, id, h.condoID).Scan(
		&item.ID, &item.ProviderID, &item.ProviderName, &item.AuthorEmail, &item.AuthorDisplay,
		&item.IsAnonymous, &item.Recommend, &item.ScorePrice, &item.ScoreQuality, &item.ScoreDeadline,
		&item.Comment, &item.ServiceDate, &item.Status, &item.CreatedAt, &item.UpdatedAt,
		&item.ReviewedBy, &item.ReviewedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "review not found")
			return
		}
		WriteServerError(w, r, "failed to update review", err)
		return
	}

	item.ReviewerEmail = &actor.Email
	item.ReviewerDisplayName = &actor.DisplayName

	if err := audit.Insert(r.Context(), tx, audit.Event{
		CondoID:     h.condoID,
		ActorUserID: audit.Ptr(actor.ID),
		Action:      action,
		EntityType:  audit.EntityReview,
		EntityID:    id,
		Payload: map[string]any{
			"from_status":   previousStatus,
			"to_status":     status,
			"provider_id":   item.ProviderID,
			"provider_name": item.ProviderName,
		},
	}); err != nil {
		WriteServerError(w, r, "failed to record audit event", err)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		WriteServerError(w, r, "failed to commit", err)
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

type AdminUserItem struct {
	ID                 uuid.UUID `json:"id"`
	Email              string    `json:"email"`
	DisplayName        string    `json:"display_name"`
	Role               string    `json:"role"`
	MustChangePassword bool      `json:"must_change_password"`
	CreatedAt          time.Time `json:"created_at"`
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.pool.Query(r.Context(), `
		SELECT id, email, display_name, role, must_change_password, created_at
		FROM users
		WHERE condo_id = $1
		ORDER BY display_name ASC, email ASC
	`, h.condoID)
	if err != nil {
		WriteServerError(w, r, "failed to list users", err)
		return
	}
	defer rows.Close()

	items := make([]AdminUserItem, 0)
	for rows.Next() {
		var item AdminUserItem
		if err := rows.Scan(
			&item.ID, &item.Email, &item.DisplayName, &item.Role, &item.MustChangePassword, &item.CreatedAt,
		); err != nil {
			WriteServerError(w, r, "failed to scan user", err)
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		WriteServerError(w, r, "failed to list users", err)
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

type resetPasswordResponse struct {
	User              AdminUserItem `json:"user"`
	TemporaryPassword string        `json:"temporary_password"`
}

func (h *AdminHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	actor := auth.UserFromContext(r.Context())
	if actor == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	tempPassword, err := auth.GenerateTemporaryPassword()
	if err != nil {
		WriteServerError(w, r, "failed to generate password", err)
		return
	}
	hash, err := auth.HashPassword(tempPassword)
	if err != nil {
		WriteServerError(w, r, "failed to hash password", err)
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		WriteServerError(w, r, "failed to begin transaction", err)
		return
	}
	defer tx.Rollback(r.Context())

	var item AdminUserItem
	err = tx.QueryRow(r.Context(), `
		UPDATE users
		SET password_hash = $1, must_change_password = true
		WHERE id = $2 AND condo_id = $3
		RETURNING id, email, display_name, role, must_change_password, created_at
	`, hash, id, h.condoID).Scan(
		&item.ID, &item.Email, &item.DisplayName, &item.Role, &item.MustChangePassword, &item.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		WriteServerError(w, r, "failed to reset password", err)
		return
	}

	if _, err := tx.Exec(r.Context(), `DELETE FROM sessions WHERE user_id = $1`, id); err != nil {
		WriteServerError(w, r, "failed to revoke sessions", err)
		return
	}

	if err := audit.Insert(r.Context(), tx, audit.Event{
		CondoID:     h.condoID,
		ActorUserID: audit.Ptr(actor.ID),
		Action:      audit.ActionUserPasswordReset,
		EntityType:  audit.EntityUser,
		EntityID:    id,
		Payload: map[string]any{
			"target_email": item.Email,
		},
	}); err != nil {
		WriteServerError(w, r, "failed to record audit event", err)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		WriteServerError(w, r, "failed to commit password reset", err)
		return
	}

	WriteJSON(w, http.StatusOK, resetPasswordResponse{
		User:              item,
		TemporaryPassword: tempPassword,
	})
}
