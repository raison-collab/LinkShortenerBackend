package repository

import (
	"context"

	"github.com/raison-collab/LinkShorternetBackend/internal/domain/entity"
)

// UserRepository defines methods for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entity.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id int64) (*entity.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// Update updates user information
	Update(ctx context.Context, user *entity.User) error

	// Delete deletes a user
	Delete(ctx context.Context, id int64) error

	// ExistsByEmail checks if email already exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// GetStats retrieves user statistics
	GetStats(ctx context.Context, userID int64) (*entity.UserStats, error)
}
