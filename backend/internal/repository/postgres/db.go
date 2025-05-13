// internal/repository/postgres/db.go
package postgres

import (
	"fmt"
	"time"

	"github.com/0xsj/alya.io/backend/internal/config"
	"github.com/0xsj/alya.io/backend/pkg/errors"
	"github.com/0xsj/alya.io/backend/pkg/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewDB(config *config.Config, logger logger.Logger) (*sqlx.DB, error) {
	log := logger.WithLayer("database")
	
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.SSLMode,
	)
	
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, errors.WrapWith(err, "failed to connect to database", 
			errors.NewDatabaseError("database connection error", errors.ErrDatabase))
	}
	
	db.SetMaxOpenConns(config.Database.MaxConns)
	db.SetMaxIdleConns(config.Database.MaxConns / 2)
	db.SetConnMaxLifetime(time.Hour)
	
	if err := db.Ping(); err != nil {
		return nil, errors.WrapWith(err, "failed to ping database",
			errors.NewDatabaseError("database connection error", errors.ErrDatabase))
	}
	
	log.Info("Connected to database successfully")
	
	return db, nil
}