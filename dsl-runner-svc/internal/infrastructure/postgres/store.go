package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Store struct {
	DB      *sqlx.DB
	Jobs    *JobRepository
	Results *JobResultRepository
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		DB:      db,
		Jobs:    NewJobRepository(db),
		Results: NewJobResultRepository(db),
	}
}

func (s *Store) WithTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
