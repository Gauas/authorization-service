package forwardauth

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

func TestVerifier(t *testing.T) {
	public, private, _ := ed25519.GenerateKey(rand.Reader)
	server := jwksServer(t, map[string]ed25519.PublicKey{"key-a": public}, nil)
	defer server.Close()
	verifier := testVerifier(server.URL)

	tests := []struct {
		name   string
		mutate func(*Claims)
		key    ed25519.PrivateKey
		kid    string
		wantOK bool
	}{
		{name: "valid", key: private, kid: "key-a", wantOK: true},
		{name: "invalid signature", key: mustPrivateKey(t), kid: "key-a"},
		{name: "expired", key: private, kid: "key-a", mutate: func(c *Claims) { c.ExpiresAt = gojwt.NewNumericDate(time.Now().Add(-time.Hour)) }},
		{name: "wrong issuer", key: private, kid: "key-a", mutate: func(c *Claims) { c.Issuer = "other" }},
		{name: "wrong audience", key: private, kid: "key-a", mutate: func(c *Claims) { c.Audience = gojwt.ClaimStrings{"other"} }},
		{name: "missing kid", key: private},
		{name: "unknown kid", key: private, kid: "key-b"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims := validClaims()
			if test.mutate != nil {
				test.mutate(&claims)
			}
			raw := signToken(t, claims, test.key, test.kid)
			identity, err := verifier.Verify(context.Background(), raw)
			if test.wantOK && err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			if !test.wantOK && err == nil {
				t.Fatal("Verify() unexpectedly succeeded")
			}
			if test.wantOK && (identity.UserKey != "user_xxx" || identity.SessionID != "session_xxx") {
				t.Fatalf("unexpected identity: %#v", identity)
			}
		})
	}

	if _, err := verifier.Verify(context.Background(), "not-a-jwt"); err == nil {
		t.Fatal("malformed JWT unexpectedly succeeded")
	}
}

func TestVerifierRejectsHS256(t *testing.T) {
	verifier := testVerifier("http://unused")
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, validClaims())
	raw, err := token.SignedString([]byte("shared-secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := verifier.Verify(context.Background(), raw); err == nil {
		t.Fatal("HS256 token was accepted")
	}
}

func TestJWKSUnknownKidRefreshAndRotation(t *testing.T) {
	publicA, privateA, _ := ed25519.GenerateKey(rand.Reader)
	publicB, privateB, _ := ed25519.GenerateKey(rand.Reader)
	var requests atomic.Int64
	var mutex sync.RWMutex
	keys := map[string]ed25519.PublicKey{"key-a": publicA}
	server := jwksServer(t, nil, func() map[string]ed25519.PublicKey {
		requests.Add(1)
		mutex.RLock()
		defer mutex.RUnlock()
		copy := make(map[string]ed25519.PublicKey, len(keys))
		for kid, key := range keys {
			copy[kid] = key
		}
		return copy
	})
	defer server.Close()

	verifier := testVerifier(server.URL)
	if _, err := verifier.Verify(context.Background(), signToken(t, validClaims(), privateA, "key-a")); err != nil {
		t.Fatal(err)
	}
	mutex.Lock()
	keys["key-b"] = publicB
	mutex.Unlock()
	before := requests.Load()

	rawB := signToken(t, validClaims(), privateB, "key-b")
	var group sync.WaitGroup
	for range 20 {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := verifier.Verify(context.Background(), rawB); err != nil {
				t.Error(err)
			}
		}()
	}
	group.Wait()
	if delta := requests.Load() - before; delta != 1 {
		t.Fatalf("unknown kid caused %d JWKS refreshes, want 1", delta)
	}
	if _, err := verifier.Verify(context.Background(), signToken(t, validClaims(), privateA, "key-a")); err != nil {
		t.Fatalf("old active token failed during rotation: %v", err)
	}
}

func BenchmarkVerifierCachedKey(b *testing.B) {
	public, private, _ := ed25519.GenerateKey(rand.Reader)
	server := jwksServer(b, map[string]ed25519.PublicKey{"key-a": public}, nil)
	defer server.Close()
	verifier := testVerifier(server.URL)
	raw := signToken(b, validClaims(), private, "key-a")
	if _, err := verifier.Verify(context.Background(), raw); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := verifier.Verify(context.Background(), raw); err != nil {
			b.Fatal(err)
		}
	}
}

func testVerifier(url string) *Verifier {
	cfg := Config{JWKSURL: url, Issuer: "gauas-auth", Audience: "gauas-api", ClockSkew: time.Second}
	return NewVerifier(NewJWKSCache(url, time.Second), cfg)
}

func validClaims() Claims {
	now := time.Now()
	return Claims{
		SessionID: "session_xxx", DeviceID: "device_xxx",
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer: "gauas-auth", Audience: gojwt.ClaimStrings{"gauas-api"}, Subject: "user_xxx", ID: "token_xxx",
			IssuedAt: gojwt.NewNumericDate(now), ExpiresAt: gojwt.NewNumericDate(now.Add(time.Minute)),
		},
	}
}

func signToken(t testing.TB, claims Claims, private ed25519.PrivateKey, kid string) string {
	t.Helper()
	token := gojwt.NewWithClaims(gojwt.SigningMethodEdDSA, claims)
	if kid != "" {
		token.Header["kid"] = kid
	}
	raw, err := token.SignedString(private)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func mustPrivateKey(t testing.TB) ed25519.PrivateKey {
	t.Helper()
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return private
}

func jwksServer(t testing.TB, fixed map[string]ed25519.PublicKey, dynamic func() map[string]ed25519.PublicKey) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		keys := fixed
		if dynamic != nil {
			keys = dynamic()
		}
		document := jwksDocument{Keys: make([]jwk, 0, len(keys))}
		for kid, key := range keys {
			document.Keys = append(document.Keys, jwk{KeyID: kid, KeyType: "OKP", Curve: "Ed25519", Algorithm: "EdDSA", X: base64.RawURLEncoding.EncodeToString(key)})
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(document); err != nil {
			t.Error(err)
		}
	}))
}
