package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raison-collab/LinkShorternetBackend/internal/delivery/http/dto"
	"github.com/raison-collab/LinkShorternetBackend/internal/usecase"
	"github.com/raison-collab/LinkShorternetBackend/pkg/logger"
)

type authHandler struct {
	userUC usecase.UserUseCase
	log    logger.Logger
}

// NewAuthHandler создает новый handler для аутентификации
func NewAuthHandler(userUC usecase.UserUseCase, log logger.Logger) *authHandler {
	return &authHandler{
		userUC: userUC,
		log:    log,
	}
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя в системе
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (h *authHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Ошибка привязки запроса:", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный формат запроса",
		})
		return
	}

	_, err := h.userUC.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.log.Error("Ошибка регистрации пользователя:", err)

		switch err {
		case usecase.ErrInvalidEmail, usecase.ErrInvalidPassword, usecase.ErrUserExists:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: err.Error(),
			})
		default:
			h.log.Error("Внутренняя ошибка при регистрации:", err)
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: "Внутренняя ошибка сервера",
			})
		}
		return
	}

	// Сразу логиним после регистрации
	_, token, err := h.userUC.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.log.Error("Ошибка входа после регистрации:", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Регистрация успешна, но не удалось выполнить вход",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.AuthResponse{
		Token: token,
	})
}

// Login godoc
// @Summary Вход в систему
// @Description Аутентификация пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Данные для входа"
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *authHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Ошибка привязки запроса:", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный формат запроса",
		})
		return
	}

	_, token, err := h.userUC.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.log.Error("Ошибка входа:", err)

		// Определяем тип ошибки и возвращаем соответствующий статус
		switch err {
		case usecase.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "Неверный email или пароль",
			})
		default:
			// Внутренняя ошибка сервера
			h.log.Error("Внутренняя ошибка при входе:", err)
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: "Внутренняя ошибка сервера",
			})
		}
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		Token: token,
	})
}
