package entity

import (
	"time"
)

// Link represents a shortened URL entity
type Link struct {
	ID          int64      `json:"id" db:"id"`
	ShortCode   string     `json:"short_code" db:"short_code"`
	OriginalURL string     `json:"original_url" db:"original_url"`
	UserID      *int64     `json:"user_id,omitempty" db:"user_id"`
	Clicks      int64      `json:"clicks" db:"clicks"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// LinkClick represents a click event on a shortened link
type LinkClick struct {
	ID        int64     `json:"id" db:"id"`
	LinkID    int64     `json:"link_id" db:"link_id"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	Referer   string    `json:"referer,omitempty" db:"referer"`
	Country   string    `json:"country,omitempty" db:"country"`
	City      string    `json:"city,omitempty" db:"city"`
	ClickedAt time.Time `json:"clicked_at" db:"clicked_at"`
}

// LinkStats represents statistics for a link
type LinkStats struct {
	LinkID       int64                  `json:"link_id"`
	TotalClicks  int64                  `json:"total_clicks"`
	UniqueClicks int64                  `json:"unique_clicks"`
	ClicksByDate map[string]int64       `json:"clicks_by_date"`
	ClicksByCountry map[string]int64    `json:"clicks_by_country"`
	ClicksByDevice map[string]int64     `json:"clicks_by_device"`
	TopReferers  []RefererStats         `json:"top_referers"`
}

// RefererStats represents referrer statistics
type RefererStats struct {
	Referer string `json:"referer"`
	Count   int64  `json:"count"`
} 