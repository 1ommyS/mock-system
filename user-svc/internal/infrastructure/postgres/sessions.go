package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type SessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, userID, refreshHash string, ip, userAgent *string) (Session, error) {
	const q = `
		INSERT INTO sessions (user_id, refresh_hash, ip, user_agent)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, refresh_hash, created_at, last_used_at, revoked_at, ip, user_agent
	`
	var session Session
	if err := r.db.GetContext(ctx, &session, q, userID, refreshHash, ip, userAgent); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (r *SessionRepository) GetByRefreshHash(ctx context.Context, refreshHash string) (Session, error) {
	const q = `
		SELECT id, user_id, refresh_hash, created_at, last_used_at, revoked_at, ip, user_agent
		FROM sessions
		WHERE refresh_hash = $1
	`
	var session Session
	if err := r.db.GetContext(ctx, &session, q, refreshHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrNotFound
		}
		return Session{}, err
	}
	return session, nil
}

func (r *SessionRepository) GetByID(ctx context.Context, sessionID string) (Session, error) {
	const q = `
		SELECT id, user_id, refresh_hash, created_at, last_used_at, revoked_at, ip, user_agent
		FROM sessions
		WHERE id = $1
	`
	var session Session
	if err := r.db.GetContext(ctx, &session, q, sessionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrNotFound
		}
		return Session{}, err
	}
	return session, nil
}

func (r *SessionRepository) UpdateLastUsed(ctx context.Context, sessionID string) error {
	const q = `UPDATE sessions SET last_used_at = now() WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, sessionID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SessionRepository) RotateRefresh(ctx context.Context, sessionID, refreshHash string) error {
	const q = `
		UPDATE sessions
		SET refresh_hash = $2, last_used_at = now()
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, q, sessionID, refreshHash)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SessionRepository) RevokeByID(ctx context.Context, sessionID string) error {
	const q = `UPDATE sessions SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, sessionID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SessionRepository) RevokeByRefreshHash(ctx context.Context, refreshHash string) error {
	const q = `UPDATE sessions SET revoked_at = now() WHERE refresh_hash = $1 AND revoked_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, refreshHash)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}
