package domain

import (
	"time"
)

type SummaryType string

const (
	SummaryTypeShort     SummaryType = "short"
	SummaryTypeLong      SummaryType = "long"
	SummaryTypeSegmented SummaryType = "segmented"
)

type SummarySegment struct {
	Index      int     `json:"index"`
	Start      float64 `json:"start"`  
	End        float64 `json:"end"`     
	Title      string  `json:"title"`    
	Content    string  `json:"content"`  
}

type Summary struct {
	ID             string           `json:"id" validate:"required"`
	VideoID        string           `json:"video_id" validate:"required"`
	Type           SummaryType      `json:"type" validate:"required,oneof=short long segmented"`
	ShortSummary   string           `json:"short_summary"` // 1-2 sentence summary
	LongSummary    string           `json:"long_summary"`  // Detailed multi-paragraph summary
	Segments       []SummarySegment `json:"segments"`      // For segmented summaries
	KeyPoints      []string         `json:"key_points"`    // List of key points
	Keywords       []string         `json:"keywords"`      // List of keywords
	Topics         []string         `json:"topics"`        // List of identified topics
	Questions      []string         `json:"questions"`     // List of generated questions from the content
	AIModel        string           `json:"ai_model"`      // AI model used for generation
	Language       string           `json:"language"`      // ISO 639-1 code
	ProcessingTime int              `json:"processing_time"` // Processing time in milliseconds
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

type SummaryRepository interface {
	Create(summary *Summary) error
	GetByID(id string) (*Summary, error)
	GetByVideoID(videoID string) (*Summary, error)
	Update(summary *Summary) error
	Delete(id string) error
	Search(query string, page, pageSize int) ([]*Summary, int, error)
}

type SummaryService interface {
	GetSummary(id string, userID string) (*Summary, error)
	GetSummaryByVideoID(videoID string, userID string) (*Summary, error)
	RegenerateSummary(videoID string, userID string) (*Summary, error)
}