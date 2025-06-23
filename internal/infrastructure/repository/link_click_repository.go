package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/raison-collab/LinkShorternetBackend/internal/domain/entity"
	"github.com/raison-collab/LinkShorternetBackend/internal/domain/repository"
)

type linkClickRepository struct {
	db *sql.DB
}

// NewLinkClickRepository создает новый репозиторий кликов
func NewLinkClickRepository(db *sql.DB) repository.LinkClickRepository {
	return &linkClickRepository{db: db}
}

func (r *linkClickRepository) Create(ctx context.Context, click *entity.LinkClick) error {
	query := `
		INSERT INTO link_clicks (link_id, ip_address, user_agent, referer, country, city, clicked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	if click.ClickedAt.IsZero() {
		click.ClickedAt = time.Now()
	}

	return r.db.QueryRowContext(
		ctx,
		query,
		click.LinkID,
		click.IPAddress,
		click.UserAgent,
		click.Referer,
		click.Country,
		click.City,
		click.ClickedAt,
	).Scan(&click.ID)
}

func (r *linkClickRepository) GetByLinkID(ctx context.Context, linkID int64, offset, limit int) ([]*entity.LinkClick, error) {
	query := `
		SELECT id, link_id, ip_address, user_agent, referer, country, city, clicked_at
		FROM link_clicks
		WHERE link_id = $1
		ORDER BY clicked_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, linkID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clicks := make([]*entity.LinkClick, 0)

	for rows.Next() {
		var click entity.LinkClick
		var referer, country, city sql.NullString

		err := rows.Scan(
			&click.ID,
			&click.LinkID,
			&click.IPAddress,
			&click.UserAgent,
			&referer,
			&country,
			&city,
			&click.ClickedAt,
		)

		if err != nil {
			return nil, err
		}

		if referer.Valid {
			click.Referer = referer.String
		}

		if country.Valid {
			click.Country = country.String
		}

		if city.Valid {
			click.City = city.String
		}

		clicks = append(clicks, &click)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clicks, nil
}

func (r *linkClickRepository) GetStats(ctx context.Context, linkID int64, from, to time.Time) (*entity.LinkStats, error) {
	// Подготавливаем статистику
	stats := &entity.LinkStats{
		LinkID:          linkID,
		ClicksByDate:    make(map[string]int64),
		ClicksByCountry: make(map[string]int64),
		ClicksByDevice:  make(map[string]int64),
		TopReferers:     []entity.RefererStats{},
	}

	// Получаем общее количество кликов
	totalQuery := `
		SELECT COUNT(*) FROM link_clicks
		WHERE link_id = $1 AND clicked_at BETWEEN $2 AND $3
	`
	err := r.db.QueryRowContext(ctx, totalQuery, linkID, from, to).Scan(&stats.TotalClicks)
	if err != nil {
		return nil, err
	}

	// Получаем количество уникальных кликов (по IP)
	uniqueQuery := `
		SELECT COUNT(DISTINCT ip_address) FROM link_clicks
		WHERE link_id = $1 AND clicked_at BETWEEN $2 AND $3
	`
	err = r.db.QueryRowContext(ctx, uniqueQuery, linkID, from, to).Scan(&stats.UniqueClicks)
	if err != nil {
		return nil, err
	}

	// Получаем статистику по дате
	dateQuery := `
		SELECT TO_CHAR(clicked_at, 'YYYY-MM-DD') as click_date, COUNT(*) as count
		FROM link_clicks
		WHERE link_id = $1 AND clicked_at BETWEEN $2 AND $3
		GROUP BY click_date
		ORDER BY click_date
	`
	dateRows, err := r.db.QueryContext(ctx, dateQuery, linkID, from, to)
	if err != nil {
		return nil, err
	}
	defer dateRows.Close()

	for dateRows.Next() {
		var date string
		var count int64
		if err := dateRows.Scan(&date, &count); err != nil {
			return nil, err
		}
		stats.ClicksByDate[date] = count
	}

	// статистику по странам
	countryQuery := `
		SELECT COALESCE(country, 'Unknown') as country, COUNT(*) as count
		FROM link_clicks
		WHERE link_id = $1 AND clicked_at BETWEEN $2 AND $3
		GROUP BY country
		ORDER BY count DESC
		LIMIT 10
	`
	countryRows, err := r.db.QueryContext(ctx, countryQuery, linkID, from, to)
	if err != nil {
		return nil, err
	}
	defer countryRows.Close()

	for countryRows.Next() {
		var country string
		var count int64
		if err := countryRows.Scan(&country, &count); err != nil {
			return nil, err
		}
		stats.ClicksByCountry[country] = count
	}

	// информация об устройстве из User-Agent
	deviceQuery := `
		SELECT
			CASE
				WHEN user_agent LIKE '%Mobile%' THEN 'Mobile'
				WHEN user_agent LIKE '%Tablet%' THEN 'Tablet'
				ELSE 'Desktop'
			END as device,
			COUNT(*) as count
		FROM link_clicks
		WHERE link_id = $1 AND clicked_at BETWEEN $2 AND $3
		GROUP BY device
		ORDER BY count DESC
	`
	deviceRows, err := r.db.QueryContext(ctx, deviceQuery, linkID, from, to)
	if err != nil {
		return nil, err
	}
	defer deviceRows.Close()

	for deviceRows.Next() {
		var device string
		var count int64
		if err := deviceRows.Scan(&device, &count); err != nil {
			return nil, err
		}
		stats.ClicksByDevice[device] = count
	}

	// топ реферреров
	refererQuery := `
		SELECT COALESCE(referer, 'Direct') as referer, COUNT(*) as count
		FROM link_clicks
		WHERE link_id = $1 AND clicked_at BETWEEN $2 AND $3
		GROUP BY referer
		ORDER BY count DESC
		LIMIT 5
	`
	refererRows, err := r.db.QueryContext(ctx, refererQuery, linkID, from, to)
	if err != nil {
		return nil, err
	}
	defer refererRows.Close()

	for refererRows.Next() {
		var refererStat entity.RefererStats
		if err := refererRows.Scan(&refererStat.Referer, &refererStat.Count); err != nil {
			return nil, err
		}
		stats.TopReferers = append(stats.TopReferers, refererStat)
	}

	return stats, nil
}

func (r *linkClickRepository) CountByLinkID(ctx context.Context, linkID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM link_clicks WHERE link_id = $1`

	var count int64
	err := r.db.QueryRowContext(ctx, query, linkID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *linkClickRepository) CountUniqueByLinkID(ctx context.Context, linkID int64) (int64, error) {
	query := `SELECT COUNT(DISTINCT ip_address) FROM link_clicks WHERE link_id = $1`

	var count int64
	err := r.db.QueryRowContext(ctx, query, linkID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
