package service

import (
	"context"
	"time"

	"telegram-api/internal/domain"
	"telegram-api/pkg/crypto"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
)

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	logger.Debug().Str("username", req.Username).Str("email", req.Email).Msg("Starting registration")

	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		logger.Error().Err(err).Str("username", req.Username).Msg("Error checking username")
		return nil, domain.ErrDatabase
	}
	if exists {
		logger.Warn().Str("username", req.Username).Msg("Username already exists")
		return nil, domain.ErrUserAlreadyExists
	}

	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		logger.Error().Err(err).Str("email", req.Email).Msg("Error checking email")
		return nil, domain.ErrDatabase
	}
	if exists {
		logger.Warn().Str("email", req.Email).Msg("Email already exists")
		return nil, domain.ErrEmailAlreadyExists
	}

	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		logger.Error().Err(err).Msg("Error hashing password")
		return nil, domain.ErrInternal
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		IsActive:     true,
		Role:         domain.RoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.Error().Err(err).Str("username", req.Username).Msg("Error creating user in DB")
		return nil, domain.ErrDatabase
	}

	logger.Info().Str("id", user.ID.String()).Str("username", user.Username).Msg("User registered")
	return user, nil
}

// Login authenticates a user and returns tokens.
func (s *AuthService) Login(
	ctx context.Context,
	req *domain.LoginRequest,
	ipAddr, userAgent string,
) (*domain.LoginResponse, error) {
	logger.Debug().Str("username", req.Username).Str("ip", ipAddr).Msg("Login attempt")

	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		logger.Warn().Err(err).Str("username", req.Username).Msg("User not found")
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive {
		logger.Warn().Str("username", req.Username).Msg("User inactive")
		return nil, domain.ErrUserInactive
	}

	if !crypto.CheckPassword(req.Password, user.PasswordHash) {
		logger.Warn().Str("username", req.Username).Msg("Incorrect password")
		return nil, domain.ErrInvalidCredentials
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		logger.Error().Err(err).Msg("Error generating access token")
		return nil, domain.ErrInternal
	}

	refreshToken, err := s.generateRefreshToken(ctx, user.ID, ipAddr, userAgent)
	if err != nil {
		logger.Error().Err(err).Msg("Error generating refresh token")
		return nil, domain.ErrInternal
	}

	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	logger.Info().Str("id", user.ID.String()).Str("username", user.Username).Msg("Login successful")

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.config.JWT.ExpiryHours * 3600,
		User:         user.ToUserInfo(),
	}, nil
}

// RefreshTokens refreshes access and refresh tokens.
func (s *AuthService) RefreshTokens(
	ctx context.Context,
	refreshTokenStr, ipAddr, userAgent string,
) (*domain.LoginResponse, error) {
	tokenHash := crypto.HashToken(refreshTokenStr)

	token, err := s.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		logger.Warn().Err(err).Msg("Refresh token not found")
		return nil, domain.ErrInvalidToken
	}

	if token.RevokedAt != nil {
		logger.Warn().Str("token_id", token.ID.String()).Msg("Token revoked")
		return nil, domain.ErrTokenRevoked
	}

	if time.Now().After(token.ExpiresAt) {
		logger.Warn().Str("token_id", token.ID.String()).Msg("Token expired")
		return nil, domain.ErrTokenExpired
	}

	user, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		logger.Error().Err(err).Msg("Token user not found")
		return nil, domain.ErrUserNotFound
	}

	if !user.IsActive {
		return nil, domain.ErrUserInactive
	}

	_ = s.tokenRepo.Revoke(ctx, token.ID)

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		logger.Error().Err(err).Msg("Error generating new access token")
		return nil, domain.ErrInternal
	}

	newRefreshToken, err := s.generateRefreshToken(ctx, user.ID, ipAddr, userAgent)
	if err != nil {
		logger.Error().Err(err).Msg("Error generating new refresh token")
		return nil, domain.ErrInternal
	}

	logger.Info().Str("user_id", user.ID.String()).Msg("Tokens refreshed")

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.config.JWT.ExpiryHours * 3600,
		User:         user.ToUserInfo(),
	}, nil
}

// Logout revokes a refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshTokenStr string) error {
	tokenHash := crypto.HashToken(refreshTokenStr)
	token, err := s.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil
	}
	logger.Info().Str("token_id", token.ID.String()).Msg("Logout")
	return s.tokenRepo.Revoke(ctx, token.ID)
}

// LogoutAll revokes all refresh tokens for a user.
func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	logger.Info().Str("user_id", userID.String()).Msg("Logout all devices")
	return s.tokenRepo.RevokeAllForUser(ctx, userID)
}
