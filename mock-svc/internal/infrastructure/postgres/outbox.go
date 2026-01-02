package postgres

import (
	"context"

	"mock-svc/internal/domain"

	"github.com/jmoiron/sqlx"
)

type OutboxRepository struct {
	db *sqlx.DB
}

func NewOutboxRepository(db *sqlx.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) Enqueue(ctx context.Context, tx *sqlx.Tx, eventType string, payload []byte, maxAttempts int) error {
	const q = `
		INSERT INTO outbox (event_type, payload, max_attempts)
		VALUES ($1, $2, $3)
	`
	_, err := tx.ExecContext(ctx, q, eventType, payload, maxAttempts)
	return err
}

func (r *OutboxRepository) ListDue(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `
		SELECT id, event_type, payload, attempts, max_attempts, last_error, next_attempt_at, created_at, updated_at
		FROM outbox
		WHERE next_attempt_at <= now() AND attempts < max_attempts
		ORDER BY next_attempt_at ASC
		LIMIT $1
	`
	var items []domain.OutboxEvent
	if err := r.db.SelectContext(ctx, &items, q, limit); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *OutboxRepository) MarkDone(ctx context.Context, id string) error {
	const q = `
		DELETE FROM outbox WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, q, id)
	return err
}

func (r *OutboxRepository) MarkAttempt(ctx context.Context, id string, nextAttemptSeconds int, lastError string) error {
	const q = `
		UPDATE outbox
		SET attempts = attempts + 1,
			last_error = $2,
			next_attempt_at = now() + ($3 || ' seconds')::interval,
			updated_at = now()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, q, id, lastError, nextAttemptSeconds)
	return err
}
