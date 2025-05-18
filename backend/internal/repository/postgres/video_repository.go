package postgres

import (
	"github.com/0xsj/alya.io/backend/internal/domain"
	"github.com/0xsj/alya.io/backend/pkg/errors"
	"github.com/0xsj/alya.io/backend/pkg/logger"
	"github.com/jmoiron/sqlx"
)

type VideoRepository struct {
	db     *sqlx.DB
	logger logger.Logger
}

func (r *VideoRepository) Delete(id string) error {
	panic("unimplemented")
}

func (r *VideoRepository) GetByYouTubeID(youtubeID string) (*domain.Video, error) {
	panic("unimplemented")
}

func (r *VideoRepository) List(page int, pageSize int, filters map[string]any) ([]*domain.Video, int, error) {
	panic("unimplemented")
}

func (r *VideoRepository) ListByStatus(status domain.VideoStatus, limit int) ([]*domain.Video, error) {
	panic("unimplemented")
}

func (r *VideoRepository) ListByUserID(userID string, page int, pageSize int) ([]*domain.Video, int, error) {
	panic("unimplemented")
}

func (r *VideoRepository) Search(query string, page int, pageSize int) ([]*domain.Video, int, error) {
	panic("unimplemented")
}

func (r *VideoRepository) Update(video *domain.Video) error {
	panic("unimplemented")
}

func (r *VideoRepository) UpdateProcessingResults(id string, transcriptID string, summaryID string) error {
	panic("unimplemented")
}

func (r *VideoRepository) UpdateStatus(id string, status domain.VideoStatus, errorMessage string) error {
	panic("unimplemented")
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
