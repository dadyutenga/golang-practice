package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		ServerPort: os.Getenv("PORT"),
	}

	// Set default port if not provided
	if config.ServerPort == "" {
		config.ServerPort = "8080"
	}

	return config, nil
}