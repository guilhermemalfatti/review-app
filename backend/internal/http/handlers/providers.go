package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gmalfatti/indica/backend/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProvidersHandler struct {
	pool    *pgxpool.Pool
	condoID uuid.UUID
}

func NewProvidersHandler(pool *pgxpool.Pool, condoID uuid.UUID) *ProvidersHandler {
	return &ProvidersHandler{pool: pool, condoID: condoID}
}

type Aggregates struct {
	HiredCount        int      `json:"hired_count"`
	RecommendCount    int      `json:"recommend_count"`
	NotRecommendCount int      `json:"not_recommend_count"`
	AvgPrice          *float64 `json:"avg_price"`
	AvgQuality        *float64 `json:"avg_quality"`
	AvgDeadline       *float64 `json:"avg_deadline"`
	AvgOverall        *float64 `json:"avg_overall"`
	LastServiceDate   *string  `json:"last_service_date"`
}

type ProviderListItem struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Category   string     `json:"category"`
	Phone      string     `json:"phone"`
	Notes      string     `json:"notes"`
	Aggregates Aggregates `json:"aggregates"`
}

type PublicReview struct {
	ID            uuid.UUID `json:"id"`
	AuthorLabel   string    `json:"author_label"`
	Recommend     bool      `json:"recommend"`
	ScorePrice    *int      `json:"score_price"`
	ScoreQuality  *int      `json:"score_quality"`
	ScoreDeadline *int      `json:"score_deadline"`
	Comment       string    `json:"comment"`
	ServiceDate   *string   `json:"service_date"`
	CreatedAt     time.Time `json:"created_at"`
}

type ProviderDetail struct {
	ID         uuid.UUID      `json:"id"`
	Name       string         `json:"name"`
	Category   string         `json:"category"`
	Phone      string         `json:"phone"`
	Notes      string         `json:"notes"`
	Aggregates Aggregates     `json:"aggregates"`
	Reviews    []PublicReview `json:"reviews"`
}

type createProviderRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Phone    string `json:"phone"`
	Notes    string `json:"notes"`
}

type createReviewRequest struct {
	IsAnonymous   bool   `json:"is_anonymous"`
	Recommend     bool   `json:"recommend"`
	ScorePrice    *int   `json:"score_price"`
	ScoreQuality  *int   `json:"score_quality"`
	ScoreDeadline *int   `json:"score_deadline"`
	Comment       string `json:"comment"`
	ServiceDate   *string `json:"service_date"`
}

const aggregatesSelect = `
	COUNT(r.id)::int AS hired_count,
	COUNT(r.id) FILTER (WHERE r.recommend)::int AS recommend_count,
	COUNT(r.id) FILTER (WHERE NOT r.recommend)::int AS not_recommend_count,
	AVG(r.score_price)::float8 AS avg_price,
	AVG(r.score_quality)::float8 AS avg_quality,
	AVG(r.score_deadline)::float8 AS avg_deadline,
	AVG(
		(COALESCE(r.score_price, 0) + COALESCE(r.score_quality, 0) + COALESCE(r.score_deadline, 0))::float8
		/ NULLIF(
			(CASE WHEN r.score_price IS NOT NULL THEN 1 ELSE 0 END) +
			(CASE WHEN r.score_quality IS NOT NULL THEN 1 ELSE 0 END) +
			(CASE WHEN r.score_deadline IS NOT NULL THEN 1 ELSE 0 END),
			0
		)
	)::float8 AS avg_overall,
	MAX(r.service_date)::text AS last_service_date
`

func scanAggregates(hired, recommend, notRecommend int, avgPrice, avgQuality, avgDeadline, avgOverall *float64, lastService *string) Aggregates {
	return Aggregates{
		HiredCount:        hired,
		RecommendCount:    recommend,
		NotRecommendCount: notRecommend,
		AvgPrice:          avgPrice,
		AvgQuality:        avgQuality,
		AvgDeadline:       avgDeadline,
		AvgOverall:        avgOverall,
		LastServiceDate:   lastService,
	}
}

func (h *ProvidersHandler) List(w http.ResponseWriter, r *http.Request) {
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	args := []any{h.condoID}
	where := `p.condo_id = $1 AND p.status = 'approved'`
	argN := 2

	if category != "" {
		where += ` AND p.category = $` + strconv.Itoa(argN)
		args = append(args, category)
		argN++
	}
	if q != "" {
		where += ` AND (p.name ILIKE $` + strconv.Itoa(argN) + ` OR p.notes ILIKE $` + strconv.Itoa(argN) + ` OR p.category ILIKE $` + strconv.Itoa(argN) + `)`
		args = append(args, "%"+q+"%")
		argN++
	}

	rows, err := h.pool.Query(r.Context(), `
		SELECT p.id, p.name, p.category, p.phone, p.notes,
			`+aggregatesSelect+`
		FROM providers p
		LEFT JOIN reviews r ON r.provider_id = p.id AND r.status = 'approved'
		WHERE `+where+`
		GROUP BY p.id
		ORDER BY p.name ASC
	`, args...)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list providers")
		return
	}
	defer rows.Close()

	items := make([]ProviderListItem, 0)
	for rows.Next() {
		var item ProviderListItem
		var avgPrice, avgQuality, avgDeadline, avgOverall *float64
		var lastService *string
		var hired, recommend, notRecommend int
		if err := rows.Scan(
			&item.ID, &item.Name, &item.Category, &item.Phone, &item.Notes,
			&hired, &recommend, &notRecommend,
			&avgPrice, &avgQuality, &avgDeadline, &avgOverall, &lastService,
		); err != nil {
			WriteError(w, http.StatusInternalServerError, "failed to scan provider")
			return
		}
		item.Aggregates = scanAggregates(hired, recommend, notRecommend, avgPrice, avgQuality, avgDeadline, avgOverall, lastService)
		if auth.UserFromContext(r.Context()) == nil {
			item.Phone = ""
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list providers")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *ProvidersHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid provider id")
		return
	}

	user := auth.UserFromContext(r.Context())
	isAdmin := user != nil && user.Role == "admin"

	var detail ProviderDetail
	var status string
	var avgPrice, avgQuality, avgDeadline, avgOverall *float64
	var lastService *string
	var hired, recommend, notRecommend int

	err = h.pool.QueryRow(r.Context(), `
		SELECT p.id, p.name, p.category, p.phone, p.notes, p.status,
			`+aggregatesSelect+`
		FROM providers p
		LEFT JOIN reviews r ON r.provider_id = p.id AND r.status = 'approved'
		WHERE p.id = $1 AND p.condo_id = $2
		GROUP BY p.id
	`, id, h.condoID).Scan(
		&detail.ID, &detail.Name, &detail.Category, &detail.Phone, &detail.Notes, &status,
		&hired, &recommend, &notRecommend,
		&avgPrice, &avgQuality, &avgDeadline, &avgOverall, &lastService,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "provider not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to get provider")
		return
	}

	if status != "approved" && !isAdmin {
		WriteError(w, http.StatusNotFound, "provider not found")
		return
	}

	detail.Aggregates = scanAggregates(hired, recommend, notRecommend, avgPrice, avgQuality, avgDeadline, avgOverall, lastService)
	if user == nil {
		detail.Phone = ""
	}
	detail.Reviews = []PublicReview{}

	rows, err := h.pool.Query(r.Context(), `
		SELECT r.id, r.is_anonymous, u.display_name, r.recommend,
			r.score_price, r.score_quality, r.score_deadline,
			r.comment, r.service_date::text, r.created_at
		FROM reviews r
		JOIN users u ON u.id = r.user_id
		WHERE r.provider_id = $1 AND r.status = 'approved'
		ORDER BY r.created_at DESC
	`, id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list reviews")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var rev PublicReview
		var isAnon bool
		var displayName string
		var serviceDate *string
		if err := rows.Scan(
			&rev.ID, &isAnon, &displayName, &rev.Recommend,
			&rev.ScorePrice, &rev.ScoreQuality, &rev.ScoreDeadline,
			&rev.Comment, &serviceDate, &rev.CreatedAt,
		); err != nil {
			WriteError(w, http.StatusInternalServerError, "failed to scan review")
			return
		}
		if isAnon {
			rev.AuthorLabel = "Anônimo"
		} else {
			rev.AuthorLabel = displayName
		}
		rev.ServiceDate = serviceDate
		detail.Reviews = append(detail.Reviews, rev)
	}
	if err := rows.Err(); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list reviews")
		return
	}

	WriteJSON(w, http.StatusOK, detail)
}

func (h *ProvidersHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createProviderRequest
	if err := DecodeJSON(w, r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Category = strings.TrimSpace(req.Category)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Notes = strings.TrimSpace(req.Notes)

	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Category == "" {
		WriteError(w, http.StatusBadRequest, "category is required")
		return
	}

	var id uuid.UUID
	var createdAt, updatedAt time.Time
	var status string
	err := h.pool.QueryRow(r.Context(), `
		INSERT INTO providers (condo_id, name, category, phone, notes, status, created_by)
		VALUES ($1, $2, $3, $4, $5, 'pending', $6)
		RETURNING id, status, created_at, updated_at
	`, h.condoID, req.Name, req.Category, req.Phone, req.Notes, user.ID).Scan(&id, &status, &createdAt, &updatedAt)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create provider")
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]any{
		"id":         id,
		"name":       req.Name,
		"category":   req.Category,
		"phone":      req.Phone,
		"notes":      req.Notes,
		"status":     status,
		"created_at": createdAt,
		"updated_at": updatedAt,
	})
}

func (h *ProvidersHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	providerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid provider id")
		return
	}

	var req createReviewRequest
	if err := DecodeJSON(w, r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if !validScore(req.ScorePrice) || !validScore(req.ScoreQuality) || !validScore(req.ScoreDeadline) {
		WriteError(w, http.StatusBadRequest, "scores must be between 1 and 5")
		return
	}

	var serviceDate *time.Time
	if req.ServiceDate != nil && strings.TrimSpace(*req.ServiceDate) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.ServiceDate))
		if err != nil {
			WriteError(w, http.StatusBadRequest, "service_date must be YYYY-MM-DD")
			return
		}
		serviceDate = &t
	}

	var providerStatus string
	err = h.pool.QueryRow(r.Context(), `
		SELECT status FROM providers WHERE id = $1 AND condo_id = $2
	`, providerID, h.condoID).Scan(&providerStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "provider not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to lookup provider")
		return
	}
	if providerStatus != "approved" {
		WriteError(w, http.StatusBadRequest, "provider must be approved to accept reviews")
		return
	}

	var existingID uuid.UUID
	err = h.pool.QueryRow(r.Context(), `
		SELECT id FROM reviews WHERE user_id = $1 AND provider_id = $2
	`, user.ID, providerID).Scan(&existingID)

	if errors.Is(err, pgx.ErrNoRows) {
		var id uuid.UUID
		var createdAt, updatedAt time.Time
		var status string
		err = h.pool.QueryRow(r.Context(), `
			INSERT INTO reviews (
				provider_id, user_id, is_anonymous, recommend,
				score_price, score_quality, score_deadline,
				comment, service_date, status
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'pending')
			RETURNING id, status, created_at, updated_at
		`, providerID, user.ID, req.IsAnonymous, req.Recommend,
			req.ScorePrice, req.ScoreQuality, req.ScoreDeadline,
			req.Comment, serviceDate,
		).Scan(&id, &status, &createdAt, &updatedAt)
		if err == nil {
			WriteJSON(w, http.StatusCreated, map[string]any{
				"id":             id,
				"provider_id":    providerID,
				"is_anonymous":   req.IsAnonymous,
				"recommend":      req.Recommend,
				"score_price":    req.ScorePrice,
				"score_quality":  req.ScoreQuality,
				"score_deadline": req.ScoreDeadline,
				"comment":        req.Comment,
				"service_date":   req.ServiceDate,
				"status":         status,
				"created_at":     createdAt,
				"updated_at":     updatedAt,
			})
			return
		}
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
			WriteError(w, http.StatusInternalServerError, "failed to create review")
			return
		}
		err = h.pool.QueryRow(r.Context(), `
			SELECT id FROM reviews WHERE user_id = $1 AND provider_id = $2
		`, user.ID, providerID).Scan(&existingID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "failed to lookup review after conflict")
			return
		}
	} else if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to lookup review")
		return
	}

	var updatedAt time.Time
	var status string
	err = h.pool.QueryRow(r.Context(), `
		UPDATE reviews SET
			is_anonymous = $1,
			recommend = $2,
			score_price = $3,
			score_quality = $4,
			score_deadline = $5,
			comment = $6,
			service_date = $7,
			status = 'pending',
			updated_at = now()
		WHERE id = $8
		RETURNING status, updated_at
	`, req.IsAnonymous, req.Recommend, req.ScorePrice, req.ScoreQuality, req.ScoreDeadline,
		req.Comment, serviceDate, existingID,
	).Scan(&status, &updatedAt)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update review")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"id":             existingID,
		"provider_id":    providerID,
		"is_anonymous":   req.IsAnonymous,
		"recommend":      req.Recommend,
		"score_price":    req.ScorePrice,
		"score_quality":  req.ScoreQuality,
		"score_deadline": req.ScoreDeadline,
		"comment":        req.Comment,
		"service_date":   req.ServiceDate,
		"status":         status,
		"updated_at":     updatedAt,
	})
}

type MyReview struct {
	ID            uuid.UUID `json:"id"`
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

func (h *ProvidersHandler) MyReview(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	providerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid provider id")
		return
	}

	var review MyReview
	err = h.pool.QueryRow(r.Context(), `
		SELECT r.id, r.is_anonymous, r.recommend, r.score_price, r.score_quality, r.score_deadline,
			r.comment, r.service_date::text, r.status, r.created_at, r.updated_at
		FROM reviews r
		JOIN providers p ON p.id = r.provider_id
		WHERE r.provider_id = $1 AND r.user_id = $2 AND p.condo_id = $3
	`, providerID, user.ID, h.condoID).Scan(
		&review.ID, &review.IsAnonymous, &review.Recommend, &review.ScorePrice, &review.ScoreQuality, &review.ScoreDeadline,
		&review.Comment, &review.ServiceDate, &review.Status, &review.CreatedAt, &review.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "review not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to lookup review")
		return
	}
	WriteJSON(w, http.StatusOK, review)
}
