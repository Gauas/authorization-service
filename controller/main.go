package controller

import "github.com/gauas/authorization-service/service"

type Controller struct {
	service *service.Service
}

func New(svc *service.Service) *Controller {
	return &Controller{service: svc}
}
