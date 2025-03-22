package configs

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// ServerConfig holds the server configuration
type ServerConfig struct {
	Port string
	Mode string
}

// DatabaseConfig holds the database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// JWTConfig holds the JWT configuration
type JWTConfig struct {
	SecretKey    string
	Issuer       string
	ExpiresInMin int
}

// LoadConfig loads the application configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "subscription_service"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			SecretKey:    getEnv("JWT_SECRET_KEY", "your-secret-key"),
			Issuer:       getEnv("JWT_ISSUER", "subscription-service"),
			ExpiresInMin: getEnvAsInt("JWT_EXPIRES_IN_MIN", 60), // 1 hour default
		},
	}

	// Validate required configuration
	if config.JWT.SecretKey == "your-secret-key" {
		fmt.Println("WARNING: Using default JWT secret key. This is insecure!")
	}

	return config, nil
}

// DatabaseURL returns the database connection URL
func (c *DatabaseConfig) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode,
	)
}

// GetJWTExpirationDuration returns the JWT expiration duration
func (c *JWTConfig) GetJWTExpirationDuration() time.Duration {
	return time.Duration(c.ExpiresInMin) * time.Minute
}

// Helper function to get environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Helper function to get environment variable as int with fallback
func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}
