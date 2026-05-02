package service

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gauas/authorization-service/packages/response"
)

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func appError(code int, msg string) error {
	return response.NewError(code, msg)
}
