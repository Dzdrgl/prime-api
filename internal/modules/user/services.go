package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// PROTOKOL: Service arayüzü
type Service interface {
	List(ctx context.Context, limit, offset int) ([]*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	Create(ctx context.Context, email, fullName, passwordHash string) (*User, error)
}

// Basit implementasyon (mock)
type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}
func (s *service) List(ctx context.Context, limit, offset int) ([]*User, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) Create(ctx context.Context, email, fullName, passwordHash string) (*User, error) {
	u := NewUser(email, fullName, passwordHash)
	if err := s.repo.Create(ctx, u); err != nil {
		if errors.Is(err, ErrEmailExists) {
			return nil, ErrEmailExists
		}
		return nil, err
	}
	return u, nil
}
