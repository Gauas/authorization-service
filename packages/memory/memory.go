package memory

import "github.com/redis/go-redis/v9"

type Store struct {
	client *redis.Client
}

func New(rdb *redis.Client) *Store {
	return &Store{client: rdb}
}
