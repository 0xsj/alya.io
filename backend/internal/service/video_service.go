// internal/service/video_service.go - Updated with real YouTube data extraction
package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/0xsj/alya.io/backend/internal/domain"
	"github.com/0xsj/alya.io/backend/pkg/errors"
	"github.com/0xsj/alya.io/backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type VideoService struct {
	repo               domain.VideoRepository
	transcriptService  *TranscriptService
	logger             logger.Logger
	httpClient         *http.Client
}

// YouTubeVideoData represents basic video metadata from YouTube page
type YouTubeVideoData struct {
	Title         string
	Description   string
	ChannelName   string
	ChannelID     string
	Duration      int64
	ViewCount     int64
	LikeCount     int64
	PublishedDate time.Time
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
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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
		Tags:       pq.StringArray{}, // Initialize empty array
		CreatedBy:  userID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Save to database
	if err := s.repo.Create(video); err != nil {
		return nil, err
	}

	// Start processing in background
	s.logger.Info("About to start background processing", "video_id", video.ID)
	go s.processVideoAsync(video.ID)
	s.logger.Info("Background processing goroutine started", "video_id", video.ID)

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

	// Step 1: Extract real video metadata from YouTube
	s.logger.Info("Extracting YouTube metadata", "youtube_id", video.YouTubeID)
	youtubeData, err := s.extractYouTubeMetadata(video.YouTubeID)
	if err != nil {
		s.logger.Warn("Failed to extract YouTube metadata, using defaults", "error", err)
		// Use fallback data if YouTube extraction fails
		youtubeData = &YouTubeVideoData{
			Title:       "YouTube Video",
			Description: "Unable to extract description",
			ChannelName: "Unknown Channel",
			Duration:    0,
		}
	}

	// Update video with extracted metadata
	video.Title = youtubeData.Title
	if youtubeData.Description != "" {
		video.Description = stringPtr(youtubeData.Description)
	}
	video.ThumbnailURL = stringPtr("https://img.youtube.com/vi/" + video.YouTubeID + "/maxresdefault.jpg")
	if youtubeData.Duration > 0 {
		video.Duration = int64Ptr(youtubeData.Duration)
	}
	video.Language = stringPtr("en") // Default to English, could be detected
	if youtubeData.ChannelName != "" {
		video.Channel = stringPtr(youtubeData.ChannelName)
	}
	if youtubeData.ChannelID != "" {
		video.ChannelID = stringPtr(youtubeData.ChannelID)
	}
	if youtubeData.ViewCount > 0 {
		video.Views = int64Ptr(youtubeData.ViewCount)
	}
	if youtubeData.LikeCount > 0 {
		video.LikeCount = int64Ptr(youtubeData.LikeCount)
	}
	if !youtubeData.PublishedDate.IsZero() {
		video.PublishedAt = timePtr(youtubeData.PublishedDate)
	}
	
	err = s.repo.Update(video)
	if err != nil {
		s.logger.Errorf("Failed to update video metadata: %v", err)
		s.repo.UpdateStatus(videoID, domain.VideoStatusFailed, stringPtr("Failed to update video metadata"))
		return
	}

	s.logger.Info("Updated video metadata", "title", video.Title, "channel", video.Channel)

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

// extractYouTubeMetadata extracts basic metadata from YouTube page
func (s *VideoService) extractYouTubeMetadata(youtubeID string) (*YouTubeVideoData, error) {
	youtubeURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", youtubeID)
	
	// Create request with browser headers
	req, err := http.NewRequest("GET", youtubeURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("YouTube returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	pageContent := string(body)
	
	// Extract metadata using various methods
	data := &YouTubeVideoData{}
	
	// Extract title from multiple possible locations
	data.Title = s.extractTitle(pageContent)
	
	// Extract description
	data.Description = s.extractDescription(pageContent)
	
	// Extract channel info
	data.ChannelName = s.extractChannelName(pageContent)
	
	// Extract view count
	data.ViewCount = s.extractViewCount(pageContent)
	
	return data, nil
}

func (s *VideoService) extractTitle(pageContent string) string {
	patterns := []string{
		`<meta property="og:title" content="([^"]+)"`,
		`<title>([^<]+)</title>`,
		`"title":"([^"]+)"`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(pageContent)
		if len(matches) > 1 {
			title := strings.TrimSpace(matches[1])
			title = strings.TrimSuffix(title, " - YouTube")
			if title != "" {
				return title
			}
		}
	}
	
	return "YouTube Video"
}

func (s *VideoService) extractDescription(pageContent string) string {
	patterns := []string{
		`<meta property="og:description" content="([^"]+)"`,
		`<meta name="description" content="([^"]+)"`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(pageContent)
		if len(matches) > 1 {
			desc := strings.TrimSpace(matches[1])
			if desc != "" && len(desc) > 10 { // Avoid generic descriptions
				return desc
			}
		}
	}
	
	return ""
}

func (s *VideoService) extractChannelName(pageContent string) string {
	patterns := []string{
		`"channelName":"([^"]+)"`,
		`"author":"([^"]+)"`,
		`<meta property="og:site_name" content="([^"]+)"`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(pageContent)
		if len(matches) > 1 {
			channel := strings.TrimSpace(matches[1])
			if channel != "" && channel != "YouTube" {
				return channel
			}
		}
	}
	
	return ""
}

func (s *VideoService) extractViewCount(pageContent string) int64 {
	patterns := []string{
		`"viewCount":"(\d+)"`,
		`"view_count":"(\d+)"`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(pageContent)
		if len(matches) > 1 {
			if count, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
				return count
			}
		}
	}
	
	return 0
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