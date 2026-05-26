package sdk

import "time"

type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresIn        int       `json:"expires_in"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type RenewResult struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type ValidateResult struct {
	UserID     string `json:"user_id"`
	Permission string `json:"permission"`
	DeviceID   string `json:"device_id"`
}

type apiResponse struct {
	Status int    `json:"status"`
	Error  string `json:"error,omitempty"`
}
