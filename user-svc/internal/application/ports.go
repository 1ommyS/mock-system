package application

import (
	"context"

	"user-svc/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash, status string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, userID string) (domain.User, error)
	UpdateStatus(ctx context.Context, userID, status string) (domain.User, error)
	UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error
	UpdateLastLogin(ctx context.Context, userID string) error
}

type RoleRepository interface {
	ReplaceUserRoles(ctx context.Context, userID string, roles []string) error
	ListByUserID(ctx context.Context, userID string) ([]string, error)
}

type SessionRepository interface {
	Create(ctx context.Context, userID, refreshHash string, ip, userAgent *string) (domain.Session, error)
	GetByRefreshHash(ctx context.Context, refreshHash string) (domain.Session, error)
	GetByID(ctx context.Context, sessionID string) (domain.Session, error)
	UpdateLastUsed(ctx context.Context, sessionID string) error
	RotateRefresh(ctx context.Context, sessionID, refreshHash string) error
	RevokeByID(ctx context.Context, sessionID string) error
	RevokeByRefreshHash(ctx context.Context, refreshHash string) error
}

type ResourceRepository interface {
	Register(ctx context.Context, resourceType, resourceID, ownerUserID string) (domain.Resource, bool, error)
	Get(ctx context.Context, resourceType, resourceID string) (domain.Resource, error)
	UpdateOwner(ctx context.Context, resourceType, resourceID, ownerUserID string) (domain.Resource, error)
	GetAccess(ctx context.Context, userID, resourceType, resourceID string) (domain.AccessCheck, error)
	ListByUser(ctx context.Context, userID, resourceType, minPermission string) ([]domain.ResourceAccess, error)
}

type GrantRepository interface {
	Create(ctx context.Context, resourceType, resourceID, userID, permission string) (domain.Grant, error)
	Delete(ctx context.Context, resourceType, resourceID, userID string) error
	ListByResource(ctx context.Context, resourceType, resourceID string) ([]domain.Grant, error)
	GetByUser(ctx context.Context, resourceType, resourceID, userID string) (domain.Grant, error)
}

type AuditRepository interface {
	Create(ctx context.Context, entry domain.AuditLog) (domain.AuditLog, error)
}

type Repositories struct {
	Users     UserRepository
	Roles     RoleRepository
	Sessions  SessionRepository
	Resources ResourceRepository
	Grants    GrantRepository
	Audit     AuditRepository
}
