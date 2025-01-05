package ports

import (
	"context"
	"errors"

	"example.com/monolithic/internal/core/domain"
)

var ErrNotFound = errors.New("not found")
var ErrDuplicateEmail = errors.New("Duplicate email")

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}
