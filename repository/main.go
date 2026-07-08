package repository

import (
	"github.com/gauas/authorization-service/model"
	"github.com/gauas/authorization-service/service"
	"gorm.io/gorm"
)

type Registry struct {
	Token Repository[model.Token]
}

func New(db *gorm.DB) *Registry {
	return &Registry{
		Token: Repository[model.Token]{db: db},
	}
}

func (r *Registry) Tokens() service.Repository[model.Token] {
	return &r.Token
}
