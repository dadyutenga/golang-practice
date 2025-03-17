package routes

import (
	"go-postgres-api/internal/config"

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
			
			userRoutes.POST("/", func(c *gin.Context) {
				c.JSON(201, gin.H{"message": "Create user"})
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

		// Add more route groups here as needed
		// Example: Product routes, Auth routes, etc.
	}
}