package forwardauth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const (
	UserKeyHeader   = "X-Gauas-User-Key"
	DeviceIDHeader  = "X-Gauas-Device-ID"
	SessionIDHeader = "X-Gauas-Session-ID"
)

type Handler struct {
	verifier *Verifier
	revoked  func(context.Context, string) (bool, error)
	metrics  *Metrics
}

func NewHandler(verifier *Verifier, revoked func(context.Context, string) (bool, error), metrics *Metrics) *Handler {
	return &Handler{verifier: verifier, revoked: revoked, metrics: metrics}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	identity, status, reason, err := h.authorize(r)

	duration := time.Since(started)
	h.metrics.Observe(status, reason, duration)
	if status >= 400 {
		logDecision(r, identity, status, reason, duration, err)
		writeError(w, status, reason)
		return
	}

	w.Header().Set(UserKeyHeader, identity.UserKey)
	w.Header().Set(DeviceIDHeader, identity.DeviceID)
	w.Header().Set(SessionIDHeader, identity.SessionID)
	w.WriteHeader(status)
}

func (h *Handler) authorize(r *http.Request) (Identity, int, string, error) {
	raw, err := bearerToken(r.Header.Get("Authorization"))
	if err != nil {
		return Identity{}, http.StatusUnauthorized, "missing_authorization", err
	}
	identity, err := h.verifier.Verify(r.Context(), raw)
	if err != nil {
		status, reason := verificationDecision(err)
		return Identity{}, status, reason, err
	}
	started := time.Now()
	revoked, err := h.revoked(r.Context(), identity.SessionID)
	h.metrics.ObserveRedisLookup(time.Since(started))
	if err != nil {
		return identity, http.StatusServiceUnavailable, "revocation_dependency", err
	}
	if revoked {
		return identity, http.StatusUnauthorized, "revoked_session", errors.New("session revoked")
	}
	return identity, http.StatusNoContent, "allowed", nil
}

func verificationDecision(err error) (int, string) {
	var dependency *DependencyError
	if errors.As(err, &dependency) {
		return http.StatusServiceUnavailable, "jwks_dependency"
	}
	return http.StatusUnauthorized, "invalid_token"
}

func bearerToken(value string) (string, error) {
	parts := strings.Fields(value)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", errors.New("bearer token is required")
	}
	return parts[1], nil
}

func writeError(w http.ResponseWriter, status int, reason string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": reason})
}

func logDecision(r *http.Request, identity Identity, status int, reason string, duration time.Duration, cause error) {
	fields := []any{
		"request_id", r.Header.Get("X-Request-ID"),
		"user_key", identity.UserKey,
		"session_id", identity.SessionID,
		"kid", identity.Kid,
		"decision", "deny",
		"reason", reason,
		"status", status,
		"latency_ms", duration.Milliseconds(),
	}
	if cause != nil {
		fields = append(fields, "error", cause.Error())
	}
	slog.Warn("authorization.decision", fields...)
}
