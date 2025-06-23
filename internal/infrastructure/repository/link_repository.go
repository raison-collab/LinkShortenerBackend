package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/raison-collab/LinkShorternetBackend/internal/domain/entity"
	"github.com/raison-collab/LinkShorternetBackend/internal/domain/repository"
)

type linkRepository struct {
	db *sql.DB
}

// NewLinkRepository создает новый репозиторий ссылок
func NewLinkRepository(db *sql.DB) repository.LinkRepository {
	return &linkRepository{db: db}
}

func (r *linkRepository) Create(ctx context.Context, link *entity.Link) error {
	query := `
		INSERT INTO links (short_code, original_url, user_id, clicks, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	now := time.Now()
	link.CreatedAt = now
	link.UpdatedAt = now

	return r.db.QueryRowContext(
		ctx,
		query,
		link.ShortCode,
		link.OriginalURL,
		link.UserID,
		link.Clicks,
		link.ExpiresAt,
		link.CreatedAt,
		link.UpdatedAt,
	).Scan(&link.ID)
}

func (r *linkRepository) GetByShortCode(ctx context.Context, shortCode string) (*entity.Link, error) {
	query := `
		SELECT id, short_code, original_url, user_id, clicks, expires_at, created_at, updated_at
		FROM links
		WHERE short_code = $1
	`

	var link entity.Link
	var userID sql.NullInt64
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(
		&link.ID,
		&link.ShortCode,
		&link.OriginalURL,
		&userID,
		&link.Clicks,
		&expiresAt,
		&link.CreatedAt,
		&link.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if userID.Valid {
		link.UserID = &userID.Int64
	}

	if expiresAt.Valid {
		link.ExpiresAt = &expiresAt.Time
	}

	return &link, nil
}

func (r *linkRepository) GetByID(ctx context.Context, id int64) (*entity.Link, error) {
	query := `
		SELECT id, short_code, original_url, user_id, clicks, expires_at, created_at, updated_at
		FROM links
		WHERE id = $1
	`

	var link entity.Link
	var userID sql.NullInt64
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&link.ID,
		&link.ShortCode,
		&link.OriginalURL,
		&userID,
		&link.Clicks,
		&expiresAt,
		&link.CreatedAt,
		&link.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if userID.Valid {
		link.UserID = &userID.Int64
	}

	if expiresAt.Valid {
		link.ExpiresAt = &expiresAt.Time
	}

	return &link, nil
}

func (r *linkRepository) GetByUserID(ctx context.Context, userID int64, offset, limit int) ([]*entity.Link, error) {
	query := `
		SELECT id, short_code, original_url, user_id, clicks, expires_at, created_at, updated_at
		FROM links
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]*entity.Link, 0)

	for rows.Next() {
		var link entity.Link
		var userIDNull sql.NullInt64
		var expiresAt sql.NullTime

		err := rows.Scan(
			&link.ID,
			&link.ShortCode,
			&link.OriginalURL,
			&userIDNull,
			&link.Clicks,
			&expiresAt,
			&link.CreatedAt,
			&link.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if userIDNull.Valid {
			link.UserID = &userIDNull.Int64
		}

		if expiresAt.Valid {
			link.ExpiresAt = &expiresAt.Time
		}

		links = append(links, &link)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func (r *linkRepository) Update(ctx context.Context, link *entity.Link) error {
	query := `
		UPDATE links
		SET original_url = $1, expires_at = $2, updated_at = $3
		WHERE id = $4
	`

	link.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		link.OriginalURL,
		link.ExpiresAt,
		link.UpdatedAt,
		link.ID,
	)

	return err
}

func (r *linkRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM links WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *linkRepository) IncrementClicks(ctx context.Context, linkID int64) error {
	query := `
		UPDATE links
		SET clicks = clicks + 1, updated_at = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), linkID)
	return err
}

func (r *linkRepository) GetExpiredLinks(ctx context.Context, before time.Time) ([]*entity.Link, error) {
	query := `
		SELECT id, short_code, original_url, user_id, clicks, expires_at, created_at, updated_at
		FROM links
		WHERE expires_at IS NOT NULL AND expires_at < $1
	`

	rows, err := r.db.QueryContext(ctx, query, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]*entity.Link, 0)

	for rows.Next() {
		var link entity.Link
		var userIDNull sql.NullInt64
		var expiresAt sql.NullTime

		err := rows.Scan(
			&link.ID,
			&link.ShortCode,
			&link.OriginalURL,
			&userIDNull,
			&link.Clicks,
			&expiresAt,
			&link.CreatedAt,
			&link.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if userIDNull.Valid {
			link.UserID = &userIDNull.Int64
		}

		if expiresAt.Valid {
			link.ExpiresAt = &expiresAt.Time
		}

		links = append(links, &link)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func (r *linkRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM links WHERE user_id = $1`

	var count int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *linkRepository) ExistsByShortCode(ctx context.Context, shortCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM links WHERE short_code = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
