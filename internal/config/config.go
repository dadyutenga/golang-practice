package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	// Database Configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Server Configuration
	ServerPort string

	// OAuth Configuration
	Auth0Domain       string
	Auth0ClientID     string
	Auth0ClientSecret string
	Auth0CallbackURL  string

	// JWT Configuration
	JWTSecret string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		// Database
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),

		// Server
		ServerPort: os.Getenv("PORT"),

		// OAuth
		Auth0Domain:       os.Getenv("AUTH0_DOMAIN"),
		Auth0ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
		Auth0ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		Auth0CallbackURL:  os.Getenv("AUTH0_CALLBACK_URL"),

		// JWT
		JWTSecret: os.Getenv("JWT_SECRET"),
	}

	// Set default values if not provided
	if config.ServerPort == "" {
		config.ServerPort = "8080"
	}

	if config.DBHost == "" {
		config.DBHost = "localhost"
	}

	if config.DBPort == "" {
		config.DBPort = "3306"
	}

	return config, nil
}
