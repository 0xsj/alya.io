
// internal/domain/notification.go
package domain

import (
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeInfo     NotificationType = "info"
	NotificationTypeWarning  NotificationType = "warning"
	NotificationTypeError    NotificationType = "error"
	NotificationTypeSuccess  NotificationType = "success"
)

// NotificationChannel represents the delivery channel for notifications
type NotificationChannel string

const (
	ChannelInApp   NotificationChannel = "in_app"
	ChannelEmail   NotificationChannel = "email"
	ChannelSMS     NotificationChannel = "sms"
	ChannelPush    NotificationChannel = "push"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusUnread   NotificationStatus = "unread"
	NotificationStatusRead     NotificationStatus = "read"
	NotificationStatusArchived NotificationStatus = "archived"
)

// Notification represents a notification sent to a user
type Notification struct {
	ID          string             `json:"id" validate:"required"`
	UserID      string             `json:"user_id" validate:"required"`
	Type        NotificationType   `json:"type" validate:"required"`
	Title       string             `json:"title" validate:"required"`
	Message     string             `json:"message" validate:"required"`
	ResourceID  string             `json:"resource_id,omitempty"`    // Related resource ID
	ResourceURL string             `json:"resource_url,omitempty"`   // URL to related resource
	Channel     NotificationChannel `json:"channel" validate:"required"`
	Status      NotificationStatus  `json:"status" validate:"required"`
	ReadAt      *time.Time          `json:"read_at,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	ExpiresAt   *time.Time          `json:"expires_at,omitempty"`
}

// NotificationTemplate represents a reusable template for notifications
type NotificationTemplate struct {
	ID          string             `json:"id" validate:"required"`
	Code        string             `json:"code" validate:"required"`      // Template identifier
	Type        NotificationType   `json:"type" validate:"required"`
	Title       string             `json:"title" validate:"required"`     // Can contain placeholders
	Message     string             `json:"message" validate:"required"`   // Can contain placeholders
	Channels    []NotificationChannel `json:"channels" validate:"required,min=1"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// NotificationRepository defines the interface for notification storage operations
type NotificationRepository interface {
	Create(notification *Notification) error
	GetByID(id string) (*Notification, error)
	MarkAsRead(id string) error
	MarkAllAsRead(userID string) error
	ArchiveNotification(id string) error
	DeleteNotification(id string) error
	ListByUser(userID string, status NotificationStatus, page, pageSize int) ([]*Notification, int, error)
	GetUnreadCount(userID string) (int, error)
	
	// Template operations
	CreateTemplate(template *NotificationTemplate) error
	GetTemplateByCode(code string) (*NotificationTemplate, error)
	UpdateTemplate(template *NotificationTemplate) error
	DeleteTemplate(id string) error
	ListTemplates() ([]*NotificationTemplate, error)
}

// NotificationService defines high-level operations for notifications
type NotificationService interface {
	SendNotification(userID string, templateCode string, data map[string]any, channels []NotificationChannel) error
	SendCustomNotification(userID string, title, message string, notificationType NotificationType, channels []NotificationChannel) error
	GetUserNotifications(userID string, status NotificationStatus, page, pageSize int) ([]*Notification, int, error)
	MarkNotificationAsRead(id string, userID string) error
	MarkAllNotificationsAsRead(userID string) error
	GetUnreadNotificationCount(userID string) (int, error)
}
