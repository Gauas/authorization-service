package config

import (
	"encoding/json"
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
	if cfg, ok := fromFile(); ok {
		return cfg
	}
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

const localConfigPath = ".config/config.json"

func fromFile() (Config, bool) {
	data, err := os.ReadFile(localConfigPath)
	if err != nil {
		return Config{}, false
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Fatalf("config: invalid %s: %v", localConfigPath, err)
	}

	get := func(key string) string {
		v := raw[key]
		if v == "" {
			log.Fatalf("config: %s is required in %s", key, localConfigPath)
		}
		return v
	}

	cfg := Config{
		Port:           get("PORT"),
		SecretKey:      get("SECRET_KEY"),
		JWTSecretKey:   get("JWT_SECRET_KEY"),
		JWTExpireSecs:  getEnvInt("JWT_EXPIRE_SECS", 900),
		RefreshTTLDays: getEnvInt("REFRESH_TTL_DAYS", 30),
		DBUrl:          get("DB_URL"),
		CacheURL:       get("CACHE_URL"),
	}

	validate(cfg)
	return cfg, true
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
