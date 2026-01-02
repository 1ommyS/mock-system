package application

import (
	"time"

	"user-svc/internal/auth"
)

type Service struct {
	Repos      Repositories
	JWT        *auth.JWTService
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func New(repos Repositories, jwt *auth.JWTService, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		Repos:      repos,
		JWT:        jwt,
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	}
}
