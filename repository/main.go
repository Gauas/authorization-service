package repository

import (
	"github.com/gauas/authorization-service/model"
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
