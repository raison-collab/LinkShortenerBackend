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
	ErrInvalidURL       = errors.New("invalid URL")
	ErrLinkNotFound     = errors.New("link not found")
	ErrLinkExpired      = errors.New("link has expired")
	ErrLinkInactive     = errors.New("link is inactive")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrShortCodeExists  = errors.New("short code already exists")
	ErrExpirationInPast = errors.New("expiration date cannot be in the past")
	ErrExpiration       = errors.New("date expired")
)

// LinkUseCase defines methods for link business logic
type LinkUseCase interface {
	CreateLink(ctx context.Context, originalURL string, userID *int64, customCode string, expiresAt *time.Time) (*entity.Link, error)
	GetLinkByShortCode(ctx context.Context, shortCode string) (*entity.Link, error)
	GetUserLinks(ctx context.Context, userID int64, offset, limit int) ([]*entity.Link, error)
	UpdateLink(ctx context.Context, linkID int64, userID int64, expiresAt *time.Time) error
	DeleteLink(ctx context.Context, linkID int64, userID int64) error
	RecordClick(ctx context.Context, shortCode, ipAddress, userAgent, referer string) (*entity.Link, error)
	GetLinkStats(ctx context.Context, linkID int64, userID int64, from, to time.Time) (*entity.LinkStats, error)
	GetLink(ctx context.Context, linkID int64, userID int64) (*entity.Link, error)
}

type linkUseCase struct {
	linkRepo      repository.LinkRepository
	linkClickRepo repository.LinkClickRepository
	shortURLLen   int
	baseURL       string
}

// NewLinkUseCase creates a new link use case
func NewLinkUseCase(linkRepo repository.LinkRepository, linkClickRepo repository.LinkClickRepository, shortURLLen int, baseURL string) LinkUseCase {
	return &linkUseCase{
		linkRepo:      linkRepo,
		linkClickRepo: linkClickRepo,
		shortURLLen:   shortURLLen,
		baseURL:       baseURL,
	}
}

// toUTC возвращает копию времени, приведённую к UTC. Если nil — оставляет nil.
func toUTC(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := t.UTC()
	return &v
}

// CreateLink создает новую короткую ссылку
func (uc *linkUseCase) CreateLink(ctx context.Context, originalURL string, userID *int64, customCode string, expiresAt *time.Time) (*entity.Link, error) {
	if !validator.IsValidURL(originalURL) {
		return nil, ErrInvalidURL
	}

	if expiresAt != nil && expiresAt.UTC().Before(time.Now().UTC()) {
		return nil, ErrExpirationInPast
	}

	var shortCode string
	if customCode != "" {
		exists, err := uc.linkRepo.ExistsByShortCode(ctx, customCode)
		if err != nil {
			return nil, fmt.Errorf("failed to check short code existence: %w", err)
		}
		if exists {
			return nil, ErrShortCodeExists
		}
		shortCode = customCode
	} else {
		for {
			shortCode = utils.GenerateShortCode(uc.shortURLLen)
			exists, err := uc.linkRepo.ExistsByShortCode(ctx, shortCode)
			if err != nil {
				return nil, fmt.Errorf("failed to check short code existence: %w", err)
			}
			if !exists {
				break
			}
		}
	}

	now := time.Now()
	link := &entity.Link{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		UserID:      userID,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.linkRepo.Create(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}

	return link, nil
}

// GetLinkByShortCode получает ссылку по короткому коду с проверкой активности и срока действия
func (uc *linkUseCase) GetLinkByShortCode(ctx context.Context, shortCode string) (*entity.Link, error) {
	link, err := uc.linkRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}
	if link == nil {
		return nil, ErrLinkNotFound
	}

	if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now().UTC()) {
		return nil, ErrLinkExpired
	}

	return link, nil
}

// GetUserLinks получает список ссылок пользователя с пагинацией
func (uc *linkUseCase) GetUserLinks(ctx context.Context, userID int64, offset, limit int) ([]*entity.Link, error) {
	links, err := uc.linkRepo.GetByUserID(ctx, userID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user links: %w", err)
	}
	return links, nil
}

// GetLink возвращает ссылку по ID с проверкой прав
func (uc *linkUseCase) GetLink(ctx context.Context, linkID int64, userID int64) (*entity.Link, error) {
	link, err := uc.linkRepo.GetByID(ctx, linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}
	if link == nil {
		return nil, ErrLinkNotFound
	}
	if link.UserID == nil || *link.UserID != userID {
		return nil, ErrUnauthorized
	}
	return link, nil
}

// UpdateLink обновляет информацию о ссылке
func (uc *linkUseCase) UpdateLink(ctx context.Context, linkID int64, userID int64, expiresAt *time.Time) error {
	link, err := uc.linkRepo.GetByID(ctx, linkID)
	if err != nil {
		return fmt.Errorf("failed to get link: %w", err)
	}
	if link == nil {
		return ErrLinkNotFound
	}

	if link.UserID == nil || *link.UserID != userID {
		return ErrUnauthorized
	}

	if expiresAt != nil && expiresAt.UTC().Before(time.Now().UTC()) {
		return ErrExpirationInPast
	}

	link.ExpiresAt = expiresAt
	link.UpdatedAt = time.Now().UTC()

	if err := uc.linkRepo.Update(ctx, link); err != nil {
		return fmt.Errorf("failed to update link: %w", err)
	}

	return nil
}

// DeleteLink удаляет ссылку с проверкой прав доступа
func (uc *linkUseCase) DeleteLink(ctx context.Context, linkID int64, userID int64) error {
	link, err := uc.linkRepo.GetByID(ctx, linkID)
	if err != nil {
		return fmt.Errorf("failed to get link: %w", err)
	}
	if link == nil {
		return ErrLinkNotFound
	}

	if link.UserID == nil || *link.UserID != userID {
		return ErrUnauthorized
	}

	if err := uc.linkRepo.Delete(ctx, linkID); err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	return nil
}

// RecordClick записывает клик по ссылке и увеличивает счетчик
func (uc *linkUseCase) RecordClick(ctx context.Context, shortCode, ipAddress, userAgent, referer string) (*entity.Link, error) {
	link, err := uc.GetLinkByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	click := &entity.LinkClick{
		LinkID:    link.ID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Referer:   referer,
		ClickedAt: time.Now(),
	}

	if link.ExpiresAt != nil && link.ExpiresAt.UTC().Before(time.Now().UTC()) {
		return nil, ErrExpiration
	}

	if err := uc.linkClickRepo.Create(ctx, click); err != nil {
		return nil, fmt.Errorf("failed to record click: %w", err)
	}

	if err := uc.linkRepo.IncrementClicks(ctx, link.ID); err != nil {
		return nil, fmt.Errorf("failed to increment clicks: %w", err)
	}

	return link, nil
}

// GetLinkStats получает статистику по ссылке за указанный период
func (uc *linkUseCase) GetLinkStats(ctx context.Context, linkID int64, userID int64, from, to time.Time) (*entity.LinkStats, error) {
	link, err := uc.linkRepo.GetByID(ctx, linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}
	if link == nil {
		return nil, ErrLinkNotFound
	}

	if link.UserID == nil || *link.UserID != userID {
		return nil, ErrUnauthorized
	}

	stats, err := uc.linkClickRepo.GetStats(ctx, linkID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return stats, nil
}
