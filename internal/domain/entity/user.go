package entity

import (
	"time"
)

// User represents a user entity
type User struct {
	ID           int64     `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole represents user roles
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// UserStats represents user statistics
type UserStats struct {
	UserID      int64 `json:"user_id"`
	TotalLinks  int64 `json:"total_links"`
	TotalClicks int64 `json:"total_clicks"`
	ActiveLinks int64 `json:"active_links"`
}
