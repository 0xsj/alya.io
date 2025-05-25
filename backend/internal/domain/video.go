// internal/domain/video.go - Fixed with db tags
package domain

import (
	"time"

	"github.com/lib/pq"
)

type VideoStatus string

const (
	VideoStatusPending    VideoStatus = "pending"
	VideoStatusProcessing VideoStatus = "processing"
	VideoStatusCompleted  VideoStatus = "completed"
	VideoStatusFailed     VideoStatus = "failed"
)

type VideoVisibility string

const (
	VideoVisibilityPublic  VideoVisibility = "public"
	VideoVisibilityPrivate VideoVisibility = "private"
)

type Video struct {
	ID            string          `json:"id" db:"id" validate:"required"`
	YouTubeID     string          `json:"youtube_id" db:"youtube_id" validate:"required"`
	Title         string          `json:"title" db:"title" validate:"required"`
	Description   *string         `json:"description" db:"description"`
	URL           string          `json:"url" db:"url" validate:"required,url"`
	ThumbnailURL  *string         `json:"thumbnail_url" db:"thumbnail_url"`
	Status        VideoStatus     `json:"status" db:"status" validate:"required,oneof=pending processing completed failed"`
	Visibility    VideoVisibility `json:"visibility" db:"visibility" validate:"required,oneof=public private"`
	Duration      *int64          `json:"duration" db:"duration"`
	Language      *string         `json:"language" db:"language"`
	TranscriptID  *string         `json:"transcript_id" db:"transcript_id"`
	SummaryID     *string         `json:"summary_id" db:"summary_id"`
	Tags          pq.StringArray  `json:"tags" db:"tags"`
	Channel       *string         `json:"channel" db:"channel"`
	ChannelID     *string         `json:"channel_id" db:"channel_id"`
	Views         *int64          `json:"views" db:"views"`
	LikeCount     *int64          `json:"like_count" db:"like_count"`
	CommentCount  *int64          `json:"comment_count" db:"comment_count"`
	PublishedAt   *time.Time      `json:"published_at" db:"published_at"`
	ProcessedAt   *time.Time      `json:"processed_at" db:"processed_at"`
	ErrorMessage  *string         `json:"error_message" db:"error_message"`
	CreatedBy     string          `json:"created_by" db:"created_by"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
	TsvDocument   string          `json:"-" db:"tsv_document"`
}

type VideoRepository interface {
	Create(video *Video) error
	GetByID(id string) (*Video, error)
	GetByYouTubeID(youtubeID string) (*Video, error)
	Update(video *Video) error
	UpdateStatus(id string, status VideoStatus, errorMessage *string) error
	UpdateProcessingResults(id string, transcriptID *string, summaryID *string) error
	Delete(id string) error
	List(page, pageSize int, filters map[string]any) ([]*Video, int, error)
	ListByUserID(userID string, page, pageSize int) ([]*Video, int, error)
	ListByStatus(status VideoStatus, limit int) ([]*Video, error)
	Search(query string, page, pageSize int) ([]*Video, int, error)
}

type VideoService interface {
	ProcessVideo(youtubeURL string, userID string) (*Video, error)
	GetVideoDetails(id string, userID string) (*Video, error)
	SearchVideos(query string, page, pageSize int, userID string) ([]*Video, int, error)
	DeleteVideo(id string, userID string) error
}