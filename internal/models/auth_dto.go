package models

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the response for successful authentication
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         User   `json:"user"`
}

// RefreshTokenRequest represents the request body for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// EmailVerificationRequest represents the request for email verification
type EmailVerificationRequest struct {
	Token string `json:"token" binding:"required"`
}

// ResendVerificationRequest represents the request to resend verification email
type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}
