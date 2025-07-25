package services

import (
	"errors"
	"go-postgres-api/internal/models"
	"go-postgres-api/internal/repositories"
	"go-postgres-api/pkg/utilis"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWT configuration
const (
	jwtSecret     = "your-secret-key" // In production, use environment variables
	jwtExpiryTime = 24 * time.Hour    // Token valid for 24 hours
)

// AuthService handles authentication logic
type AuthService struct {
	userRepo *repositories.UserRepository
}

// NewAuthService creates a new authentication service
func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: repositories.NewUserRepository(),
	}
}

// Register registers a new user
func (s *AuthService) Register(req *models.RegisterRequest) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Create new user
	user := &models.User{
		Email:    req.Email,
		Name:     req.FirstName + " " + req.LastName,
		IsActive: true,
	}

	// Set password
	if err := user.SetPassword(req.Password); err != nil {
		return nil, err
	}

	// Save user to database
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Add default role
	if err := s.userRepo.AddRole(user.ID, "user"); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
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

	// Check if user is active
	if !user.IsActive {
		authLog.ErrorMessage = "user account is inactive"
		s.userRepo.LogAuth(authLog)
		return nil, errors.New("account is inactive")
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		authLog.ErrorMessage = "invalid password"
		s.userRepo.LogAuth(authLog)
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	tokenJTI := utilis.GenerateRandomString(36)
	expirationTime := time.Now().Add(jwtExpiryTime)
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": expirationTime.Unix(),
		"iat": time.Now().Unix(),
		"jti": tokenJTI,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		authLog.ErrorMessage = "failed to generate token"
		s.userRepo.LogAuth(authLog)
		return nil, err
	}

	// Update last login time
	s.userRepo.UpdateLastLogin(user.ID)

	// Log successful login
	authLog.Success = true
	s.userRepo.LogAuth(authLog)

	return &models.AuthResponse{
		Token:     tokenString,
		ExpiresIn: int64(jwtExpiryTime.Seconds()),
		User:      *user,
	}, nil
}

// Logout blacklists a token
func (s *AuthService) Logout(tokenString string, userID uint) error {
	// Parse token to get claims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return err
	}

	// Get token JTI
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid token claims")
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return errors.New("invalid token JTI")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("invalid token expiration")
	}

	// Add token to blacklist
	blacklist := &models.TokenBlacklist{
		TokenJTI:  jti,
		UserID:    userID,
		ExpiresAt: time.Unix(int64(exp), 0),
	}

	return s.userRepo.BlacklistToken(blacklist)
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (uint, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return 0, err
	}

	// Validate token
	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	// Get claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
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

// GenerateToken generates a JWT token for a user
func (s *AuthService) GenerateToken(userID uint) (string, error) {
	tokenJTI := utilis.GenerateRandomString(36)
	expirationTime := time.Now().Add(jwtExpiryTime)
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": expirationTime.Unix(),
		"iat": time.Now().Unix(),
		"jti": tokenJTI,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ProcessOAuthUser processes OAuth user login/registration
func (s *AuthService) ProcessOAuthUser(profile map[string]interface{}, ipAddress, userAgent string) (*models.User, string, error) {
	email, ok := profile["email"].(string)
	if !ok || email == "" {
		return nil, "", errors.New("email not found in OAuth profile")
	}

	name, _ := profile["name"].(string)
	if name == "" {
		name = email // fallback to email if name not available
	}

	// Check if user exists
	existingUser, err := s.userRepo.FindByEmail(email)
	if err == nil && existingUser != nil {
		// User exists, generate token and return
		token, err := s.GenerateToken(existingUser.ID)
		if err != nil {
			return nil, "", err
		}

		// Log authentication
		// Note: You'll need to implement AuthLog repository methods
		// authLog := models.AuthLog{
		// 	UserID:    existingUser.ID,
		// 	IPAddress: ipAddress,
		// 	UserAgent: userAgent,
		// 	Action:    "oauth_login",
		// 	Success:   true,
		// }
		// s.authLogRepo.Create(&authLog)

		return existingUser, token, nil
	}

	// User doesn't exist, create new user
	oauthID, _ := profile["sub"].(string)

	user := &models.User{
		Email:    email,
		Name:     name,
		OAuthID:  &oauthID,
		IsActive: true,
		RoleID:   2, // Default role
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", err
	}

	// Generate JWT token
	token, err := s.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	// Log authentication
	// authLog := models.AuthLog{
	// 	UserID:    user.ID,
	// 	IPAddress: ipAddress,
	// 	UserAgent: userAgent,
	// 	Action:    "oauth_registration",
	// 	Success:   true,
	// }
	// s.authLogRepo.Create(&authLog)

	return user, token, nil
}
