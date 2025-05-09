
// internal/domain/search.go
package domain

import (
	"time"
)

// SearchResultType represents the type of search result
type SearchResultType string

const (
	SearchResultTypeVideo     SearchResultType = "video"
	SearchResultTypeTranscript SearchResultType = "transcript"
	SearchResultTypeSummary   SearchResultType = "summary"
	SearchResultTypeNote      SearchResultType = "note"
	SearchResultTypePlaylist  SearchResultType = "playlist"
	SearchResultTypeUser      SearchResultType = "user"
)

// SearchResult represents a single search result
type SearchResult struct {
	ID            string           `json:"id"`
	Type          SearchResultType `json:"type"`
	Title         string           `json:"title"`
	Description   string           `json:"description,omitempty"`
	ResourceID    string           `json:"resource_id"` // The actual ID of the resource
	URL           string           `json:"url,omitempty"`
	ThumbnailURL  string           `json:"thumbnail_url,omitempty"`
	Relevance     float64          `json:"relevance"` // Relevance score (0-1)
	Highlights    map[string][]string `json:"highlights,omitempty"` // Highlighted snippets from search
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// SearchQuery represents a user's search query and metadata
type SearchQuery struct {
	ID         string    `json:"id" validate:"required"`
	UserID     string    `json:"user_id"`
	Query      string    `json:"query" validate:"required"`
	Filters    map[string]any `json:"filters,omitempty"`
	ResultCount int       `json:"result_count"`
	ClientInfo map[string]string      `json:"client_info,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// SearchRepository defines the interface for search operations
type SearchRepository interface {
	Search(query string, filters map[string]any, page, pageSize int, userID string) ([]SearchResult, int, error)
	IndexVideo(video *Video) error
	IndexTranscript(transcript *Transcript) error
	IndexSummary(summary *Summary) error
	IndexNote(note *Note) error
	UpdateIndex(resourceType string, resourceID string) error
	RemoveFromIndex(resourceType string, resourceID string) error
	LogSearchQuery(query *SearchQuery) error
	GetPopularSearches(limit int, timeRange time.Duration) ([]string, error)
}

// SearchService defines high-level operations for search functionality
type SearchService interface {
	Search(query string, filters map[string]any, page, pageSize int, userID string) ([]SearchResult, int, error)
	SearchTranscripts(query string, page, pageSize int, userID string) ([]SearchResult, int, error)
	SearchVideos(query string, page, pageSize int, userID string) ([]SearchResult, int, error)
	GetSearchHistory(userID string, limit int) ([]SearchQuery, error)
	GetSearchSuggestions(partialQuery string, userID string) ([]string, error)
}
