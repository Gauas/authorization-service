package config

import (
	"encoding/json"
	"log"
	"os"

	"github.com/gauas/config-service/sdk"
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
	return fromSDK()
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
		JWTExpireSecs:  900,
		RefreshTTLDays: 30,
		DBUrl:          get("DB_URL"),
		CacheURL:       get("CACHE_URL"),
	}

	validate(cfg)
	return cfg, true
}

func fromSDK() Config {
	_ = godotenv.Load()

	secretKey := mustEnv("SECRET_KEY")

	client := sdk.New(sdk.Options{
		BaseURL:   mustEnv("CONFIG_SERVICE_URL"),
		SecretKey: mustEnv("CONFIG_SECRET_KEY"),
	})

	remote, err := client.Get("authorization-service", mustEnv("ENVIRONMENT"))
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	cfg := Config{
		Port:           remote.GetString("PORT", "8080"),
		SecretKey:      secretKey,
		JWTSecretKey:   remote.GetString("JWT_SECRET_KEY", ""),
		JWTExpireSecs:  int(remote.GetFloat64("JWT_EXPIRE_SECS", 900)),
		RefreshTTLDays: int(remote.GetFloat64("REFRESH_TTL_DAYS", 30)),
		DBUrl:          remote.GetString("DB_URL", ""),
		CacheURL:       remote.GetString("CACHE_URL", "redis://localhost:6379/0"),
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
