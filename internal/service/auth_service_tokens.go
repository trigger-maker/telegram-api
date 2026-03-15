package service

import (
	"context"
	"time"

	"telegram-api/internal/domain"
	"telegram-api/pkg/crypto"
	"telegram-api/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID   string      `json:"uid"`
	Username string      `json:"username"`
	Role     domain.Role `json:"role"`
	jwt.RegisteredClaims
}

// ValidateToken validates a JWT token and returns claims
func (s *AuthService) ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// generateAccessToken generates a JWT access token for a user
func (s *AuthService) generateAccessToken(user *domain.User) (string, error) {
	expiresAt := time.Now().Add(time.Duration(s.config.JWT.ExpiryHours) * time.Hour)

	claims := &JWTClaims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "telegram-api",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Secret))
}

// generateRefreshToken generates a refresh token for a user
func (s *AuthService) generateRefreshToken(ctx context.Context, userID uuid.UUID, ipAddr, userAgent string) (string, error) {
	tokenBytes, err := crypto.GenerateRandomBytes(32)
	if err != nil {
		return "", err
	}
	tokenStr := crypto.HashToken(string(tokenBytes))

	refreshToken := &domain.RefreshToken{
		ID:         uuid.New(),
		UserID:     userID,
		TokenHash:  crypto.HashToken(tokenStr),
		DeviceInfo: userAgent,
		IPAddress:  ipAddr,
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:  time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, refreshToken); err != nil {
		logger.Error().Err(err).Msg("Error saving refresh token")
		return "", err
	}

	return tokenStr, nil
}
