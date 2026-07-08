package service

import (
	"context"
	"time"

	"github.com/gauas/authorization-service/model"
	"github.com/gauas/authorization-service/packages/jwt"
	"github.com/gauas/authorization-service/packages/memory"
)

type Config struct {
	JWTExpireSecs  int
	RefreshTTLDays int
}

type TokenRepository interface {
	Create(ctx context.Context, entity *model.Token) (*model.Token, error)
	Delete(ctx context.Context, args ...interface{}) error
}

type CacheStore interface {
	StoreRefreshToken(ctx context.Context, token string, data memory.RefreshTokenData, ttl time.Duration) error
	GetRefreshToken(ctx context.Context, token string) (*memory.RefreshTokenData, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	IsTokenBlacklisted(ctx context.Context, tokenID int64, windowDays int) (bool, error)
	BlacklistToken(ctx context.Context, tokenID int64, ttl time.Duration) error
	TrackTokenForDevice(ctx context.Context, userID int64, deviceID, token string, ttl time.Duration) error
}

type JWTSigner interface {
	Sign(userID int64, permission, deviceID string, tokenID int64) (string, time.Time, error)
	Verify(token string) (*jwt.Claims, error)
}
