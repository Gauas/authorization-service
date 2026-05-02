package jwt

import (
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID     uuid.UUID `json:"user_id"`
	Permission string    `json:"permission"`
	DeviceID   string    `json:"device_id"`
	TokenID    int64     `json:"token_id"`
	gojwt.RegisteredClaims
}

type Manager struct {
	secret    []byte
	expireSec int
}

func NewManager(secret string, expireSec int) *Manager {
	return &Manager{secret: []byte(secret), expireSec: expireSec}
}

func (m *Manager) Sign(userID uuid.UUID, permission, deviceID string, tokenID int64) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(time.Duration(m.expireSec) * time.Second)

	claims := Claims{
		UserID:     userID,
		Permission: permission,
		DeviceID:   deviceID,
		TokenID:    tokenID,
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(exp),
		},
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("jwt: sign: %w", err)
	}
	return signed, exp, nil
}

func (m *Manager) Verify(tokenStr string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("jwt: unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt: verify: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("jwt: invalid token")
	}
	return claims, nil
}
