// internal/domain/analytics.go
package domain

import (
	"time"
)

// AnalyticsMetric represents a metric type for analytics
type AnalyticsMetric string

const (
	MetricVideoViews      AnalyticsMetric = "video_views"
	MetricSummaryViews    AnalyticsMetric = "summary_views"
	MetricTranscriptViews AnalyticsMetric = "transcript_views"
	MetricSearches        AnalyticsMetric = "searches"
	MetricSignups         AnalyticsMetric = "signups"
	MetricAPIRequests     AnalyticsMetric = "api_requests"
	MetricProcessingTime  AnalyticsMetric = "processing_time"
)

// AnalyticsPeriod represents the time period for analytics data
type AnalyticsPeriod string

const (
	PeriodHourly  AnalyticsPeriod = "hourly"
	PeriodDaily   AnalyticsPeriod = "daily"
	PeriodWeekly  AnalyticsPeriod = "weekly"
	PeriodMonthly AnalyticsPeriod = "monthly"
)

// AnalyticsDataPoint represents a single data point in analytics
type AnalyticsDataPoint struct {
	ID        string          `json:"id" validate:"required"`
	Metric    AnalyticsMetric `json:"metric" validate:"required"`
	Value     float64         `json:"value"`
	Timestamp time.Time       `json:"timestamp"`
	UserID    string          `json:"user_id,omitempty"`     // Optional user association
	ResourceID string         `json:"resource_id,omitempty"` // Optional resource association
	Dimensions map[string]string `json:"dimensions,omitempty"` // Additional dimensions (browser, OS, etc.)
}

// AnalyticsAggregate represents aggregated analytics data
type AnalyticsAggregate struct {
	Metric      AnalyticsMetric         `json:"metric"`
	Period      AnalyticsPeriod         `json:"period"`
	StartTime   time.Time               `json:"start_time"`
	EndTime     time.Time               `json:"end_time"`
	TotalValue  float64                 `json:"total_value"`
	AverageValue float64                `json:"average_value"`
	MinValue    float64                 `json:"min_value"`
	MaxValue    float64                 `json:"max_value"`
	DataPoints  []AnalyticsDataPoint    `json:"data_points,omitempty"`
	Dimensions  map[string]interface{}  `json:"dimensions,omitempty"` // Dimension breakdowns
}

// UserStats represents analytics stats for a specific user
type UserStats struct {
	UserID              string    `json:"user_id"`
	TotalVideos         int       `json:"total_videos"`
	TotalNotes          int       `json:"total_notes"`
	TotalBookmarks      int       `json:"total_bookmarks"`
	TotalPlaylists      int       `json:"total_playlists"`
	TotalWatchTime      int64     `json:"total_watch_time"` // In seconds
	LastActive          time.Time `json:"last_active"`
	SignupDate          time.Time `json:"signup_date"`
	ProcessedVideoCount int       `json:"processed_video_count"`
	QuotaUsed           int       `json:"quota_used"`
	QuotaLimit          int       `json:"quota_limit"`
}

// AnalyticsRepository defines the interface for analytics storage operations
type AnalyticsRepository interface {
	TrackEvent(metric AnalyticsMetric, value float64, userID, resourceID string, dimensions map[string]string) error
	GetDataPoints(metric AnalyticsMetric, startTime, endTime time.Time, dimensions map[string]string) ([]AnalyticsDataPoint, error)
	GetAggregate(metric AnalyticsMetric, period AnalyticsPeriod, startTime, endTime time.Time, dimensions map[string]string) (*AnalyticsAggregate, error)
	GetTopResources(metric AnalyticsMetric, resourceType string, limit int, timeRange time.Duration) ([]map[string]interface{}, error)
	GetUserStats(userID string) (*UserStats, error)
	GetSystemStats(period AnalyticsPeriod) (map[string]any, error)
}

// AnalyticsService defines high-level operations for analytics
type AnalyticsService interface {
	TrackUserEvent(userID string, metric AnalyticsMetric, value float64, resourceID string, dimensions map[string]string) error
	TrackAnonymousEvent(metric AnalyticsMetric, value float64, dimensions map[string]string) error
	GetMetricHistory(metric AnalyticsMetric, period AnalyticsPeriod, days int, dimensions map[string]string) (*AnalyticsAggregate, error)
	GetUserDashboard(userID string) (map[string]any, error)
	GetAdminDashboard() (map[string]any, error)
	GetPopularVideos(timeRange time.Duration, limit int) ([]*Video, error)
	GetPopularSearchTerms(timeRange time.Duration, limit int) ([]string, error)
}
