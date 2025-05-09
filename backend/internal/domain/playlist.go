
// internal/domain/playlist.go
package domain

import (
	"time"
)

type PlaylistVisibility string

const (
	PlaylistVisibilityPublic  PlaylistVisibility = "public"
	PlaylistVisibilityPrivate PlaylistVisibility = "private"
	PlaylistVisibilityUnlisted PlaylistVisibility = "unlisted"
)

type Playlist struct {
	ID          string             `json:"id" validate:"required"`
	UserID      string             `json:"user_id" validate:"required"`
	Title       string             `json:"title" validate:"required"`
	Description string             `json:"description"`
	Visibility  PlaylistVisibility `json:"visibility" validate:"required,oneof=public private unlisted"`
	VideoCount  int                `json:"video_count"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type PlaylistItem struct {
	ID         string    `json:"id" validate:"required"`
	PlaylistID string    `json:"playlist_id" validate:"required"`
	VideoID    string    `json:"video_id" validate:"required"`
	Position   int       `json:"position"` 
	AddedAt    time.Time `json:"added_at"`
}

type PlaylistRepository interface {
	Create(playlist *Playlist) error
	GetByID(id string) (*Playlist, error)
	Update(playlist *Playlist) error
	Delete(id string) error
	ListByUser(userID string, page, pageSize int) ([]*Playlist, int, error)
	
	AddVideo(playlistID, videoID string, position int) error
	RemoveVideo(playlistID, videoID string) error
	UpdateVideoPosition(playlistID, videoID string, newPosition int) error
	ListVideos(playlistID string, page, pageSize int) ([]*Video, int, error)
}

type PlaylistService interface {
	CreatePlaylist(userID, title, description string, visibility PlaylistVisibility) (*Playlist, error)
	GetPlaylist(id string, userID string) (*Playlist, error)
	UpdatePlaylist(id, userID string, updates map[string]any) (*Playlist, error)
	DeletePlaylist(id, userID string) error
	
	AddVideoToPlaylist(playlistID, videoID, userID string) error
	RemoveVideoFromPlaylist(playlistID, videoID, userID string) error
	ReorderPlaylistVideos(playlistID string, videoIDs []string, userID string) error
	GetPlaylistVideos(playlistID string, userID string, page, pageSize int) ([]*Video, int, error)
}
