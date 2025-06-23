package dto

import (
	"time"

	"github.com/raison-collab/LinkShorternetBackend/internal/domain/entity"
)

// UserResponse представляет ответ с данными пользователя
type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserUpdateRequest представляет запрос на обновление пользователя
type UserUpdateRequest struct {
	Email string `json:"email,omitempty" binding:"omitempty,email"`
}

// PasswordChangeRequest представляет запрос на смену пароля
type PasswordChangeRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// UserFromEntity преобразует entity в DTO
func UserFromEntity(user *entity.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
