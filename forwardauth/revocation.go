package forwardauth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRevocations struct {
	client  *redis.Client
	prefix  string
	timeout time.Duration
}

func NewRedisRevocations(client *redis.Client, prefix string, timeout time.Duration) *RedisRevocations {
	return &RedisRevocations{client: client, prefix: prefix, timeout: timeout}
}

func (r *RedisRevocations) Revoked(ctx context.Context, sessionID string) (bool, error) {
	lookup, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	exists, err := r.client.Exists(lookup, r.prefix+sessionID).Result()
	if err != nil {
		return false, fmt.Errorf("lookup session revocation: %w", err)
	}
	return exists > 0, nil
}
