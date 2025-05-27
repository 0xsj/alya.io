// internal/service/youtube_scraper.go - Updated with current working method
package service

import (
	"encoding/json"
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
)

type YouTubeScraper struct {
	client *http.Client
	logger logger.Logger
}

type PlayerResponse struct {
	Captions struct {
		PlayerCaptionsTracklistRenderer struct {
			CaptionTracks []CaptionTrack `json:"captionTracks"`
		} `json:"playerCaptionsTracklistRenderer"`
	} `json:"captions"`
	VideoDetails struct {
		VideoID       string `json:"videoId"`
		Title         string `json:"title"`
		LengthSeconds string `json:"lengthSeconds"`
		Author        string `json:"author"`
		ChannelID     string `json:"channelId"`
	} `json:"videoDetails"`
}

type CaptionTrack struct {
	BaseURL          string `json:"baseUrl"`
	Name             Name   `json:"name"`
	VSSId            string `json:"vssId"`
	LanguageCode     string `json:"languageCode"`
	IsTranslatable   bool   `json:"isTranslatable"`
	Kind             string `json:"kind,omitempty"`
}

type Name struct {
	SimpleText string `json:"simpleText"`
}

func NewYouTubeScraper(logger logger.Logger) *YouTubeScraper {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: false,
		},
	}

	return &YouTubeScraper{
		client: client,
		logger: logger.WithLayer("service.youtube_scraper"),
	}
}

func (ys *YouTubeScraper) GetVideoTranscript(videoID string) (*domain.Transcript, error) {
	ys.logger.Info("Starting transcript extraction", "video_id", videoID)

	// Get player response from YouTube page
	playerResponse, err := ys.getPlayerResponse(videoID)
	if err != nil {
		return nil, errors.WrapWith(err, "failed to get player response", 
			errors.NewExternalServiceError("youtube scraping failed", err))
	}

	// Extract caption tracks
	captionTracks := playerResponse.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks
	if len(captionTracks) == 0 {
		return nil, errors.NewNotFoundError("no captions available for this video", nil)
	}

	// Find the best caption track (prefer English, then auto-generated)
	selectedTrack := ys.selectBestCaptionTrack(captionTracks)
	ys.logger.Info("Selected caption track", 
		"language", selectedTrack.LanguageCode, 
		"name", selectedTrack.Name.SimpleText)

	// Download and parse caption content
	segments, err := ys.downloadAndParseCaptions(selectedTrack.BaseURL)
	if err != nil {
		return nil, errors.WrapWith(err, "failed to download captions", 
			errors.NewExternalServiceError("caption download failed", err))
	}

	// Create transcript object
	transcript := &domain.Transcript{
		ID:          uuid.New().String(),
		VideoID:     videoID,
		Language:    selectedTrack.LanguageCode,
		Segments:    segments,
		RawText:     ys.extractRawText(segments),
		Source:      "youtube_scraper",
		ProcessedAt: time.Now(),
		CreatedAt:   time.Now(),
	}

	ys.logger.Info("Successfully extracted transcript", 
		"video_id", videoID, 
		"segments", len(segments),
		"language", selectedTrack.LanguageCode)

	return transcript, nil
}

func (ys *YouTubeScraper) getPlayerResponse(videoID string) (*PlayerResponse, error) {
	// Construct YouTube URL
	youtubeURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	
	// Create request with proper headers to avoid detection
	req, err := http.NewRequest("GET", youtubeURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	// Add headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Make the request
	resp, err := ys.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch YouTube page")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewExternalServiceError(
			fmt.Sprintf("YouTube returned status %d", resp.StatusCode), nil)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	// Extract player response from page content
	return ys.extractPlayerResponse(string(body))
}

func (ys *YouTubeScraper) extractPlayerResponse(pageContent string) (*PlayerResponse, error) {
	// Updated patterns based on Stack Overflow post
	patterns := []string{
		// Try the new method first (Jan 2025 update)
		`ytplayer\.config\.args\.raw_player_response\s*=\s*(\{.*?\});`,
		// Fallback patterns
		`var ytInitialPlayerResponse = (\{.*?\});`,
		`window\["ytInitialPlayerResponse"\] = (\{.*?\});`,
		`ytInitialPlayerResponse":\s*(\{.*?\})(?:,"|\}$)`,
		`"playerResponse":"(\{.*?\})"`,
		// Additional patterns for embedded players
		`ytplayer\.config\s*=\s*(\{.*?\});`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(pageContent)
		
		if len(matches) > 1 {
			playerResponseJSON := matches[1]
			
			// Handle escaped JSON if needed
			if strings.Contains(playerResponseJSON, `\"`) {
				playerResponseJSON = strings.ReplaceAll(playerResponseJSON, `\"`, `"`)
				playerResponseJSON = strings.ReplaceAll(playerResponseJSON, `\\`, `\`)
			}

			var playerResponse PlayerResponse
			if err := json.Unmarshal([]byte(playerResponseJSON), &playerResponse); err != nil {
				ys.logger.Debug("Failed to parse player response with pattern", "pattern", pattern, "error", err)
				continue
			}

			return &playerResponse, nil
		}
	}

	return nil, errors.NewNotFoundError("could not find player response in page content", nil)
}

func (ys *YouTubeScraper) selectBestCaptionTrack(tracks []CaptionTrack) CaptionTrack {
	// Priority order:
	// 1. English manual captions
	// 2. Any manual captions
	// 3. English auto-generated
	// 4. Any auto-generated captions

	var englishManual, anyManual, englishAuto, anyAuto *CaptionTrack

	for i, track := range tracks {
		isEnglish := strings.HasPrefix(track.LanguageCode, "en")
		isAutoGenerated := track.Kind == "asr"

		if isEnglish && !isAutoGenerated {
			englishManual = &tracks[i]
		} else if !isAutoGenerated {
			if anyManual == nil {
				anyManual = &tracks[i]
			}
		} else if isEnglish && isAutoGenerated {
			if englishAuto == nil {
				englishAuto = &tracks[i]
			}
		} else if isAutoGenerated {
			if anyAuto == nil {
				anyAuto = &tracks[i]
			}
		}
	}

	// Return in priority order
	if englishManual != nil {
		return *englishManual
	}
	if anyManual != nil {
		return *anyManual
	}
	if englishAuto != nil {
		return *englishAuto
	}
	if anyAuto != nil {
		return *anyAuto
	}

	// Fallback to first track
	return tracks[0]
}

func (ys *YouTubeScraper) downloadAndParseCaptions(captionURL string) ([]domain.TranscriptSegment, error) {
	// Parse the URL and add format parameter for plain text
	parsedURL, err := url.Parse(captionURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse caption URL")
	}

	// Add format parameter to get XML captions
	query := parsedURL.Query()
	query.Set("fmt", "srv3") // srv3 format includes timing information
	parsedURL.RawQuery = query.Encode()

	// Download captions
	resp, err := ys.client.Get(parsedURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to download captions")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewExternalServiceError(
			fmt.Sprintf("caption download returned status %d", resp.StatusCode), nil)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read caption content")
	}

	// Parse XML captions
	return ys.parseXMLCaptions(string(content))
}

func (ys *YouTubeScraper) parseXMLCaptions(xmlContent string) ([]domain.TranscriptSegment, error) {
	// Parse XML using regex (simple approach for YouTube's specific format)
	re := regexp.MustCompile(`<text start="([^"]*)" dur="([^"]*)"[^>]*>([^<]*)</text>`)
	matches := re.FindAllStringSubmatch(xmlContent, -1)

	segments := make([]domain.TranscriptSegment, 0, len(matches))

	for i, match := range matches {
		if len(match) != 4 {
			continue
		}

		start, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			ys.logger.Warn("Failed to parse start time", "value", match[1], "error", err)
			continue
		}

		duration, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			ys.logger.Warn("Failed to parse duration", "value", match[2], "error", err)
			continue
		}

		text := ys.cleanCaptionText(match[3])
		if text == "" {
			continue
		}

		segment := domain.TranscriptSegment{
			Index:      i,
			Start:      start,
			End:        start + duration,
			Text:       text,
			Speaker:    "", // YouTube doesn't provide speaker info
			Confidence: 1.0, // Default confidence
		}

		segments = append(segments, segment)
	}

	if len(segments) == 0 {
		return nil, errors.NewNotFoundError("no valid caption segments found", nil)
	}

	return segments, nil
}

func (ys *YouTubeScraper) cleanCaptionText(text string) string {
	// Decode HTML entities
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "\n", " ")
	
	// Remove extra whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	
	return text
}

func (ys *YouTubeScraper) extractRawText(segments []domain.TranscriptSegment) string {
	var builder strings.Builder
	
	for i, segment := range segments {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(segment.Text)
	}
	
	return builder.String()
}