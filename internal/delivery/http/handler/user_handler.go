package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raison-collab/LinkShorternetBackend/internal/delivery/http/dto"
	"github.com/raison-collab/LinkShorternetBackend/internal/usecase"
	"github.com/raison-collab/LinkShorternetBackend/pkg/logger"
	"github.com/raison-collab/LinkShorternetBackend/pkg/utils"
)

type userHandler struct {
	userUC usecase.UserUseCase
	log    logger.Logger
}

// NewUserHandler создает новый handler для работы с пользователями
func NewUserHandler(userUC usecase.UserUseCase, log logger.Logger) *userHandler {
	return &userHandler{
		userUC: userUC,
		log:    log,
	}
}

// GetProfile godoc
// @Summary Получение профиля пользователя
// @Description Возвращает информацию о текущем пользователе
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Security Bearer
// @Router /users/me [get]
func (h *userHandler) GetProfile(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Требуется авторизация",
		})
		return
	}

	user, err := h.userUC.GetByID(c.Request.Context(), *userID)
	if err != nil {
		h.log.Error("Ошибка получения пользователя:", err)

		// Определяем тип ошибки и возвращаем соответствующий статус
		switch err {
		case usecase.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Пользователь не найден",
			})
		default:
			// Внутренняя ошибка сервера
			h.log.Error("Внутренняя ошибка при получении профиля:", err)
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: "Внутренняя ошибка сервера",
			})
		}
		return
	}

	c.JSON(http.StatusOK, dto.UserFromEntity(user))
}

// UpdateProfile godoc
// @Summary Обновление профиля пользователя
// @Description Обновляет информацию о пользователе
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.UserUpdateRequest true "Данные для обновления"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security Bearer
// @Router /users/me [put]
func (h *userHandler) UpdateProfile(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Требуется авторизация",
		})
		return
	}

	var req dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Ошибка привязки запроса:", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный формат запроса",
		})
		return
	}

	err := h.userUC.Update(c.Request.Context(), *userID, req.Email)
	if err != nil {
		h.log.Error("Ошибка обновления пользователя:", err)

		// Определяем тип ошибки и возвращаем соответствующий статус
		switch err {
		case usecase.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Пользователь не найден",
			})
		case usecase.ErrInvalidEmail:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: err.Error(),
			})
		case usecase.ErrUserExists:
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: err.Error(),
			})
		default:
			// Внутренняя ошибка сервера
			h.log.Error("Внутренняя ошибка при обновлении профиля:", err)
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: "Внутренняя ошибка сервера",
			})
		}
		return
	}

	// Получаем обновленного пользователя
	user, err := h.userUC.GetByID(c.Request.Context(), *userID)
	if err != nil {
		h.log.Error("Ошибка получения обновленного пользователя:", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Обновление успешно, но не удалось получить обновленные данные",
		})
		return
	}

	c.JSON(http.StatusOK, dto.UserFromEntity(user))
}

// ChangePassword godoc
// @Summary Изменение пароля
// @Description Изменяет пароль текущего пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.PasswordChangeRequest true "Старый и новый пароль"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Security Bearer
// @Router /users/me/password [put]
func (h *userHandler) ChangePassword(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Требуется авторизация",
		})
		return
	}

	var req dto.PasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Ошибка привязки запроса:", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный формат запроса",
		})
		return
	}

	err := h.userUC.ChangePassword(c.Request.Context(), *userID, req.OldPassword, req.NewPassword)
	if err != nil {
		h.log.Error("Ошибка смены пароля:", err)

		// Определяем тип ошибки и возвращаем соответствующий статус
		switch err {
		case usecase.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Пользователь не найден",
			})
		case usecase.ErrInvalidCredentials:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Текущий пароль указан неверно",
			})
		case usecase.ErrInvalidPassword:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: err.Error(),
			})
		default:
			// Внутренняя ошибка сервера
			h.log.Error("Внутренняя ошибка при смене пароля:", err)
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: "Внутренняя ошибка сервера",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Пароль успешно изменен",
	})
}

// GetStats godoc
// @Summary Получение статистики пользователя
// @Description Возвращает статистику по ссылкам пользователя
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} entity.UserStats
// @Failure 401 {object} dto.ErrorResponse
// @Security Bearer
// @Router /users/me/stats [get]
func (h *userHandler) GetStats(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Требуется авторизация",
		})
		return
	}

	stats, err := h.userUC.GetStats(c.Request.Context(), *userID)
	if err != nil {
		h.log.Error("Ошибка получения статистики пользователя:", err)

		// Определяем тип ошибки и возвращаем соответствующий статус
		switch err {
		case usecase.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Пользователь не найден",
			})
		default:
			// Внутренняя ошибка сервера
			h.log.Error("Внутренняя ошибка при получении статистики:", err)
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: "Внутренняя ошибка сервера",
			})
		}
		return
	}

	c.JSON(http.StatusOK, stats)
}

// getUserIDFromContext извлекает ID пользователя из контекста
func getUserIDFromContext(c *gin.Context) *int64 {
	if claims, exists := c.Get("claims"); exists {
		if jwtClaims, ok := claims.(*utils.Claims); ok {
			return &jwtClaims.UserID
		}
	}
	return nil
}

// handleError обрабатывает ошибки и возвращает соответствующий HTTP-ответ
func handleError(c *gin.Context, log logger.Logger, err error, defaultMessage string) {
	log.Error(defaultMessage, err)

	switch err {
	case usecase.ErrUserNotFound:
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Пользователь не найден",
		})
	case usecase.ErrInvalidCredentials:
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Неверные учетные данные",
		})
	case usecase.ErrInvalidEmail, usecase.ErrInvalidPassword:
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
	case usecase.ErrUserExists:
		c.JSON(http.StatusConflict, dto.ErrorResponse{
			Error: err.Error(),
		})
	default:
		// Внутренняя ошибка сервера
		log.Error("Внутренняя ошибка:", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Внутренняя ошибка сервера",
		})
	}
}
