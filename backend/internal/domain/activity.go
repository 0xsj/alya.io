// internal/domain/activity.go
package domain

import (
	"time"
)

type ActivityType string

const (
	ActivityTypeVideoView       ActivityType = "video_view"
	ActivityTypeVideoSummary    ActivityType = "video_summary"
	ActivityTypeNoteCreate      ActivityType = "note_create"
	ActivityTypeNoteEdit        ActivityType = "note_edit"
	ActivityTypeBookmarkCreate  ActivityType = "bookmark_create"
	ActivityTypePlaylistCreate  ActivityType = "playlist_create"
	ActivityTypePlaylistEdit    ActivityType = "playlist_edit"
	ActivityTypeVideoAdd        ActivityType = "video_add"
	ActivityTypeSearchQuery     ActivityType = "search_query"
	ActivityTypeLogin           ActivityType = "login"
	ActivityTypeAccountUpdate   ActivityType = "account_update"
)

type Activity struct {
	ID           string       `json:"id" validate:"required"`
	UserID       string       `json:"user_id" validate:"required"`
	Type         ActivityType `json:"type" validate:"required"`
	ResourceID   string       `json:"resource_id"`          // Related resource ID (video, note, etc.)
	ResourceType string       `json:"resource_type"`        // Type of related resource
	Metadata     any		  `json:"metadata,omitempty"`   // Additional context data
	IP           string       `json:"ip,omitempty"`         // User's IP address
	UserAgent    string       `json:"user_agent,omitempty"` // User's browser/device info
	CreatedAt    time.Time    `json:"created_at"`
}

type ActivityRepository interface {
	Create(activity *Activity) error
	GetByID(id string) (*Activity, error)
	ListByUser(userID string, page, pageSize int) ([]*Activity, int, error)
	ListByType(activityType ActivityType, page, pageSize int) ([]*Activity, int, error)
	ListByUserAndType(userID string, activityType ActivityType, page, pageSize int) ([]*Activity, int, error)
	ListByResource(resourceType, resourceID string, page, pageSize int) ([]*Activity, int, error)
	GetUserStats(userID string) (map[ActivityType]int, error)
}

type ActivityService interface {
	LogActivity(userID string, activityType ActivityType, resourceType, resourceID string, metadata interface{}, clientInfo map[string]string) error
	GetUserActivities(userID string, page, pageSize int) ([]*Activity, int, error)
	GetUserActivityHistory(userID string, days int) (map[string][]Activity, error)
}