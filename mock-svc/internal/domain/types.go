package domain

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

type Mock struct {
	ID               string          `db:"id"`
	Name             string          `db:"name"`
	Description      *string         `db:"description"`
	RequestMatch     json.RawMessage `db:"request_match"`
	ResponseTemplate json.RawMessage `db:"response_template"`
	Enabled          bool            `db:"enabled"`
	FamilyID         *string         `db:"family_id"`
	CreatedBy        string          `db:"created_by"`
	GeneratedBy      *string         `db:"generated_by"`
	CreatedAt        time.Time       `db:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at"`
	DeletedAt        *time.Time      `db:"deleted_at"`
}

type MockPatch struct {
	Name             *string
	Description      **string
	RequestMatch     *json.RawMessage
	ResponseTemplate *json.RawMessage
	Enabled          *bool
}

type Family struct {
	ID            string `db:"id"`
	PrimaryMockID string `db:"primary_mock_id"`
}

type DerivedRelationship struct {
	ID            string          `db:"id"`
	BaseMockID    string          `db:"base_mock_id"`
	DerivedMockID string          `db:"derived_mock_id"`
	DSLScriptID   string          `db:"dsl_script_id"`
	GenerationID  string          `db:"generation_id"`
	Params        json.RawMessage `db:"params"`
	CreatedAt     time.Time       `db:"created_at"`
}

type Generation struct {
	ID            string          `db:"id"`
	BaseMockID    string          `db:"base_mock_id"`
	DSLScriptID   string          `db:"dsl_script_id"`
	CreatedBy     string          `db:"created_by"`
	Status        string          `db:"status"`
	Params        json.RawMessage `db:"params"`
	Error         *string         `db:"error"`
	CreatedAt     time.Time       `db:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"`
	ResultMockIDs pq.StringArray  `db:"result_mock_ids"`
}

type GenerationResult struct {
	GenerationID string `db:"generation_id"`
	MockID       string `db:"mock_id"`
}

type MockSearch struct {
	MockID      string     `db:"mock_id"`
	FamilyID    *string    `db:"family_id"`
	Enabled     bool       `db:"enabled"`
	DeletedAt   *time.Time `db:"deleted_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	Method      *string    `db:"method"`
	Path        *string    `db:"path"`
	Name        string     `db:"name"`
	Description *string    `db:"description"`
}

type MockSearchFilters struct {
	Query    string
	FamilyID *string
	Enabled  *bool
	Method   *string
	Path     *string
	Limit    int
	Offset   int
}

type OutboxEvent struct {
	ID          string    `db:"id"`
	EventType   string    `db:"event_type"`
	Payload     []byte    `db:"payload"`
	Attempts    int       `db:"attempts"`
	MaxAttempts int       `db:"max_attempts"`
	LastError   *string   `db:"last_error"`
	NextAttempt time.Time `db:"next_attempt_at"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
