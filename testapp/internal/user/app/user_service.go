package app

import (
	"context"
	"fmt"

	"github.com/your-org/testapp/internal/user/domain"
	"github.com/your-org/testapp/internal/user/port"
)

type userService struct {
	repo port.UserRepository
}

// NewUserService creates a UserService that implements the driving port.
func NewUserService(repo port.UserRepository) port.UserService {
	return &userService{repo: repo}
}

func (s *userService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

func (s *userService) CreateUser(ctx context.Context, user *domain.User) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	if err := s.repo.Save(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}
