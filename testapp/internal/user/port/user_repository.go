package port

import (
	"context"

	"github.com/your-org/testapp/internal/user/domain"
)

// UserRepository is the driven port defining user persistence.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
	Save(ctx context.Context, user *domain.User) error
}
