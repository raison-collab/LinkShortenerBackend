package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/raison-collab/LinkShorternetBackend/internal/delivery/http/dto"
	"github.com/raison-collab/LinkShorternetBackend/pkg/logger"
)

// ErrorHandler создает middleware для обработки ошибок
func ErrorHandler(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				log.Error(
					"Паника в обработчике запроса",
					map[string]interface{}{
						"error": fmt.Sprintf("%v", err),
						"stack": string(stack),
						"path":  c.Request.URL.Path,
					},
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
					Error: "Внутренняя ошибка сервера",
					Code:  "INTERNAL_SERVER_ERROR",
				})
			}
		}()

		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Error(
				"Ошибка в обработчике запроса",
				map[string]interface{}{
					"error": err.Error(),
					"path":  c.Request.URL.Path,
				},
			)

			if c.Writer.Written() {
				return
			}

			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: "Внутренняя ошибка сервера",
				Code:  "INTERNAL_SERVER_ERROR",
			})
		}
	}
}
