package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raison-collab/LinkShorternetBackend/internal/delivery/http/dto"
	"github.com/raison-collab/LinkShorternetBackend/internal/infrastructure/config"
	"github.com/raison-collab/LinkShorternetBackend/internal/usecase"
	"github.com/raison-collab/LinkShorternetBackend/pkg/logger"
	"github.com/raison-collab/LinkShorternetBackend/pkg/utils"
)

type linkHandler struct {
	linkUC usecase.LinkUseCase
	log    logger.Logger
	cfg    *config.Config
}

// NewLinkHandler создает новый handler для работы со ссылками
func NewLinkHandler(linkUC usecase.LinkUseCase, log logger.Logger, cfg *config.Config) *linkHandler {
	return &linkHandler{
		linkUC: linkUC,
		log:    log,
		cfg:    cfg,
	}
}

// CreateLink godoc
// @Summary Создание короткой ссылки
// @Description Создает короткую ссылку для указанного URL
// @Tags links
// @Accept json
// @Produce json
// @Param request body dto.CreateLinkRequest true "Данные для создания ссылки"
// @Success 201 {object} dto.LinkResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security Bearer
// @Router /links [post]
func (h *linkHandler) CreateLink(c *gin.Context) {
	var req dto.CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind request:", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request format",
		})
		return
	}

	userID := getUserID(c)

	link, err := h.linkUC.CreateLink(c.Request.Context(), req.URL, userID, req.CustomCode, req.ExpiresAt)
	if err != nil {
		h.log.Error("Failed to create link:", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.LinkFromEntity(link, h.cfg.URL.BaseURL))
}

// GetUserLinks godoc
// @Summary Получение списка ссылок пользователя
// @Description Возвращает список всех ссылок текущего пользователя
// @Tags links
// @Accept json
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество элементов на странице" default(20)
// @Success 200 {array} dto.LinkResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security Bearer
// @Router /links [get]
func (h *linkHandler) GetUserLinks(c *gin.Context) {
	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		h.log.Error("Invalid pagination params:", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid pagination parameters",
		})
		return
	}

	userID := getUserID(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	links, err := h.linkUC.GetUserLinks(c.Request.Context(), *userID, pagination.GetOffset(), pagination.Limit)
	if err != nil {
		h.log.Error("Failed to get user links:", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to get links",
		})
		return
	}

	response := make([]*dto.LinkResponse, len(links))
	for i, link := range links {
		response[i] = dto.LinkFromEntity(link, h.cfg.URL.BaseURL)
	}

	c.JSON(http.StatusOK, response)
}

// GetLink godoc
// @Summary Получение информации о ссылке
// @Description Возвращает детальную информацию о конкретной ссылке
// @Tags links
// @Accept json
// @Produce json
// @Param id path int true "ID ссылки"
// @Success 200 {object} dto.LinkResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security Bearer
// @Router /links/{id} [get]
func (h *linkHandler) GetLink(c *gin.Context) {
	linkID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid link ID",
		})
		return
	}

	userID := getUserID(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	link, err := h.linkUC.GetLink(c.Request.Context(), linkID, *userID)
	if err != nil {
		h.log.Error("Failed to get link:", err)
		h.respondLinkError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.LinkFromEntity(link, h.cfg.URL.BaseURL))
}

// UpdateLink godoc
// @Summary Обновление ссылки
// @Description Обновляет информацию о ссылке
// @Tags links
// @Accept json
// @Produce json
// @Param id path int true "ID ссылки"
// @Param request body dto.UpdateLinkRequest true "Данные для обновления"
// @Success 200 {object} dto.LinkResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security Bearer
// @Router /links/{id} [put]
func (h *linkHandler) UpdateLink(c *gin.Context) {
	linkID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid link ID",
		})
		return
	}

	var req dto.UpdateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind request:", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request format",
		})
		return
	}

	userID := getUserID(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	err = h.linkUC.UpdateLink(c.Request.Context(), linkID, *userID, req.ExpiresAt)
	if err != nil {
		h.log.Error("Failed to update link:", err)
		h.respondLinkError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Link updated successfully"})
}

// DeleteLink godoc
// @Summary Удаление ссылки
// @Description Удаляет ссылку
// @Tags links
// @Accept json
// @Produce json
// @Param id path int true "ID ссылки"
// @Success 204
// @Failure 404 {object} dto.ErrorResponse
// @Security Bearer
// @Router /links/{id} [delete]
func (h *linkHandler) DeleteLink(c *gin.Context) {
	linkID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid link ID",
		})
		return
	}

	userID := getUserID(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	err = h.linkUC.DeleteLink(c.Request.Context(), linkID, *userID)
	if err != nil {
		h.log.Error("Failed to delete link:", err)
		h.respondLinkError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetLinkStats godoc
// @Summary Получение статистики по ссылке
// @Description Возвращает статистику переходов по ссылке
// @Tags links
// @Accept json
// @Produce json
// @Param id path int true "ID ссылки"
// @Param from query string false "Дата начала периода (RFC3339)"
// @Param to query string false "Дата конца периода (RFC3339)"
// @Success 200 {object} dto.LinkStatsResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security Bearer
// @Router /links/{id}/stats [get]
func (h *linkHandler) GetLinkStats(c *gin.Context) {
	linkID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid link ID",
		})
		return
	}

	userID := getUserID(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	// Парсинг дат из query параметров
	from := time.Now().AddDate(0, -1, 0) // По умолчанию - месяц назад
	to := time.Now()

	if fromStr := c.Query("from"); fromStr != "" {
		if parsedFrom, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = parsedFrom
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if parsedTo, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = parsedTo
		}
	}

	stats, err := h.linkUC.GetLinkStats(c.Request.Context(), linkID, *userID, from, to)
	if err != nil {
		h.log.Error("Failed to get link stats:", err)
		h.respondLinkError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.LinkStatsFromEntity(stats))
}

// RedirectShortURL godoc
// @Summary Переход по короткой ссылке
// @Description Перенаправляет на оригинальный URL и записывает статистику
// @Tags redirect
// @Param code path string true "Короткий код"
// @Success 302
// @Failure 404 {object} dto.ErrorResponse
// @Router /{code} [get]
func (h *linkHandler) RedirectShortURL(c *gin.Context) {
	shortCode := c.Param("code")

	link, err := h.linkUC.RecordClick(
		c.Request.Context(),
		shortCode,
		c.ClientIP(),
		c.Request.UserAgent(),
		c.Request.Referer(),
	)
	if err != nil {
		h.log.Error("Failed to process redirect:", err)
		h.respondLinkError(c, err)
		return
	}

	c.Redirect(http.StatusFound, link.OriginalURL)
}

// getUserID извлекает ID пользователя из контекста
func getUserID(c *gin.Context) *int64 {
	if claims, exists := c.Get("claims"); exists {
		if jwtClaims, ok := claims.(*utils.Claims); ok {
			return &jwtClaims.UserID
		}
	}
	return nil
}

// respondLinkError переводит бизнес-ошибки в HTTP-статусы
func (h *linkHandler) respondLinkError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrUnauthorized):
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "Forbidden"})
	case errors.Is(err, usecase.ErrLinkExpired):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "Link expired"})
	case errors.Is(err, usecase.ErrLinkNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Link not found"})
	case errors.Is(err, usecase.ErrShortCodeExists), errors.Is(err, usecase.ErrExpirationInPast), errors.Is(err, usecase.ErrInvalidURL):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Internal server error"})
	}
}
