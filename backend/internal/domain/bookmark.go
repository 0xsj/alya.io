// internal/domain/bookmark.go
package domain

import (
	"time"
)

type Bookmark struct {
	ID          string    `json:"id" validate:"required"`
	UserID      string    `json:"user_id" validate:"required"`
	VideoID     string    `json:"video_id" validate:"required"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Timestamp   float64   `json:"timestamp"` 
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BookmarkRepository interface {
	Create(bookmark *Bookmark) error
	GetByID(id string) (*Bookmark, error)
	GetByUserAndVideo(userID, videoID string) (*Bookmark, error)
	Update(bookmark *Bookmark) error
	Delete(id string) error
	ListByUser(userID string, page, pageSize int) ([]*Bookmark, int, error)
	ListByVideo(videoID string, page, pageSize int) ([]*Bookmark, int, error)
}

type BookmarkService interface {
	CreateBookmark(userID, videoID string, timestamp float64, title, description string) (*Bookmark, error)
	GetBookmarks(userID string, page, pageSize int) ([]*Bookmark, int, error)
	UpdateBookmark(id, userID string, updates map[string]any) (*Bookmark, error)
	DeleteBookmark(id, userID string) error
}