package service

import (
	"time"

	"github.com/0xsj/alya.io/backend/internal/domain"
	"github.com/0xsj/alya.io/backend/pkg/errors"
	"github.com/0xsj/alya.io/backend/pkg/logger"
	"github.com/google/uuid"
)

type VideoService struct {
	repo   domain.VideoRepository
	logger logger.Logger
}

// DeleteVideo implements domain.VideoService.
func (s *VideoService) DeleteVideo(id string, userID string) error {
	panic("unimplemented")
}

// SearchVideos implements domain.VideoService.
func (s *VideoService) SearchVideos(query string, page int, pageSize int, userID string) ([]*domain.Video, int, error) {
	panic("unimplemented")
}

func NewVideoService(repo domain.VideoRepository, logger logger.Logger) *VideoService {
	return &VideoService{
		repo:   repo,
		logger: logger.WithLayer("service.video"),
	}
}

func (s *VideoService) ProcessVideo(youtubeURL string, userID string) (*domain.Video, error) {
	// Extract YouTube ID from URL
	youtubeID, err := extractYouTubeID(youtubeURL)
	if err != nil {
		return nil, errors.NewInvalidURLError("invalid YouTube URL", err)
	}

	// Check if video already exists
	existingVideo, err := s.repo.GetByYouTubeID(youtubeID)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	if existingVideo != nil {
		return existingVideo, nil
	}

	// Create a new video record
	now := time.Now()
	video := &domain.Video{
		ID:         uuid.New().String(),
		YouTubeID:  youtubeID,
		Title:      "Pending Processing", // Will be updated later
		URL:        youtubeURL,
		Status:     domain.VideoStatusPending,
		Visibility: domain.VideoVisibilityPublic,
		CreatedBy:  userID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Save to database
	if err := s.repo.Create(video); err != nil {
		return nil, err
	}

	// Start background processing (in a real app)
	// go s.processVideoAsync(video.ID)

	return video, nil
}

func (s *VideoService) GetVideoDetails(id string, userID string) (*domain.Video, error) {
	video, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if video.Visibility == domain.VideoVisibilityPrivate && video.CreatedBy != userID {
		return nil, errors.NewForbiddenError("you don't have permission to access this video", nil)
	}

	return video, nil
}

func extractYouTubeID(url string) (string, error) {
	// This is a simplified version - you'll want more robust parsing
	// For example: youtube.com/watch?v=AbCdEf123
	// Simplified for demonstration
	return "dQw4w9WgXcQ", nil
}
