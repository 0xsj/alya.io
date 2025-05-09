
package domain

import (
	"time"
)

type UserRole string

const (
	UserRoleUser      UserRole = "user"
	UserRoleAdmin     UserRole = "admin"
	UserRolePremium   UserRole = "premium"
)

type User struct {
	ID             string    `json:"id" validate:"required"`
	Email          string    `json:"email" validate:"required,email"`
	Username       string    `json:"username" validate:"required"`
	HashedPassword string    `json:"-"`
	Name           string    `json:"name"`
	Role           UserRole  `json:"role" validate:"required,oneof=user admin premium"`
	AvatarURL      string    `json:"avatar_url"`
	APIKey         string    `json:"-"`
	LastLoginAt    time.Time `json:"last_login_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	
	PremiumExpiresAt  *time.Time `json:"premium_expires_at"`
	MonthlyQuota      int        `json:"monthly_quota"`
	RemainingQuota    int        `json:"remaining_quota"`
	QuotaResetAt      time.Time  `json:"quota_reset_at"`
}

type UserRepository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByAPIKey(apiKey string) (*User, error)
	Update(user *User) error
	UpdatePassword(id string, hashedPassword string) error
	UpdateAPIKey(id string, apiKey string) error
	DecrementQuota(id string) error
	Delete(id string) error
	List(page, pageSize int) ([]*User, int, error)
}

type UserService interface {
	Register(email, username, password string) (*User, error)
	Authenticate(emailOrUsername, password string) (*User, string, error) 
	GetProfile(id string) (*User, error)
	UpdateProfile(id string, updates map[string]any) (*User, error)
	ChangePassword(id string, currentPassword, newPassword string) error
	RegenerateAPIKey(id string) (string, error)
	HasAvailableQuota(id string) (bool, error)
}
