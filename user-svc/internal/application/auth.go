package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"user-svc/internal/domain"
	"user-svc/internal/infrastructure/postgres"

	"golang.org/x/crypto/bcrypt"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

func (s *Service) Register(ctx context.Context, email, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	user, err := s.Repos.Users.Create(ctx, email, string(hash), "ACTIVE")
	if err != nil {
		if postgres.IsUniqueViolation(err) {
			return "", ErrConflict
		}
		return "", err
	}
	if err := s.Repos.Roles.ReplaceUserRoles(ctx, user.ID, []string{"USER"}); err != nil {
		return "", err
	}
	return user.ID, nil
}

func (s *Service) Login(ctx context.Context, email, password string, ip, userAgent *string) (TokenPair, error) {
	user, err := s.Repos.Users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return TokenPair{}, ErrInvalidCredentials
		}
		return TokenPair{}, err
	}
	if user.Status != "ACTIVE" {
		return TokenPair{}, ErrForbidden
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}
	roles, err := s.Repos.Roles.ListByUserID(ctx, user.ID)
	if err != nil {
		return TokenPair{}, err
	}
	_ = s.Repos.Users.UpdateLastLogin(ctx, user.ID)

	refreshSecret, err := newToken()
	if err != nil {
		return TokenPair{}, err
	}
	tempHash := hashToken(refreshSecret)
	session, err := s.Repos.Sessions.Create(ctx, user.ID, tempHash, ip, userAgent)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken := session.ID + "." + refreshSecret
	refreshHash := hashToken(refreshToken)
	if err := s.Repos.Sessions.RotateRefresh(ctx, session.ID, refreshHash); err != nil {
		_ = s.Repos.Sessions.RevokeByID(ctx, session.ID)
		return TokenPair{}, err
	}

	accessToken, expiresAt, err := s.JWT.NewAccessToken(user.ID, roles, s.AccessTTL)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(time.Until(expiresAt).Seconds()),
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	sessionID, ok := parseRefreshToken(refreshToken)
	if !ok {
		return TokenPair{}, ErrInvalidToken
	}

	session, err := s.Repos.Sessions.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return TokenPair{}, ErrTokenRevoked
		}
		return TokenPair{}, err
	}
	if session.RevokedAt != nil {
		return TokenPair{}, ErrTokenRevoked
	}
	if time.Since(session.CreatedAt) > s.RefreshTTL {
		_ = s.Repos.Sessions.RevokeByID(ctx, session.ID)
		return TokenPair{}, ErrTokenExpired
	}
	if !matchTokenHash(session.RefreshHash, refreshToken) {
		_ = s.Repos.Sessions.RevokeByID(ctx, session.ID)
		return TokenPair{}, ErrTokenRevoked
	}

	user, err := s.Repos.Users.GetByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return TokenPair{}, ErrNotFound
		}
		return TokenPair{}, err
	}
	if user.Status != "ACTIVE" {
		return TokenPair{}, ErrForbidden
	}
	roles, err := s.Repos.Roles.ListByUserID(ctx, user.ID)
	if err != nil {
		return TokenPair{}, err
	}

	newSecret, err := newToken()
	if err != nil {
		return TokenPair{}, err
	}
	newRefresh := session.ID + "." + newSecret
	newHash := hashToken(newRefresh)
	if err := s.Repos.Sessions.RotateRefresh(ctx, session.ID, newHash); err != nil {
		return TokenPair{}, err
	}

	accessToken, expiresAt, err := s.JWT.NewAccessToken(user.ID, roles, s.AccessTTL)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefresh,
		ExpiresIn:    int64(time.Until(expiresAt).Seconds()),
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	sessionID, ok := parseRefreshToken(refreshToken)
	if !ok {
		return ErrInvalidToken
	}
	session, err := s.Repos.Sessions.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}
	if session.RevokedAt != nil {
		return nil
	}
	if !matchTokenHash(session.RefreshHash, refreshToken) {
		_ = s.Repos.Sessions.RevokeByID(ctx, session.ID)
		return nil
	}
	return s.Repos.Sessions.RevokeByID(ctx, session.ID)
}

func (s *Service) Me(ctx context.Context, userID string) (domain.User, []string, error) {
	user, err := s.Repos.Users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.User{}, nil, ErrNotFound
		}
		return domain.User{}, nil, err
	}
	roles, err := s.Repos.Roles.ListByUserID(ctx, userID)
	if err != nil {
		return domain.User{}, nil, err
	}
	return user, roles, nil
}

func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func matchTokenHash(storedHash, token string) bool {
	sum := hashToken(token)
	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(sum)) == 1
}

func parseRefreshToken(token string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", false
	}
	return parts[0], true
}
