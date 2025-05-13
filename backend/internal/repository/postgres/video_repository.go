package postgres

import (
	"github.com/jmoiron/sqlx"
)

type VideoRepository struct {
	db *sqlx.DB
}