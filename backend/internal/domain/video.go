
package domain

import "time"

type VideoStatus string

const (
	VideoStatusPending     VideoStatus = "pending"
	VideoStatusProcessing  VideoStatus = "processing"
	VideoStatusCompleted   VideoStatus = "completed"
	VideoStatusFailed      VideoStatus = "failed"
)

type VideoVisibility string

const (
	VideoVisibilityPublic  VideoVisibility = "public"
	VideoVisibilityPrivate VideoVisibility = "private"
)

type Video struct {
	ID            string          `json:"id" validate:"required"`
	YouTubeID     string          `json:"youtube_id" validate:"required"`
	Title         string          `json:"title" validate:"required"`
	Description   string          `json:"description"`
	URL           string          `json:"url" validate:"required,url"`
	ThumbnailURL  string          `json:"thumbnail_url" validate:"url"`
	Status        VideoStatus     `json:"status" validate:"required,oneof=pending processing completed failed"`
	Visibility    VideoVisibility `json:"visibility" validate:"required,oneof=public private"`
	Duration      int64           `json:"duration"`          
	Language      string          `json:"language"`          
	TranscriptID  string          `json:"transcript_id"`
	SummaryID     string          `json:"summary_id"`
	Tags          []string        `json:"tags"`
	Channel       string          `json:"channel"`           
	ChannelID     string          `json:"channel_id"`        
	Views         int64           `json:"views"`             
	LikeCount     int64           `json:"like_count"`        
	CommentCount  int64           `json:"comment_count"`     
	PublishedAt   time.Time       `json:"published_at"`      
	ProcessedAt   *time.Time      `json:"processed_at"`      
	ErrorMessage  string          `json:"error_message"`     
	CreatedBy     string          `json:"created_by"`        
	CreatedAt     time.Time       `json:"created_at"`        
	UpdatedAt     time.Time       `json:"updated_at"`        
}

type VideoRepository interface {
	Create(video *Video) error
	GetByID(id string) (*Video, error)
	GetByYouTubeID(youtubeID string) (*Video, error)
	Update(video *Video) error
	UpdateStatus(id string, status VideoStatus, errorMessage string) error
	UpdateProcessingResults(id string, transcriptID string, summaryID string) error
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
