// internal/repository/postgres/transcript_repository.go
package postgres

import (
	"encoding/json"

	"github.com/0xsj/alya.io/backend/internal/domain"
	"github.com/0xsj/alya.io/backend/pkg/errors"
	"github.com/0xsj/alya.io/backend/pkg/logger"
	"github.com/jmoiron/sqlx"
)

type TranscriptRepository struct {
	db     *sqlx.DB
	logger logger.Logger
}

func NewTranscriptRepository(db *sqlx.DB, logger logger.Logger) *TranscriptRepository {
	return &TranscriptRepository{
		db:     db,
		logger: logger.WithLayer("repository.transcript"),
	}
}

func (r *TranscriptRepository) Create(transcript *domain.Transcript) error {
	query := `
		INSERT INTO transcripts (
			id, video_id, language, segments, raw_text, source, processed_at, created_at
		) VALUES (
			:id, :video_id, :language, :segments, :raw_text, :source, :processed_at, :created_at
		)
	`

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Convert segments to JSON
	segmentsJSON, err := json.Marshal(transcript.Segments)
	if err != nil {
		return errors.Wrap(err, "failed to marshal segments")
	}

	// Create a map for named parameters
	params := map[string]interface{}{
		"id":           transcript.ID,
		"video_id":     transcript.VideoID,
		"language":     transcript.Language,
		"segments":     string(segmentsJSON),
		"raw_text":     transcript.RawText,
		"source":       transcript.Source,
		"processed_at": transcript.ProcessedAt,
		"created_at":   transcript.CreatedAt,
	}

	_, err = tx.NamedExec(query, params)
	if err != nil {
		return errors.ParsePqError(err)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	r.logger.Info("Created transcript", "transcript_id", transcript.ID, "video_id", transcript.VideoID)
	return nil
}

func (r *TranscriptRepository) GetByID(id string) (*domain.Transcript, error) {
	query := `
		SELECT id, video_id, language, segments, raw_text, source, processed_at, created_at
		FROM transcripts
		WHERE id = $1
	`

	var transcript domain.Transcript
	var segmentsJSON string

	err := r.db.QueryRow(query, id).Scan(
		&transcript.ID,
		&transcript.VideoID,
		&transcript.Language,
		&segmentsJSON,
		&transcript.RawText,
		&transcript.Source,
		&transcript.ProcessedAt,
		&transcript.CreatedAt,
	)

	if err != nil {
		if errors.IsNoRows(err) {
			return nil, errors.NewNotFoundError("transcript not found", errors.ErrNoRows)
		}
		return nil, errors.ParsePqError(err)
	}

	// Parse segments JSON
	if err := json.Unmarshal([]byte(segmentsJSON), &transcript.Segments); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal segments")
	}

	return &transcript, nil
}

func (r *TranscriptRepository) GetByVideoID(videoID string) (*domain.Transcript, error) {
	query := `
		SELECT id, video_id, language, segments, raw_text, source, processed_at, created_at
		FROM transcripts
		WHERE video_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var transcript domain.Transcript
	var segmentsJSON string

	err := r.db.QueryRow(query, videoID).Scan(
		&transcript.ID,
		&transcript.VideoID,
		&transcript.Language,
		&segmentsJSON,
		&transcript.RawText,
		&transcript.Source,
		&transcript.ProcessedAt,
		&transcript.CreatedAt,
	)

	if err != nil {
		if errors.IsNoRows(err) {
			return nil, errors.NewNotFoundError("transcript not found", errors.ErrNoRows)
		}
		return nil, errors.ParsePqError(err)
	}

	// Parse segments JSON
	if err := json.Unmarshal([]byte(segmentsJSON), &transcript.Segments); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal segments")
	}

	return &transcript, nil
}

func (r *TranscriptRepository) Update(transcript *domain.Transcript) error {
	query := `
		UPDATE transcripts
		SET 
			language = :language,
			segments = :segments,
			raw_text = :raw_text,
			source = :source,
			processed_at = :processed_at
		WHERE id = :id
	`

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Convert segments to JSON
	segmentsJSON, err := json.Marshal(transcript.Segments)
	if err != nil {
		return errors.Wrap(err, "failed to marshal segments")
	}

	// Create a map for named parameters
	params := map[string]interface{}{
		"id":           transcript.ID,
		"language":     transcript.Language,
		"segments":     string(segmentsJSON),
		"raw_text":     transcript.RawText,
		"source":       transcript.Source,
		"processed_at": transcript.ProcessedAt,
	}

	result, err := tx.NamedExec(query, params)
	if err != nil {
		return errors.ParsePqError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("transcript not found", errors.ErrNoRows)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	r.logger.Info("Updated transcript", "transcript_id", transcript.ID)
	return nil
}

func (r *TranscriptRepository) Delete(id string) error {
	query := `DELETE FROM transcripts WHERE id = $1`

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
		return errors.NewNotFoundError("transcript not found", errors.ErrNoRows)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	r.logger.Info("Deleted transcript", "transcript_id", id)
	return nil
}

func (r *TranscriptRepository) Search(query string, page, pageSize int) ([]*domain.Transcript, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// Search in raw_text using full-text search
	searchQuery := `
		SELECT id, video_id, language, segments, raw_text, source, processed_at, created_at
		FROM transcripts
		WHERE raw_text ILIKE '%' || $1 || '%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	countQuery := `
		SELECT COUNT(*)
		FROM transcripts
		WHERE raw_text ILIKE '%' || $1 || '%'
	`

	// Get total count
	var total int
	err := r.db.Get(&total, countQuery, query)
	if err != nil {
		return nil, 0, errors.ParsePqError(err)
	}

	// Get transcripts
	rows, err := r.db.Query(searchQuery, query, pageSize, offset)
	if err != nil {
		return nil, 0, errors.ParsePqError(err)
	}
	defer rows.Close()

	transcripts := make([]*domain.Transcript, 0, pageSize)
	for rows.Next() {
		var transcript domain.Transcript
		var segmentsJSON string

		err := rows.Scan(
			&transcript.ID,
			&transcript.VideoID,
			&transcript.Language,
			&segmentsJSON,
			&transcript.RawText,
			&transcript.Source,
			&transcript.ProcessedAt,
			&transcript.CreatedAt,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to scan transcript")
		}

		// Parse segments JSON
		if err := json.Unmarshal([]byte(segmentsJSON), &transcript.Segments); err != nil {
			r.logger.Warn("Failed to unmarshal segments", "transcript_id", transcript.ID, "error", err)
			transcript.Segments = []domain.TranscriptSegment{}
		}

		transcripts = append(transcripts, &transcript)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.Wrap(err, "error iterating over rows")
	}

	return transcripts, total, nil
}