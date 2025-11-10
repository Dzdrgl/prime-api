package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotFound    = errors.New("user not found")
	ErrEmailExists = errors.New("email already exists")
)

type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error) // +++
	List(ctx context.Context, limit, offset int) ([]*User, error)
	UpdateAvatar(ctx context.Context, id uuid.UUID, url *string) error
}
