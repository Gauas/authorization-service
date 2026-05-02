package infra

import (
	"github.com/gauas/authorization-service/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Infra struct {
	DB     *gorm.DB
	Memory *redis.Client
}

func New(cfg config.Config) *Infra {
	return &Infra{
		DB:     connectDatabase(cfg.DBUrl),
		Memory: connectMemory(cfg),
	}
}
