package forwardauth

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestRevocationEventHandlerDuplicateAndDelayed(t *testing.T) {
	now := time.Now()
	writes := map[string]string{}
	handler := &RevocationEventHandler{
		prefix: "revoked:sid:", timeout: time.Second,
		set: func(_ context.Context, key string, value any, ttl time.Duration) error {
			if ttl <= 0 {
				t.Fatalf("invalid TTL: %s", ttl)
			}
			writes[key] = value.(string)
			return nil
		},
	}
	event := SecurityEvent{EventID: "evt_xxx", Type: SessionRevokedEvent, SessionID: "session_xxx", ExpiresAt: now.Add(time.Minute).Unix()}
	payload, _ := json.Marshal(event)
	if err := handler.Handle(context.Background(), payload); err != nil {
		t.Fatal(err)
	}
	if err := handler.Handle(context.Background(), payload); err != nil {
		t.Fatal(err)
	}
	if len(writes) != 1 || writes["revoked:sid:session_xxx"] != "evt_xxx" {
		t.Fatalf("duplicate event was not idempotent: %#v", writes)
	}

	event.EventID, event.SessionID, event.ExpiresAt = "evt_old", "session_old", now.Add(-time.Second).Unix()
	payload, _ = json.Marshal(event)
	if err := handler.Handle(context.Background(), payload); err != nil {
		t.Fatal(err)
	}
	if _, found := writes["revoked:sid:session_old"]; found {
		t.Fatal("delayed expired event was stored")
	}
}

func TestRevocationEventHandlerRejectsMalformedEvent(t *testing.T) {
	handler := &RevocationEventHandler{}
	if err := handler.Handle(context.Background(), []byte("{")); !IsPermanentEventError(err) {
		t.Fatalf("error = %v, want permanent event error", err)
	}
}
