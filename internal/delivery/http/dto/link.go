package dto

import (
	"time"

	"github.com/raison-collab/LinkShorternetBackend/internal/domain/entity"
)

// CreateLinkRequest представляет запрос на создание ссылки
type CreateLinkRequest struct {
	URL        string     `json:"url" binding:"required,url"`
	CustomCode string     `json:"custom_code,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" example:"2025-12-31T23:59:59Z"`
}

// UpdateLinkRequest представляет запрос на обновление ссылки
type UpdateLinkRequest struct {
	ExpiresAt *time.Time `json:"expires_at,omitempty" example:"2025-12-31T23:59:59Z"`
}

// LinkResponse представляет ответ с данными ссылки
type LinkResponse struct {
	ID          int64      `json:"id"`
	ShortCode   string     `json:"short_code"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	Clicks      int64      `json:"clicks"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// LinkStatsResponse представляет статистику по ссылке
type LinkStatsResponse struct {
	LinkID          int64                  `json:"link_id"`
	TotalClicks     int64                  `json:"total_clicks"`
	UniqueClicks    int64                  `json:"unique_clicks"`
	ClicksByDate    map[string]int64       `json:"clicks_by_date"`
	ClicksByCountry map[string]int64       `json:"clicks_by_country"`
	ClicksByDevice  map[string]int64       `json:"clicks_by_device"`
	TopReferers     []RefererStatsResponse `json:"top_referers"`
}

// RefererStatsResponse представляет статистику по источникам переходов
type RefererStatsResponse struct {
	Referer string `json:"referer"`
	Count   int64  `json:"count"`
}

// LinkFromEntity преобразует entity в DTO
func LinkFromEntity(link *entity.Link, baseURL string) *LinkResponse {
	return &LinkResponse{
		ID:          link.ID,
		ShortCode:   link.ShortCode,
		ShortURL:    baseURL + "/" + link.ShortCode,
		OriginalURL: link.OriginalURL,
		Clicks:      link.Clicks,
		ExpiresAt:   link.ExpiresAt,
		CreatedAt:   link.CreatedAt,
		UpdatedAt:   link.UpdatedAt,
	}
}

// LinkStatsFromEntity преобразует entity статистики в DTO
func LinkStatsFromEntity(stats *entity.LinkStats) *LinkStatsResponse {
	referers := make([]RefererStatsResponse, len(stats.TopReferers))
	for i, ref := range stats.TopReferers {
		referers[i] = RefererStatsResponse{
			Referer: ref.Referer,
			Count:   ref.Count,
		}
	}

	return &LinkStatsResponse{
		LinkID:          stats.LinkID,
		TotalClicks:     stats.TotalClicks,
		UniqueClicks:    stats.UniqueClicks,
		ClicksByDate:    stats.ClicksByDate,
		ClicksByCountry: stats.ClicksByCountry,
		ClicksByDevice:  stats.ClicksByDevice,
		TopReferers:     referers,
	}
}
