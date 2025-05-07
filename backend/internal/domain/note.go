// internal/domain/note.go
package domain

import (
	"time"
)

type Note struct {
	ID          string    `json:"id" validate:"required"`
	UserID      string    `json:"user_id" validate:"required"`
	VideoID     string    `json:"video_id" validate:"required"`
	Title       string    `json:"title"`
	Content     string    `json:"content" validate:"required"`
	Timestamp   float64   `json:"timestamp"` 
	IsPrivate   bool      `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type NoteRepository interface {
	Create(note *Note) error
	GetByID(id string) (*Note, error)
	Update(note *Note) error
	Delete(id string) error
	ListByUser(userID string, page, pageSize int) ([]*Note, int, error)
	ListByVideo(videoID string, includePrivate bool, page, pageSize int) ([]*Note, int, error)
	ListByUserAndVideo(userID, videoID string, page, pageSize int) ([]*Note, int, error)
	Search(query string, userID string, page, pageSize int) ([]*Note, int, error)
}

type NoteService interface {
	CreateNote(userID, videoID string, content string, timestamp float64, isPrivate bool) (*Note, error)
	GetNote(id string, userID string) (*Note, error)
	UpdateNote(id, userID string, updates map[string]any) (*Note, error)
	DeleteNote(id, userID string) error
	GetNotesByVideo(videoID string, userID string, page, pageSize int) ([]*Note, int, error)
}