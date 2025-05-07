// internal/domain/settings.go
package domain

import (
	"time"
)

// SettingScope represents the scope of a setting
type SettingScope string

const (
	SettingScopeUser   SettingScope = "user"
	SettingScopeSystem SettingScope = "system"
)

// SettingDataType represents the data type of a setting value
type SettingDataType string

const (
	SettingTypeString  SettingDataType = "string"
	SettingTypeNumber  SettingDataType = "number"
	SettingTypeBoolean SettingDataType = "boolean"
	SettingTypeJSON    SettingDataType = "json"
)

// Setting represents a user or system setting
type Setting struct {
	ID          string         `json:"id" validate:"required"`
	Key         string         `json:"key" validate:"required"`         // Setting identifier
	Value       any			   `json:"value"`                           // Setting value (type depends on DataType)
	DataType    SettingDataType `json:"data_type" validate:"required"`
	Scope       SettingScope   `json:"scope" validate:"required"`
	UserID      string         `json:"user_id,omitempty"`               // Only for user settings
	Description string         `json:"description"`
	IsDefault   bool           `json:"is_default"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// UserPreferences represents user-specific preferences
type UserPreferences struct {
	ID                   string    `json:"id" validate:"required"`
	UserID               string    `json:"user_id" validate:"required"`
	Theme                string    `json:"theme" validate:"required"`           // light, dark, system
	EmailNotifications   bool      `json:"email_notifications"`
	PushNotifications    bool      `json:"push_notifications"`
	DefaultVideoQuality  string    `json:"default_video_quality"`               // auto, high, medium, low
	AutoplaySummaries    bool      `json:"autoplay_summaries"`
	ShowTranscript       bool      `json:"show_transcript"`
	Language             string    `json:"language" validate:"required"`        // ISO language code
	TimeZone             string    `json:"time_zone"`
	DateFormat           string    `json:"date_format"`
	TimeFormat           string    `json:"time_format"`                         // 12h or 24h
	CustomSettings       map[string]any `json:"custom_settings,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// SystemConfig represents system-wide configuration
type SystemConfig struct {
	ID                string    `json:"id" validate:"required"`
	Key               string    `json:"key" validate:"required"`
	Value             any		`json:"value"`
	DataType          SettingDataType `json:"data_type" validate:"required"`
	Description       string    `json:"description"`
	IsEncrypted       bool      `json:"is_encrypted"`
	RequiresRestart   bool      `json:"requires_restart"`
	LastModifiedBy    string    `json:"last_modified_by"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// FeatureFlag represents a system feature flag for progressive rollouts
type FeatureFlag struct {
	ID              string    `json:"id" validate:"required"`
	Name            string    `json:"name" validate:"required"`
	Description     string    `json:"description"`
	Enabled         bool      `json:"enabled"`
	UserPercentage  int       `json:"user_percentage"`            // 0-100 percentage for gradual rollout
	AllowedUserIDs  []string  `json:"allowed_user_ids,omitempty"` // Specific users with access
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
}

// SettingsRepository defines the interface for settings storage operations
type SettingsRepository interface {
	// General settings
	GetSetting(key string, scope SettingScope, userID string) (*Setting, error)
	SetSetting(setting *Setting) error
	DeleteSetting(key string, scope SettingScope, userID string) error
	ListSettings(scope SettingScope, userID string) ([]*Setting, error)
	
	// User preferences
	GetUserPreferences(userID string) (*UserPreferences, error)
	UpdateUserPreferences(prefs *UserPreferences) error
	
	// System config
	GetSystemConfig(key string) (*SystemConfig, error)
	SetSystemConfig(config *SystemConfig) error
	ListSystemConfig() ([]*SystemConfig, error)
	
	// Feature flags
	GetFeatureFlag(name string) (*FeatureFlag, error)
	SetFeatureFlag(flag *FeatureFlag) error
	IsFeatureEnabledForUser(flagName string, userID string) (bool, error)
	ListFeatureFlags() ([]*FeatureFlag, error)
}

// SettingsService defines high-level operations for settings management
type SettingsService interface {
	// User settings
	GetUserSetting(userID string, key string) (any, error)
	SetUserSetting(userID string, key string, value any) error
	GetUserPreferences(userID string) (*UserPreferences, error)
	UpdateUserPreferences(userID string, updates map[string]any) (*UserPreferences, error)
	
	// System settings
	GetSystemSetting(key string) (any, error)
	SetSystemSetting(key string, value any, adminID string) error
	
	// Feature flags
	IsFeatureEnabled(featureName string, userID string) (bool, error)
	EnableFeature(featureName string, adminID string) error
	DisableFeature(featureName string, adminID string) error
	RolloutFeature(featureName string, percentage int, adminID string) error
}