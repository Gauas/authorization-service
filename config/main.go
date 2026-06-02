package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	SecretKey      string
	JWTSecretKey   string
	JWTExpireSecs  int
	RefreshTTLDays int
	DBUrl          string
	CacheURL       string
}

func New() Config {
	return fromEnv()
}

func validate(cfg Config) {
	if cfg.Port == "" {
		log.Fatal("config: PORT is required")
	}
	if cfg.SecretKey == "" {
		log.Fatal("config: SECRET_KEY is required")
	}
	if cfg.JWTSecretKey == "" {
		log.Fatal("config: JWT_SECRET_KEY is required")
	}
}

func fromEnv() Config {
	_ = godotenv.Load()

	cfg := Config{
		Port:           getEnv("PORT", "8080"),
		SecretKey:      mustEnv("SECRET_KEY"),
		JWTSecretKey:   mustEnv("JWT_SECRET_KEY"),
		JWTExpireSecs:  getEnvInt("JWT_EXPIRE_SECS", 900),
		RefreshTTLDays: getEnvInt("REFRESH_TTL_DAYS", 30),
		DBUrl:          mustEnv("DB_URL"),
		CacheURL:       getEnv("CACHE_URL", "redis://localhost:6379/0"),
	}

	validate(cfg)
	return cfg
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("config: %s is required", key)
	}
	return v
}
