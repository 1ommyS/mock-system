package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(db *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) EnsureRoles(ctx context.Context, roles []string) error {
	return ensureRoles(ctx, r.db, roles)
}

func (r *RoleRepository) ReplaceUserRoles(ctx context.Context, userID string, roles []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = ensureRoles(ctx, tx, roles); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = $1`, userID); err != nil {
		return err
	}

	if len(roles) > 0 {
		query := `INSERT INTO user_roles (user_id, role_name) VALUES `
		args := make([]any, 0, len(roles)*2)
		for i, role := range roles {
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
			args = append(args, userID, role)
		}
		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *RoleRepository) ListByUserID(ctx context.Context, userID string) ([]string, error) {
	var roles []string
	if err := r.db.SelectContext(ctx, &roles, `SELECT role_name FROM user_roles WHERE user_id = $1`, userID); err != nil {
		return nil, err
	}
	return roles, nil
}

func ensureRoles(ctx context.Context, exec sqlx.ExtContext, roles []string) error {
	if len(roles) == 0 {
		return nil
	}
	for _, role := range roles {
		if _, err := exec.ExecContext(ctx, `INSERT INTO roles (name) VALUES ($1) ON CONFLICT DO NOTHING`, role); err != nil {
			return err
		}
	}
	return nil
}
