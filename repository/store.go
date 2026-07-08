package repository

import "github.com/gauas/authorization-service/service"

func TokenRepository(registry *Registry) service.TokenRepository {
	return &registry.Token
}
