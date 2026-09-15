package forwardauth

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Address            string
	JWKSURL            string
	Issuer             string
	Audience           string
	RedisURL           string
	QueueURL           string
	SecurityQueue      string
	KafkaConsumerGroup string
	RevocationPrefix   string
	ClockSkew          time.Duration
	JWKSRefresh        time.Duration
	JWKSRequestTimeout time.Duration
	RedisLookupTimeout time.Duration
	ReadHeaderTimeout  time.Duration
	IdleTimeout        time.Duration
	ShutdownTimeout    time.Duration
}

func ConfigFromEnv() (Config, error) {
	cfg := Config{
		Address:            ":" + env("PORT", "8080"),
		JWKSURL:            strings.TrimSpace(os.Getenv("JWKS_URL")),
		Issuer:             env("JWT_ISSUER", "gauas-auth"),
		Audience:           env("JWT_AUDIENCE", "gauas-api"),
		RedisURL:           strings.TrimSpace(os.Getenv("REDIS_URL")),
		QueueURL:           strings.TrimSpace(os.Getenv("QUEUE_URL")),
		SecurityQueue:      env("SECURITY_EVENT_QUEUE", "auth.session.revoked"),
		KafkaConsumerGroup: env("KAFKA_CONSUMER_GROUP", "authorization-service"),
		RevocationPrefix:   env("REVOCATION_KEY_PREFIX", "revoked:sid:"),
		ClockSkew:          envDuration("JWT_CLOCK_SKEW", 30*time.Second),
		JWKSRefresh:        envDuration("JWKS_REFRESH_INTERVAL", 5*time.Minute),
		JWKSRequestTimeout: envDuration("JWKS_REQUEST_TIMEOUT", 2*time.Second),
		RedisLookupTimeout: envDuration("REDIS_LOOKUP_TIMEOUT", 100*time.Millisecond),
		ReadHeaderTimeout:  envDuration("HTTP_READ_HEADER_TIMEOUT", 2*time.Second),
		IdleTimeout:        envDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout:    envDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
	}

	if cfg.RedisURL == "" {
		return Config{}, fmt.Errorf("REDIS_URL is required")
	}
	if cfg.QueueURL == "" {
		return Config{}, fmt.Errorf("QUEUE_URL is required")
	}
	if cfg.Issuer == "" || cfg.Audience == "" {
		return Config{}, fmt.Errorf("JWT_ISSUER and JWT_AUDIENCE are required")
	}
	if cfg.JWKSURL == "" {
		return Config{}, fmt.Errorf("JWKS_URL is required")
	}
	if cfg.ClockSkew < 0 || cfg.JWKSRefresh <= 0 || cfg.JWKSRequestTimeout <= 0 || cfg.RedisLookupTimeout <= 0 {
		return Config{}, fmt.Errorf("timeouts and refresh intervals must be positive")
	}
	return cfg, nil
}

func env(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func envDuration(name string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
