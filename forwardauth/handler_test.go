package forwardauth

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

type fakeRevocation struct {
	revoked bool
	err     error
}

func (f fakeRevocation) lookup(context.Context, string) (bool, error) { return f.revoked, f.err }

func TestHandlerTrustedHeadersAndDecisions(t *testing.T) {
	public, private, _ := ed25519.GenerateKey(rand.Reader)
	server := jwksServer(t, map[string]ed25519.PublicKey{"key-a": public}, nil)
	defer server.Close()
	raw := signToken(t, validClaims(), private, "key-a")

	tests := []struct {
		name       string
		auth       string
		revocation fakeRevocation
		wantStatus int
	}{
		{name: "valid", auth: "Bearer " + raw, wantStatus: http.StatusNoContent},
		{name: "missing authorization", wantStatus: http.StatusUnauthorized},
		{name: "revoked session", auth: "Bearer " + raw, revocation: fakeRevocation{revoked: true}, wantStatus: http.StatusUnauthorized},
		{name: "redis unavailable", auth: "Bearer " + raw, revocation: fakeRevocation{err: errors.New("redis unavailable")}, wantStatus: http.StatusServiceUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewHandler(testVerifier(server.URL), test.revocation.lookup, NewMetrics(prometheus.NewRegistry()))
			req := httptest.NewRequest(http.MethodGet, "/v1/authorization/forward-auth", nil)
			req.Header.Set("Authorization", test.auth)
			req.Header.Set(UserKeyHeader, "spoofed-user")
			req.Header.Set(DeviceIDHeader, "spoofed-device")
			req.Header.Set(SessionIDHeader, "spoofed-session")
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)
			if resp.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", resp.Code, test.wantStatus)
			}
			if test.wantStatus == http.StatusNoContent {
				if got := resp.Header().Get(UserKeyHeader); got != "user_xxx" {
					t.Fatalf("user header = %q", got)
				}
				if got := resp.Header().Get(DeviceIDHeader); got != "device_xxx" {
					t.Fatalf("device header = %q", got)
				}
				if got := resp.Header().Get(SessionIDHeader); got != "session_xxx" {
					t.Fatalf("session header = %q", got)
				}
			}
		})
	}
}

func TestHandlerJWKSUnavailable(t *testing.T) {
	_, private, _ := ed25519.GenerateKey(rand.Reader)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	handler := NewHandler(testVerifier(server.URL), fakeRevocation{}.lookup, NewMetrics(prometheus.NewRegistry()))
	request := httptest.NewRequest(http.MethodGet, "/v1/authorization/forward-auth", nil)
	request.Header.Set("Authorization", "Bearer "+signToken(t, validClaims(), private, "unknown"))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}
