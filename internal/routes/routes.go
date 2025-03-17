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
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.POST("/logout", middleware.AuthMiddleware(), authController.Logout)
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User profile
			protected.GET("/profile", authController.GetProfile)
			
			// User routes
			userRoutes := protected.Group("/users")
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
}