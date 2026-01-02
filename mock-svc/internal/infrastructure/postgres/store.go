package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Store struct {
	DB          *sqlx.DB
	Mocks       *MockRepository
	Families    *FamilyRepository
	Derived     *DerivedRepository
	Generations *GenerationRepository
	MockSearch  *MockSearchRepository
	Outbox      *OutboxRepository
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		DB:          db,
		Mocks:       NewMockRepository(db),
		Families:    NewFamilyRepository(db),
		Derived:     NewDerivedRepository(db),
		Generations: NewGenerationRepository(db),
		MockSearch:  NewMockSearchRepository(db),
		Outbox:      NewOutboxRepository(db),
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
