package account

import (
	"context"
)

type Repository interface {
	GetByOwnerID(ctx context.Context, ownerID string) (Account, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) *service {
	return &service{repo: repo}
}

func (s *service) GetByOwnerID(ctx context.Context, ownerID string) (Account, error) {
	return s.repo.GetByOwnerID(ctx, ownerID)
}
