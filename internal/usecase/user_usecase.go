package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/raison-collab/LinkShorternetBackend/internal/domain/entity"
	"github.com/raison-collab/LinkShorternetBackend/internal/domain/repository"
	"github.com/raison-collab/LinkShorternetBackend/pkg/utils"
	"github.com/raison-collab/LinkShorternetBackend/pkg/validator"
)

var (
	ErrUserNotFound       = errors.New("пользователь не найден")
	ErrUserExists         = errors.New("пользователь с такой почтой уже существует")
	ErrInvalidEmail       = errors.New("некорректный формат электронной почты")
	ErrInvalidPassword    = errors.New("некорректный формат пароля")
	ErrInvalidCredentials = errors.New("неверные учетные данные")
)

// UserUseCase defines methods for user business logic
type UserUseCase interface {
	Register(ctx context.Context, email, password string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (*entity.User, string, error)
	GetByID(ctx context.Context, userID int64) (*entity.User, error)
	Update(ctx context.Context, userID int64, email string) error
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
	GetStats(ctx context.Context, userID int64) (*entity.UserStats, error)
}

type userUseCase struct {
	userRepo  repository.UserRepository
	jwtSecret string
	jwtExpire int
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(userRepo repository.UserRepository, jwtSecret string, jwtExpire int) UserUseCase {
	return &userUseCase{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpire: jwtExpire,
	}
}

// Register регистрирует нового пользователя в системе
func (uc *userUseCase) Register(ctx context.Context, email, password string) (*entity.User, error) {
	if !validator.IsValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	if !validator.IsValidPassword(password) {
		return nil, ErrInvalidPassword
	}

	existsByEmail, err := uc.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки существования email: %w", err)
	}
	if existsByEmail {
		return nil, ErrUserExists
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	now := time.Now()
	user := &entity.User{
		Email:        email,
		PasswordHash: hashedPassword,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	return user, nil
}

// Login выполняет аутентификацию пользователя и возвращает JWT токен
func (uc *userUseCase) Login(ctx context.Context, email, password string) (*entity.User, string, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка получения пользователя: %w", err)
	}
	if user == nil {
		return nil, "", ErrInvalidCredentials
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return nil, "", ErrInvalidCredentials
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, "user", uc.jwtSecret, time.Duration(uc.jwtExpire)*time.Hour)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка генерации токена: %w", err)
	}

	return user, token, nil
}

// GetByID получает пользователя по ID
func (uc *userUseCase) GetByID(ctx context.Context, userID int64) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Update обновляет информацию о пользователе
func (uc *userUseCase) Update(ctx context.Context, userID int64, email string) error {
	if email != "" && !validator.IsValidEmail(email) {
		return ErrInvalidEmail
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("ошибка получения пользователя: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	if email != "" {
		// Проверяем, не занят ли email другим пользователем
		exists, err := uc.userRepo.ExistsByEmail(ctx, email)
		if err != nil {
			return fmt.Errorf("ошибка проверки email: %w", err)
		}
		if exists {
			existingUser, err := uc.userRepo.GetByEmail(ctx, email)
			if err != nil {
				return fmt.Errorf("ошибка получения пользователя по email: %w", err)
			}
			if existingUser != nil && existingUser.ID != userID {
				return ErrUserExists
			}
		}

		user.Email = email
	}

	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("ошибка обновления пользователя: %w", err)
	}

	return nil
}

// ChangePassword изменяет пароль пользователя
func (uc *userUseCase) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("ошибка получения пользователя: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	if !utils.CheckPasswordHash(oldPassword, user.PasswordHash) {
		return ErrInvalidCredentials
	}

	if !validator.IsValidPassword(newPassword) {
		return ErrInvalidPassword
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("ошибка обновления пользователя: %w", err)
	}

	return nil
}

// GetStats получает статистику пользователя
func (uc *userUseCase) GetStats(ctx context.Context, userID int64) (*entity.UserStats, error) {
	// Проверяем существование пользователя
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	stats, err := uc.userRepo.GetStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения статистики пользователя: %w", err)
	}
	return stats, nil
}
