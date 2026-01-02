package application

import (
	"context"
	"errors"
	"strings"

	"user-svc/internal/domain"
	"user-svc/internal/infrastructure/postgres"
)

func (s *Service) RegisterResource(ctx context.Context, resourceType, resourceID, ownerUserID string) (domain.Resource, bool, error) {
	return s.Repos.Resources.Register(ctx, resourceType, resourceID, ownerUserID)
}

func (s *Service) CheckAccess(ctx context.Context, userID, resourceType, resourceID, action string) (bool, string, bool, error) {
	access, err := s.Repos.Resources.GetAccess(ctx, userID, resourceType, resourceID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return false, "NONE", false, nil
		}
		return false, "", false, err
	}

	allowed := false
	switch strings.ToLower(action) {
	case "read":
		allowed = access.EffectivePermission == "READ" || access.EffectivePermission == "EDIT"
	case "edit":
		allowed = access.EffectivePermission == "EDIT"
	default:
		return false, "", false, ErrInvalidRequest
	}

	return allowed, access.EffectivePermission, access.IsOwner, nil
}

func (s *Service) ListResources(ctx context.Context, userID, resourceType, minPermission string) ([]domain.ResourceAccess, error) {
	return s.Repos.Resources.ListByUser(ctx, userID, resourceType, minPermission)
}

func (s *Service) CreateGrant(ctx context.Context, actorID string, roles []string, resourceType, resourceID, granteeUserID, permission string) (domain.Grant, error) {
	ok, err := checkOwnerOrAdmin(ctx, s.Repos, actorID, roles, resourceType, resourceID)
	if err != nil {
		return domain.Grant{}, err
	}
	if !ok {
		return domain.Grant{}, ErrForbidden
	}
	grant, err := s.Repos.Grants.Create(ctx, resourceType, resourceID, granteeUserID, permission)
	if err != nil {
		if postgres.IsUniqueViolation(err) {
			return domain.Grant{}, ErrConflict
		}
		if postgres.IsForeignKeyViolation(err) {
			return domain.Grant{}, ErrNotFound
		}
		return domain.Grant{}, err
	}
	return grant, nil
}

func (s *Service) DeleteGrant(ctx context.Context, actorID string, roles []string, resourceType, resourceID, granteeUserID string) error {
	ok, err := checkOwnerOrAdmin(ctx, s.Repos, actorID, roles, resourceType, resourceID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	if err := s.Repos.Grants.Delete(ctx, resourceType, resourceID, granteeUserID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *Service) ListGrants(ctx context.Context, actorID string, roles []string, resourceType, resourceID string) ([]domain.Grant, error) {
	ok, err := checkOwnerOrAdmin(ctx, s.Repos, actorID, roles, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	grants, err := s.Repos.Grants.ListByResource(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	return grants, nil
}

func (s *Service) UpdateResourceOwner(ctx context.Context, resourceType, resourceID, ownerUserID string) (domain.Resource, error) {
	res, err := s.Repos.Resources.UpdateOwner(ctx, resourceType, resourceID, ownerUserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Resource{}, ErrNotFound
		}
		if postgres.IsForeignKeyViolation(err) {
			return domain.Resource{}, ErrNotFound
		}
		return domain.Resource{}, err
	}
	return res, nil
}

func checkOwnerOrAdmin(ctx context.Context, repos Repositories, actorID string, roles []string, resourceType, resourceID string) (bool, error) {
	if hasRole(roles, "ADMIN") {
		return true, nil
	}
	res, err := repos.Resources.Get(ctx, resourceType, resourceID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return false, ErrNotFound
		}
		return false, err
	}
	return res.OwnerUserID == actorID, nil
}

func hasRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
