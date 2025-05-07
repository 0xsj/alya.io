// internal/domain/tag.go
package domain

import (
	"time"
)

type Tag struct {
	ID          string    `json:"id" validate:"required"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type VideoTag struct {
	VideoID   string    `json:"video_id" validate:"required"`
	TagID     string    `json:"tag_id" validate:"required"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type TagRepository interface {
	Create(tag *Tag) error
	GetByID(id string) (*Tag, error)
	GetByName(name string) (*Tag, error)
	Update(tag *Tag) error
	Delete(id string) error
	List(page, pageSize int) ([]*Tag, int, error)
	AddTagToVideo(videoID, tagID, userID string) error
	RemoveTagFromVideo(videoID, tagID string) error
	GetTagsByVideo(videoID string) ([]*Tag, error)
	GetVideosByTag(tagID string, page, pageSize int) ([]*Video, int, error)
}

type TagService interface {
	CreateTag(name, description, userID string) (*Tag, error)
	UpdateTag(id, name, description, userID string) (*Tag, error)
	DeleteTag(id, userID string) error
	AddTagToVideo(videoID, tagName, userID string) error
	RemoveTagFromVideo(videoID, tagName, userID string) error
	GetPopularTags(limit int) ([]*Tag, error)
}