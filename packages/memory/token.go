package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RefreshTokenData struct {
	UserID     uuid.UUID `json:"user_id"`
	DeviceID   string    `json:"device_id"`
	Permission string    `json:"permission"`
	IssuedAt   time.Time `json:"issued_at"`
	TokenID    int64     `json:"token_id"`
}

func (s *Store) NextTokenID(ctx context.Context, userID uuid.UUID) (int64, error) {
	id, err := s.client.Incr(ctx, tokenSeqKey(userID)).Result()
	if err != nil {
		return 0, fmt.Errorf("memory: next token id: %w", err)
	}
	return id, nil
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

func (s *Store) BlacklistToken(ctx context.Context, userID uuid.UUID, tokenID int64, ttl time.Duration) error {
	key := blacklistKey(userID)

	pipe := s.client.Pipeline()
	pipe.SetBit(ctx, key, tokenID, 1)
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("memory: blacklist token: %w", err)
	}
	return nil
}

func (s *Store) IsTokenBlacklisted(ctx context.Context, userID uuid.UUID, tokenID int64) (bool, error) {
	bit, err := s.client.GetBit(ctx, blacklistKey(userID), tokenID).Result()
	if err != nil {
		return false, fmt.Errorf("memory: check blacklist: %w", err)
	}
	return bit == 1, nil
}
