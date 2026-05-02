package service

import (
	"github.com/gauas/authorization-service/config"
	"github.com/gauas/authorization-service/packages/jwt"
	"github.com/gauas/authorization-service/packages/memory"
	"github.com/gauas/authorization-service/repository"
)

type Service struct {
	repo   *repository.Registry
	memory *memory.Store
	jwt    *jwt.Manager
	config config.Config
}

func New(repo *repository.Registry, mem *memory.Store, cfg config.Config) *Service {
	return &Service{
		repo:   repo,
		memory: mem,
		jwt:    jwt.NewManager(cfg.JWTSecretKey, cfg.JWTExpireSecs),
		config: cfg,
	}
}
