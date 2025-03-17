package repositories

import (
	"errors"
	"go-postgres-api/internal/database"
	"go-postgres-api/internal/models"
	"time"

	"gorm.io/gorm"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		db: database.GetDB(),
	}
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	result := r.db.Preload("Roles").Where("id = ?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, result.Error
	}
	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// AddRole adds a role to a user
func (r *UserRepository) AddRole(userID uint, roleType string) error {
	// Check if the role already exists
	var count int64
	r.db.Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.role_type = ?", userID, roleType).
		Count(&count)

	if count > 0 {
		return nil // Role already exists
	}

	// Find or create the role
	var role models.Role
	result := r.db.Where("role_type = ?", roleType).First(&role)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		role = models.Role{RoleType: roleType}
		if err := r.db.Create(&role).Error; err != nil {
			return err
		}
	} else if result.Error != nil {
		return result.Error
	}

	// Add the role to the user
	return r.db.Exec("INSERT INTO user_roles (user_id, role_id, created_at) VALUES (?, ?, ?)",
		userID, role.ID, time.Now()).Error
}

// UpdateLastLogin updates the user's last login time
func (r *UserRepository) UpdateLastLogin(userID uint) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("last_login", time.Now()).Error
}

// LogAuth logs an authentication attempt
func (r *UserRepository) LogAuth(log *models.AuthLog) error {
	return r.db.Create(log).Error
}

// BlacklistToken adds a token to the blacklist
func (r *UserRepository) BlacklistToken(blacklist *models.TokenBlacklist) error {
	return r.db.Create(blacklist).Error
}

// IsTokenBlacklisted checks if a token is blacklisted
func (r *UserRepository) IsTokenBlacklisted(tokenJTI string) (bool, error) {
	var count int64
	result := r.db.Model(&models.TokenBlacklist{}).
		Where("token_jti = ?", tokenJTI).
		Count(&count)
	return count > 0, result.Error
}