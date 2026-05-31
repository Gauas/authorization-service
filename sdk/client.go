package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gauas/authorization-service/supports"
	"github.com/google/uuid"
)

type Client struct {
	baseURL    string
	secretKey  string
	httpClient *http.Client
}

type Options struct {
	BaseURL   string
	SecretKey string
	Timeout   time.Duration
}

func New(opts Options) *Client {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		baseURL:    opts.BaseURL,
		secretKey:  opts.SecretKey,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) CreateToken(ctx context.Context, userID uuid.UUID, permission, deviceID string) (*TokenPair, error) {
	body, err := json.Marshal(map[string]interface{}{
		"user_id":    userID,
		"permission": permission,
	})
	if err != nil {
		return nil, fmt.Errorf("auth-sdk: marshal request: %w", err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, "/v1/authorization/token", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-ID", deviceID)

	var result struct {
		apiResponse
		Data *TokenPair `json:"data"`
	}
	if err := c.do(req, &result); err != nil {
		return nil, fmt.Errorf("auth-sdk: CreateToken: %w", err)
	}
	if result.Data == nil {
		return nil, fmt.Errorf("auth-sdk: CreateToken: empty response")
	}
	result.Data.ExpiresAt = time.Now().Add(time.Duration(result.Data.ExpiresIn) * time.Second)
	return result.Data, nil
}

func (c *Client) ValidateToken(ctx context.Context, token string) (*ValidateResult, error) {
	url := fmt.Sprintf("%s/v1/authorization/token/validate?token=%s", c.baseURL, token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("auth-sdk: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Secret-Key", c.secretKey)

	var result struct {
		apiResponse
		Data *ValidateResult `json:"data"`
	}
	if err := c.do(req, &result); err != nil {
		return nil, fmt.Errorf("auth-sdk: ValidateToken: %w", err)
	}
	return result.Data, nil
}

func (c *Client) RenewToken(ctx context.Context, refreshToken, deviceID string) (*RenewResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/authorization/token/renew", nil)
	if err != nil {
		return nil, fmt.Errorf("auth-sdk: create request: %w", err)
	}
	req.Header.Set("X-Refresh-Token", refreshToken)
	req.Header.Set("X-Device-ID", deviceID)

	var result struct {
		apiResponse
		Data *RenewResult `json:"data"`
	}
	if err := c.do(req, &result); err != nil {
		return nil, fmt.Errorf("auth-sdk: RenewToken: %w", err)
	}
	if result.Data == nil {
		return nil, fmt.Errorf("auth-sdk: RenewToken: empty response")
	}
	return result.Data, nil
}

func (c *Client) RevokeToken(ctx context.Context, refreshToken, deviceID string) error {
	req, err := c.newRequest(ctx, http.MethodDelete, "/v1/authorization/token", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Refresh-Token", refreshToken)
	req.Header.Set("X-Device-ID", deviceID)

	var result apiResponse
	if err := c.do(req, &result); err != nil {
		return fmt.Errorf("auth-sdk: RevokeToken: %w", err)
	}
	return nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("auth-sdk: create request: %w", err)
	}
	req.Header.Set("Secret-Key", c.secretKey)
	return req, nil
}

func (c *Client) do(req *http.Request, dest interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, supports.ReadBody(resp.Body))
	}
	if dest == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
