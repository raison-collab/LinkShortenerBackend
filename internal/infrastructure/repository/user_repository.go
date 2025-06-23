package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/raison-collab/LinkShorternetBackend/internal/domain/entity"
	"github.com/raison-collab/LinkShorternetBackend/internal/domain/repository"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	return r.db.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user entity.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Пользователь не найден
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user entity.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Пользователь не найден
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, updated_at = $3
		WHERE id = $4
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
		user.UpdatedAt,
		user.ID,
	)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}

func (r *userRepository) GetStats(ctx context.Context, userID int64) (*entity.UserStats, error) {
	query := `
		SELECT 
			$1::BIGINT as user_id,
			(SELECT COUNT(*) FROM links WHERE user_id = $1) as total_links,
			(SELECT COALESCE(SUM(clicks), 0) FROM links WHERE user_id = $1) as total_clicks,
			(SELECT COUNT(*) FROM links WHERE user_id = $1) as active_links
	`

	var stats entity.UserStats
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.UserID,
		&stats.TotalLinks,
		&stats.TotalClicks,
		&stats.ActiveLinks,
	)

	if err != nil {
		return nil, err
	}

	return &stats, nil
}
