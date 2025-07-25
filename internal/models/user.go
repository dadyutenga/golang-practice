package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Name      string    `json:"name" gorm:"not null"`
	Password  *string   `json:"-" gorm:"null"`        // Make password nullable for OAuth users
	OAuthID   *string   `json:"oauth_id" gorm:"null"` // Add OAuth ID field
	RoleID    uint      `json:"role_id" gorm:"not null;default:2"`
	Role      Role      `json:"role" gorm:"foreignKey:RoleID"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Role represents a user role
type Role struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	RoleType  string    `json:"role_type" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
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
	CreatedAt    time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TokenBlacklist represents a blacklisted JWT token
type TokenBlacklist struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TokenJTI  string    `json:"token_jti" gorm:"uniqueIndex;not null"`
	UserID    uint      `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	passwordStr := string(hashedPassword)
	u.Password = &passwordStr
	return nil
}

// CheckPassword verifies the user's password
func (u *User) CheckPassword(password string) bool {
	if u.Password == nil {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(password))
	return err == nil
}

// BeforeCreate is a GORM hook that runs before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a user
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}
