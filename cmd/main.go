package main

import (
	"context"
	"log"

	"github.com/gauas/authorization-service/config"
	"github.com/gauas/authorization-service/controller"
	"github.com/gauas/authorization-service/infra"
	"github.com/gauas/authorization-service/kernel"
	"github.com/gauas/authorization-service/middlewares"
	"github.com/gauas/authorization-service/packages/jwt"
	"github.com/gauas/authorization-service/packages/memory"
	"github.com/gauas/authorization-service/repository"
	"github.com/gauas/authorization-service/service"
)

func main() {
	cfg := config.New()

	infraSet := infra.New(cfg)
	repo := repository.New(infraSet.DB)
	cache := memory.New(infraSet.Memory)
	cache.StartBlacklistGC(context.Background(), cfg.RefreshTTLDays)

	svc := service.New(
		repository.TokenRepository(repo),
		cache,
		jwt.NewManager(cfg.JWTSecretKey, cfg.JWTExpireSecs),
		service.Config{JWTExpireSecs: cfg.JWTExpireSecs, RefreshTTLDays: cfg.RefreshTTLDays},
	)
	ctrl := controller.New(svc)
	mw := middlewares.New(cfg)

	kernel.New(ctrl, mw, cfg).Start()

	log.Println("authorization-service started")
}
