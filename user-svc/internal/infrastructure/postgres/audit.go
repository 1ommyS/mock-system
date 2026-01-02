package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type AuditRepository struct {
	db *sqlx.DB
}

func NewAuditRepository(db *sqlx.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, entry AuditLog) (AuditLog, error) {
	const q = `
		INSERT INTO audit_log (
			actor_user_id,
			event_type,
			resource_type,
			resource_id,
			ip,
			user_agent,
			metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, actor_user_id, event_type, resource_type, resource_id, ip, user_agent, metadata, created_at
	`
	var out AuditLog
	if err := r.db.GetContext(ctx, &out, q,
		entry.ActorUserID,
		entry.EventType,
		entry.ResourceType,
		entry.ResourceID,
		entry.IP,
		entry.UserAgent,
		entry.Metadata,
	); err != nil {
		return AuditLog{}, err
	}
	return out, nil
}
