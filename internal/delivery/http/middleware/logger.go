package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raison-collab/LinkShorternetBackend/pkg/logger"
)

// Logger создает middleware для логирования HTTP запросов
func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		var requestBody []byte
		if c.Request.Body != nil && c.Request.Method != http.MethodGet {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// Обработка запроса
		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logFields := map[string]interface{}{
			"status":     statusCode,
			"latency":    latency,
			"client_ip":  clientIP,
			"method":     method,
			"path":       path,
			"user_agent": c.Request.UserAgent(),
		}

		if len(requestBody) > 0 && len(requestBody) < 10000 {
			logFields["request_body"] = string(requestBody)
		}

		if statusCode >= 400 && responseWriter.body.Len() < 10000 {
			logFields["response_body"] = responseWriter.body.String()
		}

		switch {
		case statusCode >= 500:
			log.Error("Серверная ошибка", logFields)
		case statusCode >= 400:
			log.Warn("Клиентская ошибка", logFields)
		default:
			log.Info("Запрос обработан", logFields)
		}
	}
}

// responseBodyWriter - обертка для записи тела ответа
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write перехватывает запись в ответ
func (r *responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// WriteString перехватывает запись строки в ответ
func (r *responseBodyWriter) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}
