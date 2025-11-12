package user

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

// PROTOKOL: Service arayüzü
type Service interface {
	List(ctx context.Context, limit, offset int) ([]*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	Create(ctx context.Context, email, fullName, passwordHash string) (*User, error)
	UpdateAvatarKey(ctx context.Context, id uuid.UUID, key *string) error
	DeleteAvatar(ctx context.Context, id uuid.UUID) error
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
	email = strings.ToLower(strings.TrimSpace(email))
	fullName = strings.TrimSpace(fullName)
	u := NewUser(email, fullName, passwordHash)
	if err := s.repo.Create(ctx, u); err != nil {
		if errors.Is(err, ErrEmailExists) {
			return nil, ErrEmailExists
		}
		return nil, err
	}
	return u, nil
}

func (s *service) UpdateAvatarKey(ctx context.Context, id uuid.UUID, key *string) error {
	return s.repo.UpdateAvatarKey(ctx, id, key)
}

func (s *service) DeleteAvatar(ctx context.Context, id uuid.UUID) error {
	return s.repo.UpdateAvatarKey(ctx, id, nil)
}
