package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"mock-svc/internal/domain"

	"github.com/jmoiron/sqlx"
)

const (
	GenerationQueued  = "QUEUED"
	GenerationRunning = "RUNNING"
	GenerationDone    = "DONE"
	GenerationFailed  = "FAILED"
)

type GenerationInput struct {
	BaseMockID  string
	DSLScriptID string
	DSLScript   string
	Params      json.RawMessage
}

func (s *Service) PreviewGeneration(ctx context.Context, token, userID string, input GenerationInput) ([]GeneratedMock, error) {
	slog.Info(
		"generation preview start",
		"user_id",
		userID,
		"base_mock_id",
		input.BaseMockID,
		"dsl_script_id",
		input.DSLScriptID,
	)
	if input.BaseMockID == "" || input.DSLScriptID == "" || input.DSLScript == "" {
		return nil, ErrInvalidRequest
	}
	if err := validateOptionalJSON(input.Params); err != nil {
		return nil, err
	}
	if err := s.checkGenerationAccess(ctx, token, input.BaseMockID, input.DSLScriptID); err != nil {
		return nil, err
	}
	base, err := s.Repos.Mocks.GetActive(ctx, input.BaseMockID)
	if err != nil {
		return nil, mapRepoError(err)
	}
	planned, err := s.DSL.Preview(ctx, toDSLBaseMock(base), input.DSLScriptID, input.DSLScript, input.Params)
	if err != nil {
		slog.Error(
			"generation preview failed",
			"user_id",
			userID,
			"base_mock_id",
			input.BaseMockID,
			"dsl_script_id",
			input.DSLScriptID,
			"error",
			err,
		)
		return nil, err
	}
	slog.Info(
		"generation preview done",
		"user_id",
		userID,
		"base_mock_id",
		input.BaseMockID,
		"dsl_script_id",
		input.DSLScriptID,
		"count",
		len(planned),
	)
	return planned, nil
}

func (s *Service) StartGeneration(ctx context.Context, token, userID string, input GenerationInput) (domain.Generation, error) {
	if userID == "" {
		return domain.Generation{}, ErrUnauthorized
	}
	if input.BaseMockID == "" || input.DSLScriptID == "" || input.DSLScript == "" {
		return domain.Generation{}, ErrInvalidRequest
	}
	if err := validateOptionalJSON(input.Params); err != nil {
		return domain.Generation{}, err
	}
	params := ensureJSON(input.Params)
	if err := s.checkGenerationAccess(ctx, token, input.BaseMockID, input.DSLScriptID); err != nil {
		return domain.Generation{}, err
	}
	base, err := s.Repos.Mocks.GetActive(ctx, input.BaseMockID)
	if err != nil {
		return domain.Generation{}, mapRepoError(err)
	}

	gen := domain.Generation{
		BaseMockID:  base.ID,
		DSLScriptID: input.DSLScriptID,
		CreatedBy:   userID,
		Status:      GenerationQueued,
		Params:      params,
	}
	var created domain.Generation
	err = s.Tx.WithTx(ctx, func(tx *sqlx.Tx) error {
		var err error
		created, err = s.Repos.Generations.Create(ctx, tx, gen)
		return err
	})
	if err != nil {
		return domain.Generation{}, err
	}

	slog.Info(
		"generation apply queued",
		"generation_id",
		created.ID,
		"user_id",
		userID,
		"base_mock_id",
		base.ID,
		"dsl_script_id",
		input.DSLScriptID,
	)

	go s.runGeneration(created.ID, base, input.DSLScriptID, input.DSLScript, input.Params, userID)
	return created, nil
}

func (s *Service) GetGeneration(ctx context.Context, token, userID string, roles []string, generationID string) (domain.Generation, error) {
	if generationID == "" {
		return domain.Generation{}, ErrInvalidRequest
	}
	gen, err := s.Repos.Generations.Get(ctx, generationID)
	if err != nil {
		return domain.Generation{}, mapRepoError(err)
	}
	if userID != "" && userID == gen.CreatedBy {
		return gen, nil
	}
	for _, role := range roles {
		if role == "ADMIN" {
			return gen, nil
		}
	}
	allowed, _, _, err := s.Auth.CheckAccess(ctx, token, "mock", gen.BaseMockID, "edit")
	if err != nil {
		return domain.Generation{}, err
	}
	if !allowed {
		return domain.Generation{}, ErrForbidden
	}
	return gen, nil
}

func (s *Service) checkGenerationAccess(ctx context.Context, token, baseMockID, dslScriptID string) error {
	allowed, _, _, err := s.Auth.CheckAccess(ctx, token, "mock", baseMockID, "edit")
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	allowed, _, _, err = s.Auth.CheckAccess(ctx, token, "dsl_script", dslScriptID, "edit")
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return nil
}

func (s *Service) runGeneration(generationID string, base domain.Mock, dslScriptID string, dslScript string, params json.RawMessage, initiator string) {
	ctx := context.Background()
	slog.Info(
		"generation apply started",
		"generation_id",
		generationID,
		"base_mock_id",
		base.ID,
		"dsl_script_id",
		dslScriptID,
	)
	_ = s.Tx.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.Repos.Generations.UpdateStatus(ctx, tx, generationID, GenerationRunning, nil, nil)
	})

	derived, err := s.DSL.Apply(ctx, toDSLBaseMock(base), dslScriptID, dslScript, params)
	if err != nil {
		s.failGeneration(ctx, generationID, fmt.Sprintf("dsl apply error: %v", err))
		return
	}

	slog.Info(
		"generation apply produced",
		"generation_id",
		generationID,
		"count",
		len(derived),
	)

	for _, item := range derived {
		if err := validateGeneratedMock(item); err != nil {
			s.failGeneration(ctx, generationID, fmt.Sprintf("invalid derived mock: %v", err))
			return
		}
	}

	resultIDs := make([]string, 0, len(derived))
	err = s.Tx.WithTx(ctx, func(tx *sqlx.Tx) error {
		for _, item := range derived {
			enabled := base.Enabled
			if item.Enabled != nil {
				enabled = *item.Enabled
			}
			mock := domain.Mock{
				Name:             item.Name,
				Description:      item.Description,
				RequestMatch:     item.RequestMatch,
				ResponseTemplate: item.ResponseTemplate,
				Enabled:          enabled,
				FamilyID:         base.FamilyID,
				CreatedBy:        base.CreatedBy,
				GeneratedBy:      &initiator,
			}
			created, err := s.Repos.Mocks.Create(ctx, tx, mock)
			if err != nil {
				return err
			}
			method, path, err := extractMethodPath(created.RequestMatch)
			if err != nil {
				return err
			}
			search := buildMockSearch(created, &method, &path)
			if err := s.Repos.MockSearch.Upsert(ctx, tx, search); err != nil {
				return err
			}
			payload, err := json.Marshal(map[string]string{
				"mockId":      created.ID,
				"ownerUserId": base.CreatedBy,
			})
			if err != nil {
				return err
			}
			if err := s.Repos.Outbox.Enqueue(ctx, tx, "register_mock", payload, s.OutboxMaxAttempts); err != nil {
				return err
			}
			rel := domain.DerivedRelationship{
				BaseMockID:    base.ID,
				DerivedMockID: created.ID,
				DSLScriptID:   dslScriptID,
				GenerationID:  generationID,
				Params:        params,
			}
			if err := s.Repos.Derived.Create(ctx, tx, rel); err != nil {
				return err
			}
			resultIDs = append(resultIDs, created.ID)
		}
		return s.Repos.Generations.UpdateStatus(ctx, tx, generationID, GenerationDone, resultIDs, nil)
	})
	if err != nil {
		s.failGeneration(ctx, generationID, fmt.Sprintf("generation failed: %v", err))
	}
}

func (s *Service) failGeneration(ctx context.Context, generationID, message string) {
	slog.Error("generation failed", "generation_id", generationID, "error", message)
	_ = s.Tx.WithTx(ctx, func(tx *sqlx.Tx) error {
		return s.Repos.Generations.UpdateStatus(ctx, tx, generationID, GenerationFailed, nil, &message)
	})
}

func toDSLBaseMock(mock domain.Mock) DSLBaseMock {
	return DSLBaseMock{
		ID:               mock.ID,
		Name:             mock.Name,
		Description:      mock.Description,
		RequestMatch:     mock.RequestMatch,
		ResponseTemplate: mock.ResponseTemplate,
		Enabled:          mock.Enabled,
	}
}

func validateOptionalJSON(data json.RawMessage) error {
	if len(data) == 0 {
		return nil
	}
	return validateJSON(data)
}

func ensureJSON(data json.RawMessage) json.RawMessage {
	if len(data) == 0 {
		return json.RawMessage(`{}`)
	}
	return data
}

func validateGeneratedMock(item GeneratedMock) error {
	if item.Name == "" {
		return ErrValidation
	}
	if err := validateJSON(item.RequestMatch); err != nil {
		return err
	}
	if err := validateJSON(item.ResponseTemplate); err != nil {
		return err
	}
	method, path, err := extractMethodPath(item.RequestMatch)
	if err != nil {
		return err
	}
	if method == "" || path == "" {
		return ErrValidation
	}
	return nil
}
