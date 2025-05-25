// internal/service/video_service.go - Fixed for pointer fields
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
	repo               domain.VideoRepository
	transcriptService  *TranscriptService
	logger             logger.Logger
}

func NewVideoService(
	repo domain.VideoRepository, 
	transcriptService *TranscriptService,
	logger logger.Logger,
) *VideoService {
	return &VideoService{
		repo:              repo,
		transcriptService: transcriptService,
		logger:            logger.WithLayer("service.video"),
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
			err := s.repo.UpdateStatus(existingVideo.ID, domain.VideoStatusPending, nil)
			if err != nil {
				return nil, err
			}
			existingVideo.Status = domain.VideoStatusPending
			existingVideo.ErrorMessage = nil
			// Start processing in background
			go s.processVideoAsync(existingVideo.ID)
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

	// Start processing in background
	go s.processVideoAsync(video.ID)

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

	// Delete the video (transcript will be cascade deleted due to foreign key)
	return s.repo.Delete(id)
}

// GetVideoWithTranscript returns video details along with transcript
func (s *VideoService) GetVideoWithTranscript(id string, userID string) (*domain.Video, *domain.Transcript, error) {
	// Get video details
	video, err := s.GetVideoDetails(id, userID)
	if err != nil {
		return nil, nil, err
	}

	// Get transcript (will extract if doesn't exist)
	transcript, err := s.transcriptService.GetTranscriptByVideoID(video.ID, userID)
	if err != nil {
		s.logger.Warn("Failed to get transcript", "video_id", video.ID, "error", err)
		// Return video without transcript rather than failing completely
		return video, nil, nil
	}

	return video, transcript, nil
}

// Helper functions to create pointers
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// processVideoAsync handles video processing in the background
func (s *VideoService) processVideoAsync(videoID string) {
	s.logger.Info("Starting video processing", "video_id", videoID)

	// Update status to processing
	err := s.repo.UpdateStatus(videoID, domain.VideoStatusProcessing, nil)
	if err != nil {
		s.logger.Errorf("Failed to update video status to processing: %v", err)
		return
	}

	// Get video details
	video, err := s.repo.GetByID(videoID)
	if err != nil {
		s.logger.Errorf("Failed to get video by ID: %v", err)
		s.repo.UpdateStatus(videoID, domain.VideoStatusFailed, stringPtr("Failed to retrieve video details"))
		return
	}

	// Step 1: Extract basic video metadata (simulated for now)
	// In a real implementation, you would use YouTube API here
	video.Title = "Sample YouTube Video"
	video.Description = stringPtr("This is a sample description for the video.")
	video.ThumbnailURL = stringPtr("https://img.youtube.com/vi/" + video.YouTubeID + "/maxresdefault.jpg")
	video.Duration = int64Ptr(600)
	video.Language = stringPtr("en")
	video.Channel = stringPtr("Sample Channel")
	video.ChannelID = stringPtr("UCxxxx")
	video.Views = int64Ptr(1000)
	video.LikeCount = int64Ptr(100)
	video.CommentCount = int64Ptr(50)
	video.PublishedAt = timePtr(time.Now().Add(-24 * time.Hour))
	
	err = s.repo.Update(video)
	if err != nil {
		s.logger.Errorf("Failed to update video metadata: %v", err)
		s.repo.UpdateStatus(videoID, domain.VideoStatusFailed, stringPtr("Failed to update video metadata"))
		return
	}

	// Step 2: Extract transcript
	transcript, err := s.transcriptService.extractAndSaveTranscript(video.YouTubeID)
	if err != nil {
		s.logger.Errorf("Failed to extract transcript: %v", err)
		// Don't fail the entire process if transcript extraction fails
		s.logger.Warn("Continuing without transcript", "video_id", videoID)
	}

	var transcriptID *string
	if transcript != nil {
		transcriptID = stringPtr(transcript.ID)
		s.logger.Info("Successfully extracted transcript", "video_id", videoID, "transcript_id", transcript.ID)
	}

	// Step 3: Update video with processing results
	// For now, we'll just mark as completed
	// In the future, this is where you would generate summaries, etc.
	var summaryID *string // TODO: Implement summary generation

	err = s.repo.UpdateProcessingResults(videoID, transcriptID, summaryID)
	if err != nil {
		s.logger.Errorf("Failed to update processing results: %v", err)
		s.repo.UpdateStatus(videoID, domain.VideoStatusFailed, stringPtr("Failed to update processing results"))
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