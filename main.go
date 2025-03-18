package main

import (
	"log"
	"os"

	"go-postgres-api/internal/config"
	"go-postgres-api/internal/database"
	"go-postgres-api/internal/middleware"
	"go-postgres-api/internal/models"
	"go-postgres-api/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables with a relative path
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Error loading .env file: %v", err)
		log.Println("Continuing with environment variables from the system...")
	}

	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Check if migrations have already been applied
	var migrationCount int64
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'").Count(&migrationCount)

	// Only run migrations if the users table doesn't exist
	if migrationCount == 0 {
		log.Println("Running database migrations...")
		// Auto migrate models
		err = db.AutoMigrate(
			&models.User{},
			&models.Role{},
			&models.AuthLog{},
			&models.TokenBlacklist{},
		)
		if err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}
		log.Println("Database migrations completed successfully")
	} else {
		log.Println("Skipping migrations as database schema already exists")
	}

	// Get the underlying SQL DB to set up connection pool parameters
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get SQL DB: %v", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// Initialize Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(middleware.CORSMiddleware())

	// Set up routes
	routes.SetupRoutes(router, cfg)

	// Start the server with specific IP and port
	serverAddress := "31.220.82.177:9000"
	log.Printf("Server running on http://%s", serverAddress)
	log.Fatal(router.Run(serverAddress))
}
