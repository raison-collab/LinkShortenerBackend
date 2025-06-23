package router

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/raison-collab/LinkShorternetBackend/internal/delivery/http/handler"
	"github.com/raison-collab/LinkShorternetBackend/internal/delivery/http/middleware"
	"github.com/raison-collab/LinkShorternetBackend/internal/infrastructure/config"
	"github.com/raison-collab/LinkShorternetBackend/internal/infrastructure/repository"
	"github.com/raison-collab/LinkShorternetBackend/internal/usecase"
	"github.com/raison-collab/LinkShorternetBackend/pkg/logger"
)

// NewRouter creates and configures a new router
func NewRouter(db *sql.DB, redisClient *redis.Client, cfg *config.Config, log logger.Logger) *gin.Engine {
	// Create repositories
	userRepo := repository.NewUserRepository(db)
	linkRepo := repository.NewLinkRepository(db)
	linkClickRepo := repository.NewLinkClickRepository(db)

	// Create use cases
	userUC := usecase.NewUserUseCase(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireHours)
	linkUC := usecase.NewLinkUseCase(linkRepo, linkClickRepo, cfg.URL.ShortURLLength, cfg.URL.BaseURL)

	// Create handlers
	authHandler := handler.NewAuthHandler(userUC, log)
	linkHandler := handler.NewLinkHandler(linkUC, log, cfg)
	userHandler := handler.NewUserHandler(userUC, log)

	// Create Gin router
	router := gin.New()

	// Global middleware
	router.Use(middleware.ErrorHandler(log)) // Должен быть первым для перехвата паники
	router.Use(middleware.Logger(log))
	router.Use(middleware.CORS(cfg))
	router.Use(middleware.RateLimiter(redisClient, cfg.RateLimit.Requests, cfg.RateLimit.WindowMinutes))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": cfg.App.Name,
		})
	})

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Short URL redirect (must be before API routes)
	router.GET("/:code", linkHandler.RedirectShortURL)

	// API routes
	api := router.Group("/api/v1")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.Auth(cfg.JWT.Secret))
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/me", userHandler.GetProfile)
				users.PUT("/me", userHandler.UpdateProfile)
				users.PUT("/me/password", userHandler.ChangePassword)
				users.GET("/me/stats", userHandler.GetStats)
			}

			// Link routes
			links := protected.Group("/links")
			{
				links.POST("", linkHandler.CreateLink)
				links.GET("", linkHandler.GetUserLinks)
				links.GET("/:id", linkHandler.GetLink)
				links.PUT("/:id", linkHandler.UpdateLink)
				links.DELETE("/:id", linkHandler.DeleteLink)
				links.GET("/:id/stats", linkHandler.GetLinkStats)
			}
		}

		// Public redirect inside API prefix (optional convenience)
		api.GET("/:code", linkHandler.RedirectShortURL)
	}

	return router
}
