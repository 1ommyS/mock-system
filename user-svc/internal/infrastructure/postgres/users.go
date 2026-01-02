package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash, status string) (User, error) {
	const q = `
		INSERT INTO users (email, password_hash, status)
		VALUES ($1, $2, $3)
		RETURNING id, email, password_hash, status, created_at, updated_at, last_login_at
	`
	var user User
	if err := r.db.GetContext(ctx, &user, q, strings.ToLower(email), passwordHash, status); err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	const q = `
		SELECT id, email, password_hash, status, created_at, updated_at, last_login_at
		FROM users
		WHERE lower(email) = lower($1)
	`
	var user User
	if err := r.db.GetContext(ctx, &user, q, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (User, error) {
	const q = `
		SELECT id, email, password_hash, status, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`
	var user User
	if err := r.db.GetContext(ctx, &user, q, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return user, nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, userID, status string) (User, error) {
	const q = `
		UPDATE users
		SET status = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, email, password_hash, status, created_at, updated_at, last_login_at
	`
	var user User
	if err := r.db.GetContext(ctx, &user, q, userID, status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return user, nil
}

func (r *UserRepository) UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error {
	const q = `
		UPDATE users
		SET password_hash = $2, updated_at = now()
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, q, userID, passwordHash)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	const q = `
		UPDATE users
		SET last_login_at = now(), updated_at = now()
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, q, userID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}
