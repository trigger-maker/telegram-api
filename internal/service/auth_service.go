// Package service provides business logic services.
package service

import (
	"telegram-api/internal/config"
	"telegram-api/internal/domain"
)

// AuthService manages authentication operations.
type AuthService struct {
	userRepo  domain.UserRepository
	tokenRepo domain.RefreshTokenRepository
	cacheRepo domain.CacheRepository
	config    *config.Config
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(
	userRepo domain.UserRepository,
	tokenRepo domain.RefreshTokenRepository,
	cacheRepo domain.CacheRepository,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		cacheRepo: cacheRepo,
		config:    cfg,
	}
}
