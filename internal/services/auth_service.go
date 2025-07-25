package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"go-postgres-api/internal/models"
	"go-postgres-api/internal/repositories"
	"go-postgres-api/pkg/utilis"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWT configuration
const (
	accessTokenExpiryTime   = 15 * time.Minute   // Access token valid for 15 minutes
	refreshTokenExpiryTime  = 7 * 24 * time.Hour // Refresh token valid for 7 days
	verificationTokenExpiry = 24 * time.Hour     // Email verification token valid for 24 hours
)

// AuthService handles authentication logic
type AuthService struct {
	userRepo     *repositories.UserRepository
	emailService *EmailService
}


// getJWTSecret returns the JWT secret from environment variables
func (s *AuthService) getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-fallback-secret-key" // For development only
	}
	return secret
}

// Register registers a new user and sends verification email
func (s *AuthService) Register(req *models.RegisterRequest) (*models.SuccessResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Create new user with is_verified = false
	user := &models.User{
		Email:      req.Email,
		Name:       req.FirstName + " " + req.LastName,
		IsVerified: false,
		IsActive:   true,
		RoleID:     2, // Default role
	}

	// Set password
	if err := user.SetPassword(req.Password); err != nil {
		return nil, err
	}

	// Save user to database
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Generate email verification token
	verificationToken, err := s.generateEmailVerificationToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(user.Email, verificationToken); err != nil {
		// Log error but don't fail registration
		// In production, you might want to use a job queue for email sending
		// log.Printf("Failed to send verification email: %v", err)
	}

	return &models.SuccessResponse{
		Message: "User registered successfully. Please check your email to verify your account.",
	}, nil
}

// generateEmailVerificationToken generates a secure email verification token
func (s *AuthService) generateEmailVerificationToken(userID uint) (string, error) {
	// Generate secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	// Store token in database
	verificationToken := &models.EmailVerificationToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(verificationTokenExpiry),
		Used:      false,
	}

	if err := s.userRepo.CreateEmailVerificationToken(verificationToken); err != nil {
		return "", err
	}

	return token, nil
}

// VerifyEmail verifies a user's email using the verification token
func (s *AuthService) VerifyEmail(token string) (*models.SuccessResponse, error) {
	// Find and validate token
	verificationToken, err := s.userRepo.FindEmailVerificationToken(token)
	if err != nil {
		return nil, errors.New("invalid or expired verification token")
	}

	if verificationToken.Used {
		return nil, errors.New("verification token already used")
	}

	if time.Now().After(verificationToken.ExpiresAt) {
		return nil, errors.New("verification token expired")
	}

	// Update user verification status
	if err := s.userRepo.UpdateUserVerification(verificationToken.UserID, true); err != nil {
		return nil, err
	}

	// Mark token as used
	if err := s.userRepo.MarkEmailTokenAsUsed(verificationToken.ID); err != nil {
		return nil, err
	}

	return &models.SuccessResponse{
		Message: "Email verified successfully. You can now log in.",
	}, nil
}

// ResendVerificationEmail resends verification email
func (s *AuthService) ResendVerificationEmail(email string) (*models.SuccessResponse, error) {
	// Find user
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.IsVerified {
		return nil, errors.New("email already verified")
	}

	// Generate new verification token
	verificationToken, err := s.generateEmailVerificationToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(user.Email, verificationToken); err != nil {
		return nil, err
	}

	return &models.SuccessResponse{
		Message: "Verification email sent successfully.",
	}, nil
}

// Login authenticates a user and returns JWT tokens
func (s *AuthService) Login(req *models.LoginRequest, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}

	// Create auth log
	authLog := &models.AuthLog{
		Action:    "login",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   false,
	}

	// Check if user exists
	if user == nil {
		authLog.ErrorMessage = "user not found"
		s.userRepo.LogAuth(authLog)
		return nil, errors.New("invalid email or password")
	}

	authLog.UserID = user.ID

	// Check if email is verified
	if !user.IsVerified {
		authLog.ErrorMessage = "email not verified"
		s.userRepo.LogAuth(authLog)
		return nil, errors.New("please verify your email address before logging in")
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		authLog.ErrorMessage = "invalid password"
		s.userRepo.LogAuth(authLog)
		return nil, errors.New("invalid email or password")
	}

	// Generate access token
	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		authLog.ErrorMessage = "failed to generate access token"
		s.userRepo.LogAuth(authLog)
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		authLog.ErrorMessage = "failed to generate refresh token"
		s.userRepo.LogAuth(authLog)
		return nil, err
	}

	// Update last login time
	s.userRepo.UpdateLastLogin(user.ID)

	// Log successful login
	authLog.Success = true
	s.userRepo.LogAuth(authLog)

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenExpiryTime.Seconds()),
		User:         *user,
	}, nil
}

// generateAccessToken generates a JWT access token
func (s *AuthService) generateAccessToken(userID uint) (string, error) {
	tokenJTI := utilis.GenerateRandomString(36)
	expirationTime := time.Now().Add(accessTokenExpiryTime)
	claims := jwt.MapClaims{
		"sub":  userID,
		"exp":  expirationTime.Unix(),
		"iat":  time.Now().Unix(),
		"jti":  tokenJTI,
		"type": "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.getJWTSecret()))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// generateRefreshToken generates a refresh token
func (s *AuthService) generateRefreshToken(userID uint) (string, error) {
	// Generate secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	// Store refresh token in database
	refreshToken := &models.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(refreshTokenExpiryTime),
		Used:      false,
	}

	if err := s.userRepo.CreateRefreshToken(refreshToken); err != nil {
		return "", err
	}

	return token, nil
}

// RefreshAccessToken generates a new access token using refresh token
func (s *AuthService) RefreshAccessToken(refreshTokenString string) (*models.AuthResponse, error) {
	// Find and validate refresh token
	refreshToken, err := s.userRepo.FindRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	// Get user
	user, err := s.userRepo.FindByID(refreshToken.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Generate new refresh token
	newRefreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Mark old refresh token as used
	if err := s.userRepo.MarkRefreshTokenAsUsed(refreshToken.ID); err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(accessTokenExpiryTime.Seconds()),
		User:         *user,
	}, nil
}

// Logout blacklists a token
func (s *AuthService) Logout(tokenString string, userID uint) error {
	// Parse token to get claims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.getJWTSecret()), nil
	})
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return errors.New("invalid token")
	}

	// Get JTI for blacklisting
	jti, ok := claims["jti"].(string)
	if !ok {
		return errors.New("invalid token JTI")
	}

	// Get expiration time
	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("invalid token expiration")
	}

	// Add token to blacklist
	blacklistedToken := &models.TokenBlacklist{
		TokenJTI:  jti,
		UserID:    userID,
		ExpiresAt: time.Unix(int64(exp), 0),
	}

	return s.userRepo.BlacklistToken(blacklistedToken)
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (uint, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.getJWTSecret()), nil
	})
	if err != nil {
		return 0, err
	}

	// Validate token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}

	// Check if token is blacklisted
	jti, ok := claims["jti"].(string)
	if !ok {
		return 0, errors.New("invalid token JTI")
	}

	isBlacklisted, err := s.userRepo.IsTokenBlacklisted(jti)
	if err != nil {
		return 0, err
	}
	if isBlacklisted {
		return 0, errors.New("token is blacklisted")
	}

	// Get user ID
	userID, ok := claims["sub"].(float64)
	if !ok {
		return 0, errors.New("invalid user ID in token")
	}

	return uint(userID), nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}
