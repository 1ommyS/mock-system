package application

import (
	"context"
	"errors"

	"user-svc/internal/domain"
)

func (s *Service) UpdateUser(ctx context.Context, userID string, status *string, roles []string) (domain.User, []string, error) {
	var user domain.User
	var err error

	if status != nil && *status != "" {
		if !isValidStatus(*status) {
			return domain.User{}, nil, ErrInvalidRequest
		}
		user, err = s.Repos.Users.UpdateStatus(ctx, userID, *status)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.User{}, nil, ErrNotFound
			}
			return domain.User{}, nil, err
		}
	} else {
		user, err = s.Repos.Users.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.User{}, nil, ErrNotFound
			}
			return domain.User{}, nil, err
		}
	}

	if roles != nil {
		if err := s.Repos.Roles.ReplaceUserRoles(ctx, userID, roles); err != nil {
			return domain.User{}, nil, err
		}
	}

	currentRoles, err := s.Repos.Roles.ListByUserID(ctx, userID)
	if err != nil {
		return domain.User{}, nil, err
	}

	return user, currentRoles, nil
}

func isValidStatus(status string) bool {
	switch status {
	case "ACTIVE", "BLOCKED", "DELETED":
		return true
	default:
		return false
	}
}
