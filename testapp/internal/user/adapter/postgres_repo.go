package adapter

import (
	"context"
	"fmt"
	"sync"

	"github.com/your-org/testapp/internal/user/domain"
	"github.com/your-org/testapp/internal/user/port"
)

type postgresRepo struct {
	mu    sync.RWMutex
	store map[string]*domain.User
}

// NewPostgresRepo creates an in-memory user repository implementing the driven port.
// Replace with a real PostgreSQL implementation in production.
func NewPostgresRepo() port.UserRepository {
	return &postgresRepo{store: make(map[string]*domain.User)}
}

func (r *postgresRepo) FindByID(_ context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.store[id]
	if !ok {
		return nil, fmt.Errorf("%w: id=%s", domain.ErrUserNotFound, id)
	}
	return u, nil
}

func (r *postgresRepo) Save(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[user.ID] = user
	return nil
}
