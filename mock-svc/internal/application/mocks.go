package application

import (
	"bytes"
	"context"
	"encoding/json"

	"mock-svc/internal/domain"

	"github.com/jmoiron/sqlx"
)

const maxJSONSize = 256 * 1024

type CreateMockInput struct {
	Name             string
	Description      *string
	RequestMatch     json.RawMessage
	ResponseTemplate json.RawMessage
	Enabled          *bool
}

func (s *Service) CreateMock(ctx context.Context, userID, token string, input CreateMockInput) (domain.Mock, error) {
	if userID == "" {
		return domain.Mock{}, ErrUnauthorized
	}
	if input.Name == "" || len(input.RequestMatch) == 0 || len(input.ResponseTemplate) == 0 {
		return domain.Mock{}, ErrInvalidRequest
	}
	if err := validateJSON(input.RequestMatch); err != nil {
		return domain.Mock{}, err
	}
	if err := validateJSON(input.ResponseTemplate); err != nil {
		return domain.Mock{}, err
	}
	method, path, err := extractMethodPath(input.RequestMatch)
	if err != nil {
		return domain.Mock{}, err
	}
	if method == "" || path == "" {
		return domain.Mock{}, ErrValidation
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	mock := domain.Mock{
		Name:             input.Name,
		Description:      input.Description,
		RequestMatch:     input.RequestMatch,
		ResponseTemplate: input.ResponseTemplate,
		Enabled:          enabled,
		CreatedBy:        userID,
	}

	var created domain.Mock
	err = s.Tx.WithTx(ctx, func(tx *sqlx.Tx) error {
		var err error
		created, err = s.Repos.Mocks.Create(ctx, tx, mock)
		if err != nil {
			return err
		}
		search := buildMockSearch(created, &method, &path)
		if err := s.Repos.MockSearch.Upsert(ctx, tx, search); err != nil {
			return err
		}
		payload, err := json.Marshal(map[string]string{
			"mockId":      created.ID,
			"ownerUserId": created.CreatedBy,
		})
		if err != nil {
			return err
		}
		return s.Repos.Outbox.Enqueue(ctx, tx, "register_mock", payload, s.OutboxMaxAttempts)
	})
	if err != nil {
		return domain.Mock{}, err
	}
	return created, nil
}

func (s *Service) GetMock(ctx context.Context, token, id string) (domain.Mock, error) {
	if id == "" {
		return domain.Mock{}, ErrInvalidRequest
	}
	allowed, _, _, err := s.Auth.CheckAccess(ctx, token, "mock", id, "read")
	if err != nil {
		return domain.Mock{}, err
	}
	if !allowed {
		return domain.Mock{}, ErrForbidden
	}
	mock, err := s.Repos.Mocks.GetActive(ctx, id)
	if err != nil {
		return domain.Mock{}, mapRepoError(err)
	}
	return mock, nil
}

func (s *Service) UpdateMock(ctx context.Context, token, id string, patch domain.MockPatch) (domain.Mock, error) {
	if id == "" {
		return domain.Mock{}, ErrInvalidRequest
	}
	allowed, _, _, err := s.Auth.CheckAccess(ctx, token, "mock", id, "edit")
	if err != nil {
		return domain.Mock{}, err
	}
	if !allowed {
		return domain.Mock{}, ErrForbidden
	}
	if patch.Name == nil && patch.Description == nil && patch.RequestMatch == nil && patch.ResponseTemplate == nil && patch.Enabled == nil {
		return domain.Mock{}, ErrInvalidRequest
	}
	if patch.RequestMatch != nil {
		if err := validateJSON(*patch.RequestMatch); err != nil {
			return domain.Mock{}, err
		}
		method, path, err := extractMethodPath(*patch.RequestMatch)
		if err != nil {
			return domain.Mock{}, err
		}
		if method == "" || path == "" {
			return domain.Mock{}, ErrValidation
		}
	}
	if patch.ResponseTemplate != nil {
		if err := validateJSON(*patch.ResponseTemplate); err != nil {
			return domain.Mock{}, err
		}
	}

	var updated domain.Mock
	err = s.Tx.WithTx(ctx, func(tx *sqlx.Tx) error {
		var err error
		updated, err = s.Repos.Mocks.Update(ctx, tx, id, patch)
		if err != nil {
			return err
		}
		method, path, err := extractMethodPath(updated.RequestMatch)
		if err != nil {
			return err
		}
		search := buildMockSearch(updated, &method, &path)
		return s.Repos.MockSearch.Update(ctx, tx, search)
	})
	if err != nil {
		return domain.Mock{}, mapRepoError(err)
	}
	return updated, nil
}

func (s *Service) DeleteMock(ctx context.Context, token, id string) error {
	if id == "" {
		return ErrInvalidRequest
	}
	allowed, _, _, err := s.Auth.CheckAccess(ctx, token, "mock", id, "edit")
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}

	return s.Tx.WithTx(ctx, func(tx *sqlx.Tx) error {
		mock, err := s.Repos.Mocks.SoftDelete(ctx, tx, id)
		if err != nil {
			return err
		}
		method, path, err := extractMethodPath(mock.RequestMatch)
		if err != nil {
			return err
		}
		search := buildMockSearch(mock, &method, &path)
		return s.Repos.MockSearch.Update(ctx, tx, search)
	})
}

func (s *Service) ListMocks(ctx context.Context, token string, filters domain.MockSearchFilters) ([]domain.Mock, error) {
	items, err := s.Auth.ListResources(ctx, token, "mock", "read")
	if err != nil {
		return nil, err
	}
	allowed := make([]string, 0, len(items))
	for _, item := range items {
		allowed = append(allowed, item.ResourceID)
	}
	ids, err := s.Repos.MockSearch.Search(ctx, allowed, filters)
	if err != nil {
		return nil, err
	}
	mocks, err := s.Repos.Mocks.ListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	if len(mocks) == 0 {
		return nil, nil
	}
	byID := make(map[string]domain.Mock, len(mocks))
	for _, mock := range mocks {
		byID[mock.ID] = mock
	}
	ordered := make([]domain.Mock, 0, len(ids))
	for _, id := range ids {
		if mock, ok := byID[id]; ok {
			ordered = append(ordered, mock)
		}
	}
	return ordered, nil
}

func validateJSON(data json.RawMessage) error {
	if len(data) > maxJSONSize {
		return ErrValidation
	}
	if len(data) == 0 {
		return ErrValidation
	}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return ErrValidation
	}
	if !json.Valid(data) {
		return ErrValidation
	}
	return nil
}

func extractMethodPath(data json.RawMessage) (string, string, error) {
	if len(data) == 0 {
		return "", "", ErrValidation
	}
	var payload struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", "", ErrValidation
	}
	return payload.Method, payload.Path, nil
}

func buildMockSearch(mock domain.Mock, method, path *string) domain.MockSearch {
	return domain.MockSearch{
		MockID:      mock.ID,
		FamilyID:    mock.FamilyID,
		Enabled:     mock.Enabled,
		DeletedAt:   mock.DeletedAt,
		UpdatedAt:   mock.UpdatedAt,
		Method:      method,
		Path:        path,
		Name:        mock.Name,
		Description: mock.Description,
	}
}
