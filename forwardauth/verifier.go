package forwardauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	SessionID string `json:"sid"`
	DeviceID  string `json:"device_id"`
	gojwt.RegisteredClaims
}

type Identity struct {
	UserKey   string
	DeviceID  string
	SessionID string
	Kid       string
}

type Verifier struct {
	keys     *JWKSCache
	issuer   string
	audience string
	leeway   time.Duration
}

func NewVerifier(keys *JWKSCache, cfg Config) *Verifier {
	return &Verifier{
		keys: keys, issuer: cfg.Issuer, audience: cfg.Audience,
		leeway: cfg.ClockSkew,
	}
}

func (v *Verifier) Verify(ctx context.Context, raw string) (Identity, error) {
	header, err := tokenHeader(raw)
	if err != nil {
		return Identity{}, err
	}

	claims := new(Claims)
	options := []gojwt.ParserOption{
		gojwt.WithIssuer(v.issuer), gojwt.WithAudience(v.audience),
		gojwt.WithExpirationRequired(), gojwt.WithIssuedAt(), gojwt.WithLeeway(v.leeway),
	}

	if header.Algorithm != gojwt.SigningMethodEdDSA.Alg() || strings.TrimSpace(header.KeyID) == "" {
		return Identity{}, fmt.Errorf("EdDSA token with kid is required")
	}
	key, err := v.keys.Key(ctx, header.KeyID)
	if err != nil {
		return Identity{}, err
	}
	options = append(options, gojwt.WithValidMethods([]string{gojwt.SigningMethodEdDSA.Alg()}))

	parsed, err := gojwt.ParseWithClaims(raw, claims, func(*gojwt.Token) (any, error) { return key, nil }, options...)
	if err != nil {
		return Identity{}, fmt.Errorf("verify JWT: %w", err)
	}
	if !parsed.Valid {
		return Identity{}, fmt.Errorf("verify JWT: invalid token")
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return Identity{}, fmt.Errorf("JWT subject is required")
	}

	sid := strings.TrimSpace(claims.SessionID)
	if sid == "" {
		return Identity{}, fmt.Errorf("JWT session ID is required")
	}
	if strings.TrimSpace(claims.ID) == "" || strings.TrimSpace(claims.DeviceID) == "" {
		return Identity{}, fmt.Errorf("JWT jti and device_id are required")
	}

	deviceID := strings.TrimSpace(claims.DeviceID)
	return Identity{
		UserKey: claims.Subject, DeviceID: deviceID, SessionID: sid,
		Kid: header.KeyID,
	}, nil
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
}

func tokenHeader(raw string) (jwtHeader, error) {
	dot := strings.IndexByte(raw, '.')
	if dot <= 0 || dot > 4096 {
		return jwtHeader{}, fmt.Errorf("parse JWT header: malformed token")
	}
	encoded := raw[:dot]
	buffer := make([]byte, base64.RawURLEncoding.DecodedLen(len(encoded)))
	n, err := base64.RawURLEncoding.Decode(buffer, []byte(encoded))
	if err != nil {
		return jwtHeader{}, fmt.Errorf("parse JWT header: %w", err)
	}
	var header jwtHeader
	if err := json.Unmarshal(buffer[:n], &header); err != nil {
		return jwtHeader{}, fmt.Errorf("parse JWT header: %w", err)
	}
	return header, nil
}
