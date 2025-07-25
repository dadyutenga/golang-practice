package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Email      string    `json:"email" gorm:"unique;not null"`
	Name       string    `json:"name" gorm:"not null"`
	Password   string    `json:"-" gorm:"not null"`
	IsVerified bool      `json:"is_verified" gorm:"default:false"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	RoleID     uint      `json:"role_id" gorm:"not null;default:2"`
	Role       Role      `json:"role" gorm:"foreignKey:RoleID"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Role represents a user role
type Role struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	RoleType  string    `json:"role_type" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UserID    uint      `json:"-" gorm:"index"`
}

// AuthLog represents an authentication log entry
type AuthLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id"`
	Action       string    `json:"action" gorm:"not null"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	Success      bool      `json:"success" gorm:"not null"`
	ErrorMessage string    `json:"error_message"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TokenBlacklist represents a blacklisted JWT token
type TokenBlacklist struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TokenJTI  string    `json:"token_jti" gorm:"uniqueIndex;not null"`
	UserID    uint      `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Used      bool      `json:"used" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Used      bool      `json:"used" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifies the user's password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
