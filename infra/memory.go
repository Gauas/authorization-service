package infra

import (
	"context"
	"log"

	"github.com/gauas/authorization-service/config"
	"github.com/redis/go-redis/v9"
)

func connectMemory(cfg config.Config) *redis.Client {
	opts, err := redis.ParseURL(cfg.CacheURL)
	if err != nil {
		log.Fatalf("infra: invalid cache URL: %v", err)
	}

	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("infra: failed to connect to memory store: %v", err)
	}

	log.Println("infra: memory store connected")

	return client
}
