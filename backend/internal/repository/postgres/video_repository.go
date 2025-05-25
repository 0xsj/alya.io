package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/0xsj/alya.io/backend/internal/domain"
	"github.com/0xsj/alya.io/backend/pkg/errors"
	"github.com/0xsj/alya.io/backend/pkg/logger"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type VideoRepository struct {
	db     *sqlx.DB
	logger logger.Logger
}

func NewVideoRepository(db *sqlx.DB, logger logger.Logger) *VideoRepository {
	return &VideoRepository{
		db:     db,
		logger: logger.WithLayer("repository.video"),
	}
}


func (r *VideoRepository) Create(video *domain.Video) error {
	query := `
		INSERT INTO videos (
			id, youtube_id, title, description, url, thumbnail_url,
			status, visibility, duration, language, transcript_id, summary_id,
			tags, channel, channel_id, views, like_count, comment_count,
			published_at, processed_at, error_message, created_by, created_at, updated_at
		) VALUES (
			:id, :youtube_id, :title, :description, :url, :thumbnail_url,
			:status, :visibility, :duration, :language, :transcript_id, :summary_id,
			:tags, :channel, :channel_id, :views, :like_count, :comment_count,
			:published_at, :processed_at, :error_message, :created_by, :created_at, :updated_at
		)
	`

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if video.Tags == nil {
		video.Tags = []string{}
	}

	_, err = tx.NamedExec(query, video)
	if err != nil {
		return errors.ParsePqError(err)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (r *VideoRepository) GetByID(id string) (*domain.Video, error) {
	query := `
		SELECT * FROM videos
		WHERE id = $1
	`

	var video domain.Video
	err := r.db.Get(&video, query, id)
	if err != nil {
		if errors.IsNoRows(err) {
			return nil, errors.NewNotFoundError("video not found", errors.ErrNoRows)
		}
		return nil, errors.ParsePqError(err)
	}

	return &video, nil
}

func (r *VideoRepository) GetByYouTubeID(youtubeID string) (*domain.Video, error) {
	query := `
		SELECT * FROM videos
		WHERE youtube_id = $1
	`

	var video domain.Video
	err := r.db.Get(&video, query, youtubeID)
	if err != nil {
		if errors.IsNoRows(err) {
			return nil, errors.NewNotFoundError("video not found", errors.ErrNoRows)
		}
		return nil, errors.ParsePqError(err)
	}
	return &video, nil
}

func (r *VideoRepository) Update(video *domain.Video) error {
	query := `
		UPDATE videos
		SET 
			title = :title,
			description = :description,
			url = :url,
			thumbnail_url = :thumbnail_url,
			status = :status,
			visibility = :visibility,
			duration = :duration,
			language = :language,
			transcript_id = :transcript_id,
			summary_id = :summary_id,
			tags = :tags,
			channel = :channel,
			channel_id = :channel_id,
			views = :views,
			like_count = :like_count,
			comment_count = :comment_count,
			published_at = :published_at,
			processed_at = :processed_at,
			error_message = :error_message,
			updated_at = :updated_at
		WHERE id = :id
	`
	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if video.Tags == nil {
		video.Tags = []string{}
	}

	video.UpdatedAt = time.Now()
	result, err := tx.NamedExec(query, video)
	if err != nil {
		return errors.ParsePqError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("video not found", errors.ErrNoRows)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (r *VideoRepository) UpdateStatus(id string, status domain.VideoStatus, errorMessage *string) error {
	query := `
		UPDATE videos
		SET 
			status = $1,
			error_message = $2,
			updated_at = $3
		WHERE id = $4
	`

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	now := time.Now()
	result, err := tx.Exec(query, status, errorMessage, now, id)
	if err != nil {
		return errors.ParsePqError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}
	if rowsAffected == 0 {
		return errors.NewNotFoundError("video not found", errors.ErrNoRows)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}
	return nil
}


func (r *VideoRepository) UpdateProcessingResults(id string, transcriptID *string, summaryID *string) error {
	query := `
		UPDATE videos
		SET 
			transcript_id = $1,
			summary_id = $2,
			status = $3,
			processed_at = $4,
			updated_at = $5
		WHERE id = $6
	`

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()
	now := time.Now()
	result, err := tx.Exec(query, transcriptID, summaryID, domain.VideoStatusCompleted, now, now, id)
	if err != nil {
		return errors.ParsePqError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}
	if rowsAffected == 0 {
		return errors.NewNotFoundError("video not found", errors.ErrNoRows)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (r *VideoRepository) Delete(id string) error {
	query := `
		DELETE FROM videos
		WHERE id = $1
	`

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	result, err := tx.Exec(query, id)
	if err != nil {
		return errors.ParsePqError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("video not found", errors.ErrNoRows)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (r *VideoRepository) List(page, pageSize int, filters map[string]any) ([]*domain.Video, int, error) {
	baseQuery := "SELECT * FROM videos"
	countQuery := "SELECT COUNT(*) FROM videos"
	
	// Process filters
	var conditions []string
	var args []any
	var namedArgs map[string]any
	
	argIndex := 1
	
	if len(filters) > 0 {
		namedArgs = make(map[string]any)
		conditions = make([]string, 0, len(filters))

		// unused
		fmt.Println(namedArgs)
		
		for key, value := range filters {
			switch key {
			case "status":
				conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
				args = append(args, value)
				argIndex++
			case "visibility":
				conditions = append(conditions, fmt.Sprintf("visibility = $%d", argIndex))
				args = append(args, value)
				argIndex++
			case "created_by":
				conditions = append(conditions, fmt.Sprintf("created_by = $%d", argIndex))
				args = append(args, value)
				argIndex++
			case "tags":
				if tags, ok := value.([]string); ok && len(tags) > 0 {
					conditions = append(conditions, fmt.Sprintf("tags && $%d", argIndex))
					args = append(args, pq.Array(tags))
					argIndex++
				}
			case "search":
				if search, ok := value.(string); ok && search != "" {
					conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex+1))
					searchTerm := "%" + search + "%"
					args = append(args, searchTerm, searchTerm)
					argIndex += 2
				}
			}
		}
	}
	
	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	
	offset := (page - 1) * pageSize
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)
	
	var total int
	err := r.db.Get(&total, countQuery, args[:len(args)-2]...)
	if err != nil {
		return nil, 0, errors.ParsePqError(err)
	}
	
	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, 0, errors.ParsePqError(err)
	}
	defer rows.Close()
	
	videos := make([]*domain.Video, 0, pageSize)
	for rows.Next() {
		var video domain.Video
		if err := rows.Scan(
			&video.ID, &video.YouTubeID, &video.Title, &video.Description, &video.URL,
			&video.ThumbnailURL, &video.Status, &video.Visibility, &video.Duration,
			&video.Language, &video.TranscriptID, &video.SummaryID, pq.Array(&video.Tags),
			&video.Channel, &video.ChannelID, &video.Views, &video.LikeCount,
			&video.CommentCount, &video.PublishedAt, &video.ProcessedAt, &video.ErrorMessage,
			&video.CreatedBy, &video.CreatedAt, &video.UpdatedAt,
		); err != nil {
			return nil, 0, errors.Wrap(err, "failed to scan video")
		}
		videos = append(videos, &video)
	}
	
	if err := rows.Err(); err != nil {
		return nil, 0, errors.Wrap(err, "error iterating over rows")
	}
	
	return videos, total, nil
}

func (r *VideoRepository) ListByUserID(userID string, page, pageSize int) ([]*domain.Video, int, error) {
	filters := map[string]any{
		"created_by": userID,
	}
	return r.List(page, pageSize, filters)
}

func (r *VideoRepository) ListByStatus(status domain.VideoStatus, limit int) ([]*domain.Video, error) {
	query := `
		SELECT * FROM videos
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`
	
	if limit <= 0 {
		limit = 10
	}
	
	rows, err := r.db.Query(query, status, limit)
	if err != nil {
		return nil, errors.ParsePqError(err)
	}
	defer rows.Close()
	
	videos := make([]*domain.Video, 0, limit)
	for rows.Next() {
		var video domain.Video
		if err := rows.Scan(
			&video.ID, &video.YouTubeID, &video.Title, &video.Description, &video.URL,
			&video.ThumbnailURL, &video.Status, &video.Visibility, &video.Duration,
			&video.Language, &video.TranscriptID, &video.SummaryID, pq.Array(&video.Tags),
			&video.Channel, &video.ChannelID, &video.Views, &video.LikeCount,
			&video.CommentCount, &video.PublishedAt, &video.ProcessedAt, &video.ErrorMessage,
			&video.CreatedBy, &video.CreatedAt, &video.UpdatedAt,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan video")
		}
		videos = append(videos, &video)
	}
	
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating over rows")
	}
	
	return videos, nil
}

func (r *VideoRepository) Search(query string, page, pageSize int) ([]*domain.Video, int, error) {
	filters := map[string]any{
		"search": query,
	}
	return r.List(page, pageSize, filters)
}