package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/raison-collab/LinkShorternetBackend/internal/delivery/http/dto"
	"github.com/raison-collab/LinkShorternetBackend/pkg/utils"
)

// Auth создает middleware для проверки JWT токена
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "Требуется заголовок авторизации",
				Code:  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		// Проверяем формат: "Bearer <token>" или просто "<token>"
		var token string
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			token = authHeader
		}
		claims, err := utils.ValidateJWT(token, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "Недействительный или истекший токен",
				Code:  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Сохраняем claims в контексте для использования в handlers
		c.Set("claims", claims)
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)

		c.Next()
	}
}
