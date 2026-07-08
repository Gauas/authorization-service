package service

type Service struct {
	Repo   TokenRepository
	Cache  CacheStore
	Signer JWTSigner
	Config Config
}

func New(repo TokenRepository, cache CacheStore, signer JWTSigner, cfg Config) *Service {
	return &Service{Repo: repo, Cache: cache, Signer: signer, Config: cfg}
}
