package grpc

import (
	"github.com/gauas/authorization-service/grpc/runtime"
	"github.com/gauas/authorization-service/service"
)

type Server = runtime.Server

func Register(port string, service *service.Service) *Server {
	return runtime.Register(port, "authorization-service", func(server runtime.ServiceRegistrar) {
		// register gRPC service actions here
	})
}
