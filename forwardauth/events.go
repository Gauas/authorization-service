package forwardauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const SessionRevokedEvent = "auth.session.revoked"

type SecurityEvent struct {
	EventID   string `json:"event_id"`
	Type      string `json:"type"`
	SessionID string `json:"session_id"`
	ExpiresAt int64  `json:"expires_at"`
}

type PermanentEventError struct{ Err error }

func (e *PermanentEventError) Error() string { return e.Err.Error() }
func (e *PermanentEventError) Unwrap() error { return e.Err }

func IsPermanentEventError(err error) bool {
	var permanent *PermanentEventError
	return errors.As(err, &permanent)
}

type RevocationConsumer interface {
	Run(context.Context, func(context.Context, []byte) error) error
	Close() error
}

type RevocationEventHandler struct {
	set     func(context.Context, string, any, time.Duration) error
	prefix  string
	timeout time.Duration
}

func NewRevocationEventHandler(client *redis.Client, prefix string, timeout time.Duration) *RevocationEventHandler {
	return &RevocationEventHandler{
		set: func(ctx context.Context, key string, value any, ttl time.Duration) error {
			return client.Set(ctx, key, value, ttl).Err()
		},
		prefix: prefix, timeout: timeout,
	}
}

func (h *RevocationEventHandler) Handle(ctx context.Context, payload []byte) error {
	var event SecurityEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return &PermanentEventError{Err: fmt.Errorf("decode security event: %w", err)}
	}
	if event.Type != SessionRevokedEvent || strings.TrimSpace(event.EventID) == "" || strings.TrimSpace(event.SessionID) == "" {
		return &PermanentEventError{Err: fmt.Errorf("invalid session revocation event")}
	}
	ttl := time.Until(time.Unix(event.ExpiresAt, 0))
	if ttl <= 0 {
		return nil
	}
	write, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	if err := h.set(write, h.prefix+event.SessionID, event.EventID, ttl); err != nil {
		return fmt.Errorf("store session revocation: %w", err)
	}
	return nil
}
