package port

import (
	"context"

	"github.com/your-org/testapp/internal/user/domain"
)

// UserService is the driving port defining user-related use cases.
type UserService interface {
	GetUser(ctx context.Context, id string) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) error
}
