package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raison-collab/LinkShorternetBackend/internal/delivery/http/dto"
	"github.com/redis/go-redis/v9"
)

// RateLimiter создает middleware для ограничения количества запросов
func RateLimiter(redisClient *redis.Client, maxRequests int, windowMinutes int) gin.HandlerFunc {
	if redisClient == nil {
		return SimpleRateLimiter(maxRequests, windowMinutes)
	}

	return func(c *gin.Context) {
		ctx := context.Background()
		clientIP := c.ClientIP()

		key := fmt.Sprintf("rate_limit:%s", clientIP)

		// Инкрементируем счетчик
		pipe := redisClient.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Duration(windowMinutes)*time.Minute)
		_, err := pipe.Exec(ctx)

		if err != nil {
			// При ошибке Redis пропускаем запрос
			c.Next()
			return
		}

		count := incr.Val()

		c.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(max(0, maxRequests-int(count))))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Duration(windowMinutes)*time.Minute).Unix(), 10))

		// Проверяем лимит
		if count > int64(maxRequests) {
			c.JSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Error: "Слишком много запросов. Попробуйте позже.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SimpleRateLimiter создает простой in-memory rate limiter для случаев когда Redis недоступен
func SimpleRateLimiter(maxRequests int, windowMinutes int) gin.HandlerFunc {
	type clientInfo struct {
		count     int
		resetTime time.Time
		mu        sync.Mutex
	}

	clients := &sync.Map{}

	// Периодическая очистка старых записей
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			clients.Range(func(key, value interface{}) bool {
				if info, ok := value.(*clientInfo); ok {
					info.mu.Lock()
					if now.After(info.resetTime) {
						clients.Delete(key)
					}
					info.mu.Unlock()
				}
				return true
			})
		}
	}()

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Получаем или создаем информацию о клиенте
		value, _ := clients.LoadOrStore(clientIP, &clientInfo{
			count:     0,
			resetTime: now.Add(time.Duration(windowMinutes) * time.Minute),
			mu:        sync.Mutex{},
		})

		info := value.(*clientInfo)
		info.mu.Lock()
		defer info.mu.Unlock()

		// Проверяем, нужно ли сбросить счетчик
		if now.After(info.resetTime) {
			info.count = 0
			info.resetTime = now.Add(time.Duration(windowMinutes) * time.Minute)
		}

		info.count++

		c.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(max(0, maxRequests-info.count)))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(info.resetTime.Unix(), 10))

		// Проверяем лимит
		if info.count > maxRequests {
			c.JSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Error: "Слишком много запросов. Попробуйте позже.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// max возвращает максимальное из двух чисел
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
