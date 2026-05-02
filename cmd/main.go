package main

import (
	"log"

	"github.com/gauas/authorization-service/config"
	"github.com/gauas/authorization-service/controller"
	"github.com/gauas/authorization-service/infra"
	"github.com/gauas/authorization-service/kernel"
	"github.com/gauas/authorization-service/middlewares"
	"github.com/gauas/authorization-service/packages/memory"
	"github.com/gauas/authorization-service/repository"
	"github.com/gauas/authorization-service/service"
)

func main() {
	Config := config.New()

	infraInstance := infra.New(Config)

	repositoryInstance := repository.New(infraInstance.DB)

	memoryInstance := memory.New(infraInstance.Memory)

	serviceInstance := service.New(repositoryInstance, memoryInstance, Config)

	controllerInstance := controller.New(serviceInstance)

	middlewareInstance := middlewares.New(Config)

	kernel.New(controllerInstance, middlewareInstance, Config).Start()

	log.Println("authorization-service started")
}
