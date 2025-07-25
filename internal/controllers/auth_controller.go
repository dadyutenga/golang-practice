package controllers

import (
	"go-postgres-api/internal/config"
	"go-postgres-api/internal/models"
	"go-postgres-api/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthController handles authentication requests
type AuthController struct {
	authService *services.AuthService
}

// NewAuthController creates a new authentication controller
func NewAuthController(cfg *config.Config) *AuthController {
	return &AuthController{
		authService: services.NewAuthService(),
	}
}

// Register handles user registration
func (c *AuthController) Register(ctx *gin.Context) {
	var req models.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	response, err := c.authService.Register(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// Login handles user login
func (c *AuthController) Login(ctx *gin.Context) {
	var req models.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	ipAddress := ctx.ClientIP()
	userAgent := ctx.GetHeader("User-Agent")

	response, err := c.authService.Login(&req, ipAddress, userAgent)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// Logout handles user logout
func (c *AuthController) Logout(ctx *gin.Context) {
	// Get token from Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "authorization header is required"})
		return
	}

	// Extract token from "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid authorization header format"})
		return
	}

	tokenString := tokenParts[1]

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "user ID not found in context"})
		return
	}

	// Blacklist token
	err := c.authService.Logout(tokenString, userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// GetProfile returns the user's profile
func (c *AuthController) GetProfile(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "user ID not found in context"})
		return
	}

	// Get user from database
	user, err := c.authService.GetUserByID(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// VerifyEmail handles email verification
func (c *AuthController) VerifyEmail(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "verification token is required"})
		return
	}

	response, err := c.authService.VerifyEmail(token)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// ResendVerificationEmail handles resending verification email
func (c *AuthController) ResendVerificationEmail(ctx *gin.Context) {
	var req models.ResendVerificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	response, err := c.authService.ResendVerificationEmail(req.Email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req models.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	response, err := c.authService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
