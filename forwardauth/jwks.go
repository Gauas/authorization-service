package forwardauth

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type jwk struct {
	KeyID     string `json:"kid"`
	KeyType   string `json:"kty"`
	Curve     string `json:"crv"`
	Algorithm string `json:"alg"`
	X         string `json:"x"`
}

type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

type DependencyError struct{ Err error }

func (e *DependencyError) Error() string { return e.Err.Error() }
func (e *DependencyError) Unwrap() error { return e.Err }

type keySet struct {
	keys map[string]ed25519.PublicKey
}

type JWKSCache struct {
	url     string
	client  *http.Client
	keys    atomic.Pointer[keySet]
	refresh sync.Mutex
	observe func(bool)
}

func (c *JWKSCache) SetMetrics(metrics *Metrics) { c.observe = metrics.ObserveJWKSRefresh }

func NewJWKSCache(url string, timeout time.Duration) *JWKSCache {
	cache := &JWKSCache{url: url, client: &http.Client{Timeout: timeout}, observe: func(bool) {}}
	cache.keys.Store(&keySet{keys: map[string]ed25519.PublicKey{}})
	return cache
}

func (c *JWKSCache) Key(ctx context.Context, kid string) (ed25519.PublicKey, error) {
	if key, ok := c.load(kid); ok {
		return key, nil
	}
	if c.url == "" {
		return nil, fmt.Errorf("unknown kid %q", kid)
	}

	c.refresh.Lock()
	defer c.refresh.Unlock()
	if key, ok := c.load(kid); ok {
		return key, nil
	}
	if err := c.refreshLocked(ctx); err != nil {
		return nil, &DependencyError{Err: err}
	}
	key, ok := c.load(kid)
	if !ok {
		return nil, fmt.Errorf("unknown kid %q", kid)
	}
	return key, nil
}

func (c *JWKSCache) Refresh(ctx context.Context) error {
	if c.url == "" {
		return nil
	}
	c.refresh.Lock()
	defer c.refresh.Unlock()
	return c.refreshLocked(ctx)
}

func (c *JWKSCache) Run(ctx context.Context, interval time.Duration) {
	if c.url == "" {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = c.Refresh(ctx)
		}
	}
}

func (c *JWKSCache) load(kid string) (ed25519.PublicKey, bool) {
	key, ok := c.keys.Load().keys[kid]
	return key, ok
}

func (c *JWKSCache) refreshLocked(ctx context.Context) error {
	success := false
	defer func() { c.observe(success) }()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return fmt.Errorf("create JWKS request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch JWKS: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch JWKS: status %d", resp.StatusCode)
	}
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	var document jwksDocument
	if err := decoder.Decode(&document); err != nil {
		return fmt.Errorf("decode JWKS: %w", err)
	}

	keys := make(map[string]ed25519.PublicKey, len(document.Keys))
	for _, candidate := range document.Keys {
		if candidate.KeyID == "" || candidate.KeyType != "OKP" || candidate.Curve != "Ed25519" || candidate.Algorithm != "EdDSA" {
			continue
		}
		raw, err := base64.RawURLEncoding.DecodeString(candidate.X)
		if err != nil || len(raw) != ed25519.PublicKeySize {
			continue
		}
		keys[candidate.KeyID] = ed25519.PublicKey(append([]byte(nil), raw...))
	}
	if len(keys) == 0 {
		return fmt.Errorf("JWKS contains no usable Ed25519 keys")
	}
	c.keys.Store(&keySet{keys: keys})
	success = true
	return nil
}
