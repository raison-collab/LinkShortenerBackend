package repository

import (
	"context"
	"time"

	"github.com/raison-collab/LinkShorternetBackend/internal/domain/entity"
)

// LinkRepository defines methods for link data access
type LinkRepository interface {
	// Create creates a new link
	Create(ctx context.Context, link *entity.Link) error

	// GetByShortCode retrieves a link by its short code
	GetByShortCode(ctx context.Context, shortCode string) (*entity.Link, error)

	// GetByID retrieves a link by its ID
	GetByID(ctx context.Context, id int64) (*entity.Link, error)

	// GetByUserID retrieves all links for a specific user
	GetByUserID(ctx context.Context, userID int64, offset, limit int) ([]*entity.Link, error)

	// Update updates an existing link
	Update(ctx context.Context, link *entity.Link) error

	// Delete deletes a link by ID
	Delete(ctx context.Context, id int64) error

	// IncrementClicks increments the click count for a link
	IncrementClicks(ctx context.Context, linkID int64) error

	// GetExpiredLinks retrieves all expired links
	GetExpiredLinks(ctx context.Context, before time.Time) ([]*entity.Link, error)

	// CountByUserID counts links for a specific user
	CountByUserID(ctx context.Context, userID int64) (int64, error)

	// ExistsByShortCode checks if a short code already exists
	ExistsByShortCode(ctx context.Context, shortCode string) (bool, error)
}

// LinkClickRepository defines methods for link click data access
type LinkClickRepository interface {
	// Create records a new click
	Create(ctx context.Context, click *entity.LinkClick) error

	// GetByLinkID retrieves all clicks for a specific link
	GetByLinkID(ctx context.Context, linkID int64, offset, limit int) ([]*entity.LinkClick, error)

	// GetStats retrieves statistics for a link
	GetStats(ctx context.Context, linkID int64, from, to time.Time) (*entity.LinkStats, error)

	// CountByLinkID counts clicks for a specific link
	CountByLinkID(ctx context.Context, linkID int64) (int64, error)

	// CountUniqueByLinkID counts unique clicks for a specific link
	CountUniqueByLinkID(ctx context.Context, linkID int64) (int64, error)
}
