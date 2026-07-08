package service

import "github.com/gauas/authorization-service/model"

type Service struct {
	Repo   Repository[model.Token]
	Cache  CacheStore
	Signer JWTSigner
	Config Config
}

func New(repo Repository[model.Token], cache CacheStore, signer JWTSigner, cfg Config) *Service {
	return &Service{Repo: repo, Cache: cache, Signer: signer, Config: cfg}
}
