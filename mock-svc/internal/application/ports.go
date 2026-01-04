package application

import (
	"context"
	"encoding/json"

	"mock-svc/internal/domain"

	"github.com/jmoiron/sqlx"
)

type TxManager interface {
	WithTx(ctx context.Context, fn func(*sqlx.Tx) error) error
}

type MockRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, mock domain.Mock) (domain.Mock, error)
	Get(ctx context.Context, id string) (domain.Mock, error)
	GetActive(ctx context.Context, id string) (domain.Mock, error)
	Update(ctx context.Context, tx *sqlx.Tx, id string, patch domain.MockPatch) (domain.Mock, error)
	UpdateFamily(ctx context.Context, tx *sqlx.Tx, id string, familyID *string) (domain.Mock, error)
	SoftDelete(ctx context.Context, tx *sqlx.Tx, id string) (domain.Mock, error)
	ListByIDs(ctx context.Context, ids []string) ([]domain.Mock, error)
	ListByFamily(ctx context.Context, familyID string) ([]domain.Mock, error)
}

type FamilyRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, primaryMockID string) (domain.Family, error)
	Get(ctx context.Context, id string) (domain.Family, error)
	UpdatePrimary(ctx context.Context, tx *sqlx.Tx, id, newPrimaryID string) (domain.Family, error)
}

type DerivedRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, rel domain.DerivedRelationship) error
	ListByDerivedIDs(ctx context.Context, derivedIDs []string) ([]domain.DerivedRelationship, error)
}

type GenerationRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, gen domain.Generation) (domain.Generation, error)
	Get(ctx context.Context, id string) (domain.Generation, error)
	UpdateStatus(ctx context.Context, tx *sqlx.Tx, id, status string, resultIDs []string, errMsg *string) error
}

type MockSearchRepository interface {
	Upsert(ctx context.Context, tx *sqlx.Tx, entry domain.MockSearch) error
	Update(ctx context.Context, tx *sqlx.Tx, entry domain.MockSearch) error
	Search(ctx context.Context, allowedIDs []string, filters domain.MockSearchFilters) ([]string, error)
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, tx *sqlx.Tx, eventType string, payload []byte, maxAttempts int) error
	ListDue(ctx context.Context, limit int) ([]domain.OutboxEvent, error)
	MarkDone(ctx context.Context, id string) error
	MarkAttempt(ctx context.Context, id string, nextAttemptSeconds int, lastError string) error
}

type Repositories struct {
	Mocks       MockRepository
	Families    FamilyRepository
	Derived     DerivedRepository
	Generations GenerationRepository
	MockSearch  MockSearchRepository
	Outbox      OutboxRepository
}

type AuthClient interface {
	CheckAccess(ctx context.Context, token, resourceType, resourceID, action string) (allowed bool, effectivePermission string, isOwner bool, err error)
	ListResources(ctx context.Context, token, resourceType, minPermission string) ([]ResourceAccess, error)
	RegisterResource(ctx context.Context, resourceType, resourceID, ownerUserID string) error
}

type DSLRunner interface {
	Preview(ctx context.Context, baseMock DSLBaseMock, dslScriptID string, dslScript string, params json.RawMessage) ([]GeneratedMock, error)
	Apply(ctx context.Context, baseMock DSLBaseMock, dslScriptID string, dslScript string, params json.RawMessage) ([]GeneratedMock, error)
}

type ResourceAccess struct {
	ResourceID          string
	EffectivePermission string
	IsOwner             bool
}

type DSLBaseMock struct {
	ID               string
	Name             string
	Description      *string
	RequestMatch     json.RawMessage
	ResponseTemplate json.RawMessage
	Enabled          bool
}

type GeneratedMock struct {
	Name             string          `json:"name"`
	Description      *string         `json:"description,omitempty"`
	RequestMatch     json.RawMessage `json:"requestMatch"`
	ResponseTemplate json.RawMessage `json:"responseTemplate"`
	Enabled          *bool           `json:"enabled,omitempty"`
}
