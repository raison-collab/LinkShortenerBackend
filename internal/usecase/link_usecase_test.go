package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/raison-collab/LinkShorternetBackend/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLinkRepository is a mock implementation of LinkRepository
type MockLinkRepository struct {
	mock.Mock
}

func (m *MockLinkRepository) Create(ctx context.Context, link *entity.Link) error {
	args := m.Called(ctx, link)
	return args.Error(0)
}

func (m *MockLinkRepository) GetByShortCode(ctx context.Context, shortCode string) (*entity.Link, error) {
	args := m.Called(ctx, shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Link), args.Error(1)
}

func (m *MockLinkRepository) GetByID(ctx context.Context, id int64) (*entity.Link, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Link), args.Error(1)
}

func (m *MockLinkRepository) GetByUserID(ctx context.Context, userID int64, offset, limit int) ([]*entity.Link, error) {
	args := m.Called(ctx, userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Link), args.Error(1)
}

func (m *MockLinkRepository) Update(ctx context.Context, link *entity.Link) error {
	args := m.Called(ctx, link)
	return args.Error(0)
}

func (m *MockLinkRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockLinkRepository) IncrementClicks(ctx context.Context, linkID int64) error {
	args := m.Called(ctx, linkID)
	return args.Error(0)
}

func (m *MockLinkRepository) GetExpiredLinks(ctx context.Context, before time.Time) ([]*entity.Link, error) {
	args := m.Called(ctx, before)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Link), args.Error(1)
}

func (m *MockLinkRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLinkRepository) ExistsByShortCode(ctx context.Context, shortCode string) (bool, error) {
	args := m.Called(ctx, shortCode)
	return args.Bool(0), args.Error(1)
}

// MockLinkClickRepository is a mock implementation of LinkClickRepository
type MockLinkClickRepository struct {
	mock.Mock
}

func (m *MockLinkClickRepository) Create(ctx context.Context, click *entity.LinkClick) error {
	args := m.Called(ctx, click)
	return args.Error(0)
}

func (m *MockLinkClickRepository) GetByLinkID(ctx context.Context, linkID int64, offset, limit int) ([]*entity.LinkClick, error) {
	args := m.Called(ctx, linkID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.LinkClick), args.Error(1)
}

func (m *MockLinkClickRepository) GetStats(ctx context.Context, linkID int64, from, to time.Time) (*entity.LinkStats, error) {
	args := m.Called(ctx, linkID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.LinkStats), args.Error(1)
}

func (m *MockLinkClickRepository) CountByLinkID(ctx context.Context, linkID int64) (int64, error) {
	args := m.Called(ctx, linkID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLinkClickRepository) CountUniqueByLinkID(ctx context.Context, linkID int64) (int64, error) {
	args := m.Called(ctx, linkID)
	return args.Get(0).(int64), args.Error(1)
}

// Tests

func TestLinkUseCase_CreateLink(t *testing.T) {
	ctx := context.Background()
	mockLinkRepo := new(MockLinkRepository)
	mockClickRepo := new(MockLinkClickRepository)

	uc := NewLinkUseCase(mockLinkRepo, mockClickRepo, 6, "http://localhost:8080")

	t.Run("Success - Create link with auto-generated code", func(t *testing.T) {
		// Mock expectations
		mockLinkRepo.On("ExistsByShortCode", ctx, mock.AnythingOfType("string")).Return(false, nil)
		mockLinkRepo.On("Create", ctx, mock.AnythingOfType("*entity.Link")).Return(nil)

		// Execute
		link, err := uc.CreateLink(ctx, "https://example.com", nil, "", nil)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, link)
		assert.Equal(t, "https://example.com", link.OriginalURL)
		assert.NotEmpty(t, link.ShortCode)
		assert.Len(t, link.ShortCode, 6)

		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("Success - Create link with custom code", func(t *testing.T) {
		customCode := "custom123"

		// Mock expectations
		mockLinkRepo.On("ExistsByShortCode", ctx, customCode).Return(false, nil)
		mockLinkRepo.On("Create", ctx, mock.AnythingOfType("*entity.Link")).Return(nil)

		// Execute
		link, err := uc.CreateLink(ctx, "https://example.com", nil, customCode, nil)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, link)
		assert.Equal(t, customCode, link.ShortCode)

		mockLinkRepo.AssertExpectations(t)
	})
}

func TestLinkUseCase_GetLinkByShortCode(t *testing.T) {
	ctx := context.Background()
	mockLinkRepo := new(MockLinkRepository)
	mockClickRepo := new(MockLinkClickRepository)

	uc := NewLinkUseCase(mockLinkRepo, mockClickRepo, 6, "http://localhost:8080")

	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		expectedLink := &entity.Link{
			ID:          1,
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
			Clicks:      0,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		// Mock expectations
		mockLinkRepo.On("GetByShortCode", ctx, "abc123").Return(expectedLink, nil)

		// Execute
		link, err := uc.GetLinkByShortCode(ctx, "abc123")

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, link)
		assert.Equal(t, expectedLink, link)

		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("Error - Link not found", func(t *testing.T) {
		// Mock expectations
		mockLinkRepo.On("GetByShortCode", ctx, "notfound").Return(nil, nil)

		// Execute
		link, err := uc.GetLinkByShortCode(ctx, "notfound")

		// Assertions
		assert.Error(t, err)
		assert.Equal(t, ErrLinkNotFound, err)
		assert.Nil(t, link)

		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("Error - Link expired", func(t *testing.T) {
		expiredTime := time.Now().Add(-1 * time.Hour)
		expiredLink := &entity.Link{
			ID:          1,
			ShortCode:   "expired",
			OriginalURL: "https://example.com",
			ExpiresAt:   &expiredTime,
		}

		// Mock expectations
		mockLinkRepo.On("GetByShortCode", ctx, "expired").Return(expiredLink, nil)

		// Execute
		link, err := uc.GetLinkByShortCode(ctx, "expired")

		// Assertions
		assert.Error(t, err)
		assert.Equal(t, ErrLinkExpired, err)
		assert.Nil(t, link)

		mockLinkRepo.AssertExpectations(t)
	})
}
