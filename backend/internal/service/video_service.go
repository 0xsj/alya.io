// internal/service/video_service.go
package service

import (
	"net/url"
	"regexp"
	"strings"
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
		// If the video exists but is in a failed state, we could try to process it again
		if existingVideo.Status == domain.VideoStatusFailed {
			// Reset the status to pending
			err := s.repo.UpdateStatus(existingVideo.ID, domain.VideoStatusPending, "")
			if err != nil {
				return nil, err
			}
			existingVideo.Status = domain.VideoStatusPending
			existingVideo.ErrorMessage = ""
			// In a real app, you'd then start the processing in a background goroutine
			// go s.processVideoAsync(existingVideo.ID)
		}
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

	// In a real implementation, start the processing in a background goroutine
	// go s.processVideoAsync(video.ID)

	return video, nil
}

func (s *VideoService) GetVideoDetails(id string, userID string) (*domain.Video, error) {
	video, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check visibility - private videos can only be accessed by their creator
	if video.Visibility == domain.VideoVisibilityPrivate && video.CreatedBy != userID {
		return nil, errors.NewForbiddenError("you don't have permission to access this video", nil)
	}

	return video, nil
}

func (s *VideoService) SearchVideos(query string, page int, pageSize int, userID string) ([]*domain.Video, int, error) {
	// Validate input
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Search for videos and limit visibility
	filters := map[string]any{
		"search": query,
	}
	
	// We could add additional filters like "visibility = public OR created_by = userID"
	// but let's keep it simple for now and filter in memory
	
	videos, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}

	// Filter out private videos not owned by the user
	filteredVideos := make([]*domain.Video, 0, len(videos))
	for _, video := range videos {
		if video.Visibility == domain.VideoVisibilityPublic || video.CreatedBy == userID {
			filteredVideos = append(filteredVideos, video)
		}
	}

	return filteredVideos, total, nil
}

func (s *VideoService) DeleteVideo(id string, userID string) error {
	// Get the video first to check permissions
	video, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Only the creator can delete a video
	if video.CreatedBy != userID {
		return errors.NewForbiddenError("you don't have permission to delete this video", nil)
	}

	// Delete the video
	return s.repo.Delete(id)
}

// processVideoAsync is a helper function that would process the video in the background
// In a real implementation, this would be part of a worker or background job system
func (s *VideoService) processVideoAsync(videoID string) {
	// This would be a much more sophisticated implementation in a real system
	// For now, just simulate the process

	// 1. Update status to processing
	err := s.repo.UpdateStatus(videoID, domain.VideoStatusProcessing, "")
	if err != nil {
		s.logger.Errorf("Failed to update video status to processing: %v", err)
		return
	}

	// 2. Fetch the video metadata from YouTube API
	// This is where you'd use an external YouTube API client
	// For now, we'll simulate with arbitrary values
	video, err := s.repo.GetByID(videoID)
	if err != nil {
		s.logger.Errorf("Failed to get video by ID: %v", err)
		return
	}

	video.Title = "Sample YouTube Video"
	video.Description = "This is a sample description for the video."
	video.ThumbnailURL = "https://img.youtube.com/vi/" + video.YouTubeID + "/maxresdefault.jpg"
	video.Duration = 600 
	video.Language = "en"
	video.Channel = "Sample Channel"
	video.ChannelID = "UCxxxx"
	video.Views = 1000
	video.LikeCount = 100
	video.CommentCount = 50
	video.PublishedAt = time.Now().Add(-24 * time.Hour) 
	
	err = s.repo.Update(video)
	if err != nil {
		s.logger.Errorf("Failed to update video metadata: %v", err)
		s.repo.UpdateStatus(videoID, domain.VideoStatusFailed, "Failed to update video metadata")
		return
	}

	transcriptID := uuid.New().String()
	
	summaryID := uuid.New().String()
	
	err = s.repo.UpdateProcessingResults(videoID, transcriptID, summaryID)
	if err != nil {
		s.logger.Errorf("Failed to update processing results: %v", err)
		s.repo.UpdateStatus(videoID, domain.VideoStatusFailed, "Failed to update processing results")
		return
	}
	
	s.logger.Infof("Successfully processed video: %s", videoID)
}

func extractYouTubeID(youtubeURL string) (string, error) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:youtube\.com\/watch\?v=|youtu\.be\/)([^&?\/]+)`),
		regexp.MustCompile(`youtube\.com\/embed\/([^\/\?]+)`),
		regexp.MustCompile(`youtube\.com\/v\/([^\/\?]+)`),
		regexp.MustCompile(`youtube\.com\/shorts\/([^\/\?]+)`),
	}

	youtubeURL = strings.TrimSpace(youtubeURL)

	_, err := url.ParseRequestURI(youtubeURL)
	if err != nil {
		return "", errors.Wrap(err, "invalid URL format")
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(youtubeURL)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", errors.NewInvalidURLError("could not extract YouTube video ID from URL", nil)
}