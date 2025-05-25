// internal/service/transcript_service.go
package service

import (
	"github.com/0xsj/alya.io/backend/internal/domain"
	"github.com/0xsj/alya.io/backend/pkg/errors"
	"github.com/0xsj/alya.io/backend/pkg/logger"
)

type TranscriptService struct {
	repo            domain.TranscriptRepository
	youtubeScraper  *YouTubeScraper
	logger          logger.Logger
}

func NewTranscriptService(
	repo domain.TranscriptRepository,
	youtubeScraper *YouTubeScraper,
	logger logger.Logger,
) *TranscriptService {
	return &TranscriptService{
		repo:           repo,
		youtubeScraper: youtubeScraper,
		logger:         logger.WithLayer("service.transcript"),
	}
}

func (s *TranscriptService) GetTranscript(id string, userID string) (*domain.Transcript, error) {
	transcript, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// TODO: Add permission checks if needed
	// For now, transcripts are accessible to all users

	return transcript, nil
}

func (s *TranscriptService) GetTranscriptByVideoID(videoID string, userID string) (*domain.Transcript, error) {
	// First, try to get existing transcript
	transcript, err := s.repo.GetByVideoID(videoID)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	// If transcript exists, return it
	if transcript != nil {
		s.logger.Info("Found existing transcript", "video_id", videoID, "transcript_id", transcript.ID)
		return transcript, nil
	}

	// If no transcript exists, try to extract it from YouTube
	s.logger.Info("No existing transcript found, attempting to extract from YouTube", "video_id", videoID)
	
	transcript, err = s.extractAndSaveTranscript(videoID)
	if err != nil {
		return nil, err
	}

	return transcript, nil
}

func (s *TranscriptService) SearchTranscripts(query string, page, pageSize int, userID string) ([]*domain.Transcript, int, error) {
	// Validate input
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	if query == "" {
		return nil, 0, errors.NewValidationError("search query cannot be empty", nil)
	}

	// TODO: Add user-specific filtering if needed
	// For now, search across all transcripts

	transcripts, total, err := s.repo.Search(query, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	s.logger.Info("Transcript search completed", 
		"query", query, 
		"results", len(transcripts), 
		"total", total,
		"user_id", userID)

	return transcripts, total, nil
}

// ForceExtractTranscript extracts transcript for a video even if one already exists
func (s *TranscriptService) ForceExtractTranscript(videoID string, userID string) (*domain.Transcript, error) {
	s.logger.Info("Force extracting transcript", "video_id", videoID, "user_id", userID)
	
	transcript, err := s.extractAndSaveTranscript(videoID)
	if err != nil {
		return nil, err
	}

	return transcript, nil
}

// RefreshTranscript re-extracts and updates an existing transcript
func (s *TranscriptService) RefreshTranscript(videoID string, userID string) (*domain.Transcript, error) {
	s.logger.Info("Refreshing transcript", "video_id", videoID, "user_id", userID)

	// Extract new transcript
	newTranscript, err := s.youtubeScraper.GetVideoTranscript(videoID)
	if err != nil {
		return nil, err
	}

	// Check if an existing transcript exists
	existingTranscript, err := s.repo.GetByVideoID(videoID)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	if existingTranscript != nil {
		// Update existing transcript
		existingTranscript.Language = newTranscript.Language
		existingTranscript.Segments = newTranscript.Segments
		existingTranscript.RawText = newTranscript.RawText
		existingTranscript.Source = newTranscript.Source
		existingTranscript.ProcessedAt = newTranscript.ProcessedAt

		err = s.repo.Update(existingTranscript)
		if err != nil {
			return nil, err
		}

		s.logger.Info("Updated existing transcript", "transcript_id", existingTranscript.ID)
		return existingTranscript, nil
	} else {
		// Create new transcript
		err = s.repo.Create(newTranscript)
		if err != nil {
			return nil, err
		}

		s.logger.Info("Created new transcript", "transcript_id", newTranscript.ID)
		return newTranscript, nil
	}
}

// extractAndSaveTranscript is a helper method to extract and save a transcript
func (s *TranscriptService) extractAndSaveTranscript(videoID string) (*domain.Transcript, error) {
	// Extract transcript using YouTube scraper
	transcript, err := s.youtubeScraper.GetVideoTranscript(videoID)
	if err != nil {
		s.logger.Error("Failed to extract transcript from YouTube", "video_id", videoID, "error", err)
		return nil, err
	}

	// Save transcript to database
	err = s.repo.Create(transcript)
	if err != nil {
		s.logger.Error("Failed to save transcript to database", "video_id", videoID, "error", err)
		return nil, err
	}

	s.logger.Info("Successfully extracted and saved transcript", 
		"video_id", videoID, 
		"transcript_id", transcript.ID,
		"language", transcript.Language,
		"segments", len(transcript.Segments))

	return transcript, nil
}

// GetTranscriptText returns just the raw text of a transcript
func (s *TranscriptService) GetTranscriptText(videoID string, userID string) (string, error) {
	transcript, err := s.GetTranscriptByVideoID(videoID, userID)
	if err != nil {
		return "", err
	}

	return transcript.RawText, nil
}

// GetTranscriptSegments returns transcript segments with optional time filtering
func (s *TranscriptService) GetTranscriptSegments(videoID string, userID string, startTime, endTime *float64) ([]domain.TranscriptSegment, error) {
	transcript, err := s.GetTranscriptByVideoID(videoID, userID)
	if err != nil {
		return nil, err
	}

	segments := transcript.Segments

	// Filter by time range if provided
	if startTime != nil || endTime != nil {
		filteredSegments := make([]domain.TranscriptSegment, 0)
		
		for _, segment := range segments {
			// Check if segment overlaps with the requested time range
			if startTime != nil && segment.End < *startTime {
				continue
			}
			if endTime != nil && segment.Start > *endTime {
				continue
			}
			filteredSegments = append(filteredSegments, segment)
		}
		
		segments = filteredSegments
	}

	return segments, nil
}

// ValidateTranscriptExists checks if a transcript exists for a video
func (s *TranscriptService) ValidateTranscriptExists(videoID string) (bool, error) {
	_, err := s.repo.GetByVideoID(videoID)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}