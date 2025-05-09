
package domain

import (
	"time"
)

type TranscriptSegment struct {
	Index      int     `json:"index"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`  
	Text       string  `json:"text"`  
	Speaker    string  `json:"speaker"` 
	Confidence float64 `json:"confidence"`
}

type Transcript struct {
	ID          string              `json:"id" validate:"required"`
	VideoID     string              `json:"video_id" validate:"required"`
	Language    string              `json:"language"`  
	Segments    []TranscriptSegment `json:"segments"`
	RawText     string              `json:"raw_text"`  
	Source      string              `json:"source"`    
	ProcessedAt time.Time           `json:"processed_at"`
	CreatedAt   time.Time           `json:"created_at"`
}

type TranscriptRepository interface {
	Create(transcript *Transcript) error
	GetByID(id string) (*Transcript, error)
	GetByVideoID(videoID string) (*Transcript, error)
	Update(transcript *Transcript) error
	Delete(id string) error
	Search(query string, page, pageSize int) ([]*Transcript, int, error)
}

type TranscriptService interface {
	GetTranscript(id string, userID string) (*Transcript, error)
	GetTranscriptByVideoID(videoID string, userID string) (*Transcript, error)
	SearchTranscripts(query string, page, pageSize int, userID string) ([]*Transcript, int, error)
}
