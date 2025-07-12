package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/raison-collab/LinkShorternetBackend/pkg/logger"
)

// Config holds all configuration for our application
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	URL       URLConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
	Log       LogConfig
}

// AppConfig holds application configuration
type AppConfig struct {
	Name  string
	Port  string
	Env   string
	Debug bool
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret      string
	ExpireHours int
}

// URLConfig holds URL configuration
type URLConfig struct {
	BaseURL        string
	ShortURLLength int
	APIHost        string // Host for API documentation
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests      int
	WindowMinutes int
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level    string
	Format   string
	Output   string // console, file, both
	FilePath string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist in production
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	config := &Config{
		App: AppConfig{
			Name:  getEnv("APP_NAME", "link-shortener"),
			Port:  getEnv("APP_PORT", "8080"),
			Env:   getEnv("APP_ENV", "development"),
			Debug: getEnvAsBool("APP_DEBUG", true),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "link_shortener"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "your-secret-key-here"),
			ExpireHours: getEnvAsInt("JWT_EXPIRE_HOURS", 24),
		},
		URL: URLConfig{
			BaseURL:        getEnv("BASE_URL", "http://localhost:8080"),
			ShortURLLength: getEnvAsInt("SHORT_URL_LENGTH", 6),
			APIHost:        getEnv("API_HOST", "localhost:8080"),
		},
		CORS: CORSConfig{
			AllowOrigins: getEnvAsStringSlice("CORS_ALLOW_ORIGINS", []string{"*"}),
			AllowMethods: getEnvAsStringSlice("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}),
			AllowHeaders: getEnvAsStringSlice("CORS_ALLOW_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization"}),
		},
		RateLimit: RateLimitConfig{
			Requests:      getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			WindowMinutes: getEnvAsInt("RATE_LIMIT_WINDOW_MINUTES", 1),
		},
		Log: LogConfig{
			Level:    getEnv("LOG_LEVEL", "debug"),
			Format:   getEnv("LOG_FORMAT", "json"),
			Output:   getEnv("LOG_OUTPUT", "console"),
			FilePath: getEnv("LOG_FILE_PATH", ""),
		},
	}

	log := logger.NewWithConfig(logger.Config{
		Level:    config.Log.Level,
		Format:   config.Log.Format,
		Output:   logger.LogOutput(config.Log.Output),
		FilePath: config.Log.FilePath,
	})

	// Прячем пароли
	configCopy := *config
	configCopy.Database.Password = "[MASKED]"
	configCopy.Redis.Password = "[MASKED]"
	configCopy.JWT.Secret = "[MASKED]"

	// Конвертируем конфиг в JSON для логирования
	configJSON, err := json.MarshalIndent(configCopy, "", "  ")
	if err != nil {
		log.Warnf("Failed to marshal config for debug logging: %v", err)
	} else {
		log.Debugf("Loaded configuration: %s", string(configJSON))
	}

	return config, nil
}

// GetDSN returns PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

// GetRedisAddr returns Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	strValue := getEnv(key, "")
	if value, err := strconv.ParseBool(strValue); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsStringSlice(key string, defaultValue []string) []string {
	strValue := getEnv(key, "")
	if strValue == "" {
		return defaultValue
	}

	// Простой парсинг строки в виде "value1,value2,value3"
	values := []string{}
	for _, v := range strings.Split(strValue, ",") {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}

	if len(values) == 0 {
		return defaultValue
	}
	return values
}
