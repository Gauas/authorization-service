package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gauas/authorization-service/packages/bitmap"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RefreshTokenData struct {
	UserID     uuid.UUID `json:"user_id"`
	DeviceID   string    `json:"device_id"`
	Permission string    `json:"permission"`
	TokenID    int64     `json:"token_id"`
}

func (s *Store) StoreRefreshToken(ctx context.Context, token string, data RefreshTokenData, ttl time.Duration) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("memory: marshal refresh token: %w", err)
	}

	return s.client.Set(ctx, refreshKey(token), raw, ttl).Err()
}

func (s *Store) GetRefreshToken(ctx context.Context, token string) (*RefreshTokenData, error) {
	raw, err := s.client.Get(ctx, refreshKey(token)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("memory: get refresh token: %w", err)
	}

	var data RefreshTokenData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("memory: unmarshal refresh token: %w", err)
	}

	return &data, nil
}

func (s *Store) DeleteRefreshToken(ctx context.Context, token string) error {
	return s.client.Del(ctx, refreshKey(token)).Err()
}

func (s *Store) TrackTokenForDevice(ctx context.Context, userID uuid.UUID, deviceID, token string, ttl time.Duration) error {
	indexKey := deviceIndexKey(userID, deviceID)

	pipe := s.client.Pipeline()
	pipe.SAdd(ctx, indexKey, token)
	pipe.Expire(ctx, indexKey, ttl)

	_, err := pipe.Exec(ctx)

	return err
}

func (s *Store) BlacklistToken(ctx context.Context, tokenID int64, ttl time.Duration) error {
	offset, ok := bitmap.Offset(tokenID)
	if !ok {
		return fmt.Errorf("memory: invalid token id")
	}
	_ = ttl

	pipe := s.client.Pipeline()
	pipe.SetBit(ctx, blacklistBucketKey(time.Now().UTC()), offset, 1)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("memory: blacklist token: %w", err)
	}
	return nil
}

func (s *Store) IsTokenBlacklisted(ctx context.Context, tokenID int64, windowDays int) (bool, error) {
	offset, ok := bitmap.Offset(tokenID)
	if !ok {
		return false, nil
	}
	if windowDays < 1 {
		windowDays = 1
	}
	now := time.Now().UTC()
	for i := 0; i < windowDays; i++ {
		day := now.AddDate(0, 0, -i)
		bit, err := s.client.GetBit(ctx, blacklistBucketKey(day), offset).Result()
		if err != nil {
			return false, fmt.Errorf("memory: check blacklist: %w", err)
		}
		if bit == 1 {
			return true, nil
		}
	}
	return false, nil
}
