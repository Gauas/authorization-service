package main

import (
	"context"
	"log"

	"github.com/gauas/authorization-service/app"
	"github.com/gauas/authorization-service/config"
	"github.com/gauas/authorization-service/controller"
	"github.com/gauas/authorization-service/grpc"
	"github.com/gauas/authorization-service/http"
	"github.com/gauas/authorization-service/infra"
	"github.com/gauas/authorization-service/middlewares"
	"github.com/gauas/authorization-service/packages/memory"
	"github.com/gauas/authorization-service/repository"
	"github.com/gauas/authorization-service/service"
)

func main() {
	cfg := config.New()

	infraInstance := infra.New(cfg)
	repo := repository.New(infraInstance.DB)

	mem := memory.New(infraInstance.Memory)
	mem.StartBlacklistGC(context.Background(), cfg.RefreshTTLDays)

	svc := service.New(repo, mem, cfg)
	ctrl := controller.New(svc)
	mw := middlewares.New(cfg)

	httpServer := http.Register(ctrl, mw, cfg)

	grpcServer := grpc.Register(cfg.GRPCPort, svc)

	if err := app.Start(httpServer, grpcServer); err != nil {
		log.Fatal(err)
	}
}
