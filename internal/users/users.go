package users

import (
	"factual-docs/internal/services/database"
	"factual-docs/internal/services/redis"
)

type Service struct {
	Repo  *Repository
	redis redis.Service
}

func New(db database.Service, redis redis.Service) *Service {
	return &Service{
		Repo:  NewRepository(db),
		redis: redis,
	}
}
