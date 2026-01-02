package application

import (
	"context"

	"mock-svc/internal/domain"

	"github.com/jmoiron/sqlx"
)

type FamilyView struct {
	FamilyID      string
	PrimaryMockID string
	Items         []FamilyItem
}

type FamilyItem struct {
	MockID            string
	IsPrimary         bool
	DerivedFromBaseID *string
	ViaDSLScriptID    *string
}

func (s *Service) CreateFamily(ctx context.Context, token, primaryMockID string) (domain.Family, error) {
	if primaryMockID == "" {
		return domain.Family{}, ErrInvalidRequest
	}
	allowed, _, _, err := s.Auth.CheckAccess(ctx, token, "mock", primaryMockID, "edit")
	if err != nil {
		return domain.Family{}, err
	}
	if !allowed {
		return domain.Family{}, ErrForbidden
	}
	primary, err := s.Repos.Mocks.GetActive(ctx, primaryMockID)
	if err != nil {
		return domain.Family{}, mapRepoError(err)
	}
	if primary.FamilyID != nil {
		return domain.Family{}, ErrConflict
	}

	var created domain.Family
	err = s.Tx.WithTx(ctx, func(tx *sqlx.Tx) error {
		family, err := s.Repos.Families.Create(ctx, tx, primaryMockID)
		if err != nil {
			return err
		}
		updated, err := s.Repos.Mocks.UpdateFamily(ctx, tx, primaryMockID, &family.ID)
		if err != nil {
			return err
		}
		method, path, err := extractMethodPath(updated.RequestMatch)
		if err != nil {
			return err
		}
		search := buildMockSearch(updated, &method, &path)
		if err := s.Repos.MockSearch.Update(ctx, tx, search); err != nil {
			return err
		}
		created = family
		return nil
	})
	if err != nil {
		return domain.Family{}, err
	}
	return created, nil
}

func (s *Service) AddFamilyMock(ctx context.Context, token, userID string, roles []string, familyID, mockID string) error {
	if familyID == "" || mockID == "" {
		return ErrInvalidRequest
	}
	allowed, _, _, err := s.Auth.CheckAccess(ctx, token, "mock", mockID, "edit")
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	family, err := s.Repos.Families.Get(ctx, familyID)
	if err != nil {
		return mapRepoError(err)
	}
	primary, err := s.Repos.Mocks.GetActive(ctx, family.PrimaryMockID)
	if err != nil {
		return mapRepoError(err)
	}
	if !isOwnerOrAdmin(userID, roles, primary.CreatedBy) {
		return ErrForbidden
	}
	mock, err := s.Repos.Mocks.GetActive(ctx, mockID)
	if err != nil {
		return mapRepoError(err)
	}
	if mock.FamilyID != nil && *mock.FamilyID != familyID {
		return ErrConflict
	}

	return s.Tx.WithTx(ctx, func(tx *sqlx.Tx) error {
		updated, err := s.Repos.Mocks.UpdateFamily(ctx, tx, mockID, &familyID)
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
}

func (s *Service) GetFamily(ctx context.Context, token, familyID string) (FamilyView, error) {
	if familyID == "" {
		return FamilyView{}, ErrInvalidRequest
	}
	family, err := s.Repos.Families.Get(ctx, familyID)
	if err != nil {
		return FamilyView{}, mapRepoError(err)
	}
	allowed, _, _, err := s.Auth.CheckAccess(ctx, token, "mock", family.PrimaryMockID, "read")
	if err != nil {
		return FamilyView{}, err
	}
	if !allowed {
		return FamilyView{}, ErrForbidden
	}

	mocks, err := s.Repos.Mocks.ListByFamily(ctx, familyID)
	if err != nil {
		return FamilyView{}, err
	}

	accessItems, err := s.Auth.ListResources(ctx, token, "mock", "read")
	if err != nil {
		return FamilyView{}, err
	}
	allowedSet := make(map[string]struct{}, len(accessItems))
	for _, item := range accessItems {
		allowedSet[item.ResourceID] = struct{}{}
	}

	ids := make([]string, 0, len(mocks))
	for _, mock := range mocks {
		if _, ok := allowedSet[mock.ID]; ok {
			ids = append(ids, mock.ID)
		}
	}
	rels, err := s.Repos.Derived.ListByDerivedIDs(ctx, ids)
	if err != nil {
		return FamilyView{}, err
	}
	byDerived := make(map[string]domain.DerivedRelationship, len(rels))
	for _, rel := range rels {
		byDerived[rel.DerivedMockID] = rel
	}

	items := make([]FamilyItem, 0, len(ids))
	for _, mock := range mocks {
		if _, ok := allowedSet[mock.ID]; !ok {
			continue
		}
		item := FamilyItem{
			MockID:    mock.ID,
			IsPrimary: mock.ID == family.PrimaryMockID,
		}
		if rel, ok := byDerived[mock.ID]; ok {
			item.DerivedFromBaseID = &rel.BaseMockID
			item.ViaDSLScriptID = &rel.DSLScriptID
		}
		items = append(items, item)
	}

	return FamilyView{
		FamilyID:      family.ID,
		PrimaryMockID: family.PrimaryMockID,
		Items:         items,
	}, nil
}

func (s *Service) UpdateFamilyPrimary(ctx context.Context, token, userID string, roles []string, familyID, newPrimaryID string) (domain.Family, error) {
	if familyID == "" || newPrimaryID == "" {
		return domain.Family{}, ErrInvalidRequest
	}
	allowed, _, _, err := s.Auth.CheckAccess(ctx, token, "mock", newPrimaryID, "edit")
	if err != nil {
		return domain.Family{}, err
	}
	if !allowed {
		return domain.Family{}, ErrForbidden
	}
	family, err := s.Repos.Families.Get(ctx, familyID)
	if err != nil {
		return domain.Family{}, mapRepoError(err)
	}
	currentPrimary, err := s.Repos.Mocks.GetActive(ctx, family.PrimaryMockID)
	if err != nil {
		return domain.Family{}, mapRepoError(err)
	}
	if !isOwnerOrAdmin(userID, roles, currentPrimary.CreatedBy) {
		return domain.Family{}, ErrForbidden
	}
	newPrimary, err := s.Repos.Mocks.GetActive(ctx, newPrimaryID)
	if err != nil {
		return domain.Family{}, mapRepoError(err)
	}
	if newPrimary.FamilyID != nil && *newPrimary.FamilyID != familyID {
		return domain.Family{}, ErrConflict
	}

	var updatedFamily domain.Family
	err = s.Tx.WithTx(ctx, func(tx *sqlx.Tx) error {
		family, err := s.Repos.Families.UpdatePrimary(ctx, tx, familyID, newPrimaryID)
		if err != nil {
			return err
		}
		updatedFamily = family
		updated, err := s.Repos.Mocks.UpdateFamily(ctx, tx, newPrimaryID, &familyID)
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
		return domain.Family{}, err
	}
	return updatedFamily, nil
}

func isOwnerOrAdmin(userID string, roles []string, ownerID string) bool {
	if userID != "" && userID == ownerID {
		return true
	}
	for _, role := range roles {
		if role == "ADMIN" {
			return true
		}
	}
	return false
}
