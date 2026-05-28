package memory

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const SCAN_COUNT int64 = 200

func (s *Store) CleanBlacklist(ctx context.Context, now time.Time, keepDays int) error {
	if keepDays < 1 {
		keepDays = 1
	}
	cutoff := now.UTC().AddDate(0, 0, -keepDays)
	pattern := fmt.Sprintf("%s:*", globalBlacklistKey())

	var cursor uint64
	for {
		keys, next, err := s.client.Scan(ctx, cursor, pattern, SCAN_COUNT).Result()
		if err != nil {
			return fmt.Errorf("memory: scan blacklist buckets: %w", err)
		}
		for _, key := range keys {
			parts := strings.Split(key, ":")
			if len(parts) < 3 {
				continue
			}
			day, err := time.Parse(DAY_FMT, parts[len(parts)-1])
			if err != nil {
				continue
			}
			if !day.Before(cutoff) {
				continue
			}
			if err := s.delBucket(ctx, key); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}

func (s *Store) delBucket(ctx context.Context, key string) error {
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("memory: delete expired blacklist bucket %s: %w", key, err)
	}
	return nil
}

func (s *Store) StartBlacklistGC(ctx context.Context, keepDays int) {
	go func() {
		_ = s.CleanBlacklist(ctx, time.Now().UTC(), keepDays)
		for {
			now := time.Now().UTC()
			next := now.Add(24 * time.Hour).Truncate(24 * time.Hour)
			timer := time.NewTimer(time.Until(next))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				_ = s.CleanBlacklist(ctx, time.Now().UTC(), keepDays)
			}
		}
	}()
}
