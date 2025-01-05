package services

import (
	"context"
	"errors"

	"example.com/monolithic/internal/core/domain"
	"example.com/monolithic/internal/core/ports"
)

// Service errors
var (
	ErrInvalidInput   = errors.New("invalid input")
	ErrUserNotFound   = errors.New("user not found")
	ErrDuplicateEmail = errors.New("email already exists")
)

type UserService struct {
	repo ports.UserRepository
}

func NewUserService(repo ports.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, user *domain.User) error {
	// Validate input
	if err := s.validateUser(user); err != nil {
		return ErrInvalidInput
	}

	// Check for duplicate email
	exists, err := s.repo.ExistsByEmail(ctx, user.Email)
	if err != nil {
		return err
	}
	if exists {
		return ErrDuplicateEmail
	}

	// Create user
	return s.repo.Create(ctx, user)
}

func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *UserService) validateUser(user *domain.User) error {
	if user.Email == "" {
		return errors.New("email is required")
	}
	// Add more validation as needed
	return nil
}
