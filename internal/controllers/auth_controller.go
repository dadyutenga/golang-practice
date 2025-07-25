package controllers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"go-postgres-api/internal/authenticator"
	"go-postgres-api/internal/config"
	"go-postgres-api/internal/models"
	"go-postgres-api/internal/services"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AuthController handles authentication requests
type AuthController struct {
	authService   *services.AuthService
	authenticator *authenticator.Authenticator
}

// NewAuthController creates a new authentication controller
func NewAuthController(cfg *config.Config) *AuthController {
	auth, err := authenticator.New(cfg)
	if err != nil {
		panic("Failed to initialize authenticator: " + err.Error())
	}

	return &AuthController{
		authService:   services.NewAuthService(),
		authenticator: auth,
	}
}

// Register handles user registration
func (c *AuthController) Register(ctx *gin.Context) {
	var req models.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := c.authService.Register(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, user)
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

// OAuthLogin initiates OAuth login flow
func (c *AuthController) OAuthLogin(ctx *gin.Context) {
	state, err := generateRandomState()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate state"})
		return
	}

	// Store state in session for verification
	session := sessions.Default(ctx)
	session.Set("state", state)
	session.Save()

	// Redirect to OAuth provider
	url := c.authenticator.AuthCodeURL(state)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

// OAuthCallback handles OAuth callback
func (c *AuthController) OAuthCallback(ctx *gin.Context) {
	session := sessions.Default(ctx)

	// Verify state parameter
	if ctx.Query("state") != session.Get("state") {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid state parameter"})
		return
	}

	// Clear state from session
	session.Delete("state")
	session.Save()

	// Exchange code for token
	token, err := c.authenticator.Exchange(context.Background(), ctx.Query("code"))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "failed to exchange token"})
		return
	}

	// Verify ID token
	idToken, err := c.authenticator.VerifyIDToken(context.Background(), token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "failed to verify ID token"})
		return
	}

	// Extract user information
	var profile map[string]interface{}
	if err := idToken.Claims(&profile); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to get user profile"})
		return
	}

	// Process OAuth user (create or login)
	user, jwtToken, err := c.authService.ProcessOAuthUser(profile, ctx.ClientIP(), ctx.GetHeader("User-Agent"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Store profile in session for the middleware
	session.Set("profile", profile)
	session.Save()

	// Return user info and JWT token
	ctx.JSON(http.StatusOK, gin.H{
		"user":         user,
		"access_token": jwtToken,
		"profile":      profile,
	})
}

// generateRandomState generates a random state for OAuth
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
