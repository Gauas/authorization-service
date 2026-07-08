package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	response "github.com/gauas/authorization-service/dto/response"
	"github.com/gauas/authorization-service/model"
	"github.com/gauas/authorization-service/packages/jwt"
	"github.com/gauas/authorization-service/packages/memory"
	"github.com/google/uuid"
)

const (
	SEC_PER_DAY      = 86400
	BLACKLIST_BUFFER = 1
)

func (s *Service) CreateToken(ctx context.Context, userID int64, permission, deviceID string) (*response.TokenPair, error) {
	if deviceID == "" {
		return nil, appError(http.StatusBadRequest, "device_id is required")
	}

	refreshToken, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("service: generate refresh token: %w", err)
	}

	ttl := time.Duration(s.Config.RefreshTTLDays) * 24 * time.Hour
	now := time.Now()
	key, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("service: uuid: %w", err)
	}

	record, err := s.Repo.Create(ctx, &model.Token{
		Key:          key,
		UserID:       userID,
		DeviceID:     deviceID,
		Permission:   permission,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(ttl),
	})
	if err != nil {
		return nil, fmt.Errorf("service: persist refresh token: %w", err)
	}
	accessToken, expiresAt, err := s.Signer.Sign(userID, permission, deviceID, record.ID)
	if err != nil {
		_ = s.Repo.Delete(ctx, "id = ?", record.ID)
		return nil, fmt.Errorf("service: sign access token: %w", err)
	}

	data := memory.RefreshTokenData{
		UserID:     userID,
		DeviceID:   deviceID,
		Permission: permission,
		TokenID:    record.ID,
	}
	if err := s.Cache.StoreRefreshToken(ctx, refreshToken, data, ttl); err != nil {
		_ = s.Repo.Delete(ctx, "refresh_token = ?", refreshToken)
		return nil, fmt.Errorf("service: cache refresh token: %w", err)
	}

	_ = s.Cache.TrackTokenForDevice(ctx, userID, deviceID, refreshToken, ttl)

	return &response.TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int(time.Until(expiresAt).Seconds()),
		ExpiresAt:        expiresAt,
		RefreshExpiresAt: now.Add(ttl),
	}, nil
}

func (s *Service) ValidateToken(ctx context.Context, tokenStr string) (*jwt.Claims, error) {
	claims, err := s.Signer.Verify(tokenStr)
	if err != nil {
		return nil, appError(http.StatusUnauthorized, "invalid or expired token")
	}

	blacklisted, err := s.Cache.IsTokenBlacklisted(ctx, claims.TokenID, s.blacklistWindowDays())
	if err != nil {
		return nil, fmt.Errorf("service: check blacklist: %w", err)
	}
	if blacklisted {
		return nil, appError(http.StatusUnauthorized, "token has been revoked")
	}

	return claims, nil
}

func (s *Service) blacklistWindowDays() int {
	window := (s.Config.JWTExpireSecs + (SEC_PER_DAY - 1)) / SEC_PER_DAY
	if window < 1 {
		return 1
	}
	return window + BLACKLIST_BUFFER
}

func (s *Service) RenewToken(ctx context.Context, refreshToken, deviceID string) (*response.RenewResult, error) {
	if refreshToken == "" {
		return nil, appError(http.StatusBadRequest, "refresh_token is required")
	}
	if deviceID == "" {
		return nil, appError(http.StatusBadRequest, "device_id is required")
	}

	data, err := s.Cache.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("service: get refresh token: %w", err)
	}
	if data == nil {
		return nil, appError(http.StatusUnauthorized, "refresh token not found or expired")
	}
	if data.DeviceID != deviceID {
		return nil, appError(http.StatusUnauthorized, "device mismatch")
	}

	accessToken, expiresAt, err := s.Signer.Sign(data.UserID, data.Permission, deviceID, data.TokenID)
	if err != nil {
		return nil, fmt.Errorf("service: sign renewed token: %w", err)
	}

	return &response.RenewResult{
		AccessToken: accessToken,
		ExpiresIn:   int(time.Until(expiresAt).Seconds()),
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *Service) RevokeToken(ctx context.Context, refreshToken, deviceID string) error {
	if refreshToken == "" {
		return appError(http.StatusBadRequest, "refresh_token is required")
	}

	data, err := s.Cache.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("service: get refresh token: %w", err)
	}
	if data == nil {
		return nil
	}
	if deviceID != "" && data.DeviceID != deviceID {
		return appError(http.StatusUnauthorized, "device mismatch")
	}

	accessTTL := time.Duration(s.Config.JWTExpireSecs) * time.Second
	_ = s.Cache.BlacklistToken(ctx, data.TokenID, accessTTL)
	_ = s.Repo.Delete(ctx, "refresh_token = ?", refreshToken)
	return s.Cache.DeleteRefreshToken(ctx, refreshToken)
}
