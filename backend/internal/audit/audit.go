package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	ActionProviderCreated   = "provider.created"
	ActionProviderApproved  = "provider.approved"
	ActionProviderRejected  = "provider.rejected"
	ActionReviewCreated     = "review.created"
	ActionReviewApproved    = "review.approved"
	ActionReviewRejected    = "review.rejected"
	ActionReviewSuperseded  = "review.superseded"
	ActionUserPasswordReset = "user.password_reset"

	EntityProvider = "provider"
	EntityReview   = "review"
	EntityUser     = "user"
)

type Execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type Event struct {
	CondoID     uuid.UUID
	ActorUserID *uuid.UUID
	Action      string
	EntityType  string
	EntityID    uuid.UUID
	Payload     map[string]any
}

func Insert(ctx context.Context, db Execer, ev Event) error {
	payload := ev.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal audit payload: %w", err)
	}
	_, err = db.Exec(ctx, `
		INSERT INTO audit_events (condo_id, actor_user_id, action, entity_type, entity_id, payload)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
	`, ev.CondoID, ev.ActorUserID, ev.Action, ev.EntityType, ev.EntityID, string(raw))
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

func Ptr(id uuid.UUID) *uuid.UUID {
	return &id
}
