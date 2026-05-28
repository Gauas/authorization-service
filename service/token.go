package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gauas/authorization-service/model"
	"github.com/gauas/authorization-service/packages/jwt"
	"github.com/gauas/authorization-service/packages/memory"
	"github.com/google/uuid"
)

type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresIn        int       `json:"expires_in"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type RenewResult struct {
	AccessToken string    `json:"access_token"`
	ExpiresIn   int       `json:"expires_in"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func (s *Service) CreateToken(ctx context.Context, userID uuid.UUID, permission, deviceID string) (*TokenPair, error) {
	if deviceID == "" {
		return nil, appError(http.StatusBadRequest, "device_id is required")
	}

	tokenID, err := s.memory.NextTokenID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: next token id: %w", err)
	}

	accessToken, expiresAt, err := s.jwt.Sign(userID, permission, deviceID, tokenID)
	if err != nil {
		return nil, fmt.Errorf("service: sign access token: %w", err)
	}

	refreshToken, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("service: generate refresh token: %w", err)
	}

	ttl := time.Duration(s.config.RefreshTTLDays) * 24 * time.Hour
	now := time.Now()

	if _, err := s.repo.Token.Create(ctx, &model.Token{
		UserID:       userID,
		DeviceID:     deviceID,
		Permission:   permission,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(ttl),
	}); err != nil {
		return nil, fmt.Errorf("service: persist refresh token: %w", err)
	}

	data := memory.RefreshTokenData{
		UserID:     userID,
		DeviceID:   deviceID,
		Permission: permission,
		TokenID:    tokenID,
	}
	if err := s.memory.StoreRefreshToken(ctx, refreshToken, data, ttl); err != nil {
		_ = s.repo.Token.Delete(ctx, "refresh_token = ?", refreshToken)
		return nil, fmt.Errorf("service: cache refresh token: %w", err)
	}

	_ = s.memory.TrackTokenForDevice(ctx, userID, deviceID, refreshToken, ttl)

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int(time.Until(expiresAt).Seconds()),
		ExpiresAt:        expiresAt,
		RefreshExpiresAt: now.Add(ttl),
	}, nil
}

func (s *Service) ValidateToken(ctx context.Context, tokenStr string) (*jwt.Claims, error) {
	claims, err := s.jwt.Verify(tokenStr)
	if err != nil {
		return nil, appError(http.StatusUnauthorized, "invalid or expired token")
	}

	blacklisted, err := s.memory.IsTokenBlacklisted(ctx, claims.UserID, claims.TokenID)
	if err != nil {
		return nil, fmt.Errorf("service: check blacklist: %w", err)
	}
	if blacklisted {
		return nil, appError(http.StatusUnauthorized, "token has been revoked")
	}

	return claims, nil
}

func (s *Service) RenewToken(ctx context.Context, refreshToken, deviceID string) (*RenewResult, error) {
	if refreshToken == "" {
		return nil, appError(http.StatusBadRequest, "refresh_token is required")
	}
	if deviceID == "" {
		return nil, appError(http.StatusBadRequest, "device_id is required")
	}

	data, err := s.memory.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("service: get refresh token: %w", err)
	}
	if data == nil {
		return nil, appError(http.StatusUnauthorized, "refresh token not found or expired")
	}
	if data.DeviceID != deviceID {
		return nil, appError(http.StatusUnauthorized, "device mismatch")
	}

	tokenID, err := s.memory.NextTokenID(ctx, data.UserID)
	if err != nil {
		return nil, fmt.Errorf("service: next token id: %w", err)
	}

	accessToken, expiresAt, err := s.jwt.Sign(data.UserID, data.Permission, deviceID, tokenID)
	if err != nil {
		return nil, fmt.Errorf("service: sign renewed token: %w", err)
	}

	return &RenewResult{
		AccessToken: accessToken,
		ExpiresIn:   int(time.Until(expiresAt).Seconds()),
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *Service) RevokeToken(ctx context.Context, refreshToken, deviceID string) error {
	if refreshToken == "" {
		return appError(http.StatusBadRequest, "refresh_token is required")
	}

	data, err := s.memory.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("service: get refresh token: %w", err)
	}
	if data == nil {
		return nil
	}
	if deviceID != "" && data.DeviceID != deviceID {
		return appError(http.StatusUnauthorized, "device mismatch")
	}

	accessTTL := time.Duration(s.config.JWTExpireSecs) * time.Second
	_ = s.memory.BlacklistToken(ctx, data.UserID, data.TokenID, accessTTL)
	_ = s.repo.Token.Delete(ctx, "refresh_token = ?", refreshToken)
	return s.memory.DeleteRefreshToken(ctx, refreshToken)
}
