package postgres

import "github.com/jmoiron/sqlx"

type Store struct {
	DB        *sqlx.DB
	Users     *UserRepository
	Roles     *RoleRepository
	Sessions  *SessionRepository
	Resources *ResourceRepository
	Grants    *GrantRepository
	Audit     *AuditRepository
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		DB:        db,
		Users:     NewUserRepository(db),
		Roles:     NewRoleRepository(db),
		Sessions:  NewSessionRepository(db),
		Resources: NewResourceRepository(db),
		Grants:    NewGrantRepository(db),
		Audit:     NewAuditRepository(db),
	}
}
