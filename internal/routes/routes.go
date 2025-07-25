package routes

import (
	"go-postgres-api/internal/config"
	"go-postgres-api/internal/controllers"
	"go-postgres-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(router *gin.Engine, cfg *config.Config) {
	// API v1 routes group
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
			})
		})

		// Auth routes
		authController := controllers.NewAuthController()
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", authController.Register)
			authRoutes.POST("/login", authController.Login)
			authRoutes.GET("/oauth/login", authController.OAuthLogin)
			authRoutes.GET("/oauth/callback", authController.OAuthCallback)

			// Protected routes
			protected := authRoutes.Group("/")
			protected.Use(middleware.AuthMiddleware())
			{
				protected.POST("/logout", authController.Logout)
				protected.GET("/profile", authController.GetProfile)
			}
		}

		// User routes
		userRoutes := v1.Group("/users")
		{
			userRoutes.GET("/", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Get all users"})
			})

			userRoutes.GET("/:id", func(c *gin.Context) {
				id := c.Param("id")
				c.JSON(200, gin.H{"message": "Get user", "id": id})
			})

			userRoutes.PUT("/:id", func(c *gin.Context) {
				id := c.Param("id")
				c.JSON(200, gin.H{"message": "Update user", "id": id})
			})

			userRoutes.DELETE("/:id", func(c *gin.Context) {
				id := c.Param("id")
				c.JSON(200, gin.H{"message": "Delete user", "id": id})
			})
		}
	}
}
