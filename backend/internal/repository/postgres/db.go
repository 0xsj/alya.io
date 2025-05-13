// internal/repository/postgres/db.go
package postgres

import (
	"fmt"
	"time"

	"github.com/0xsj/alya.io/backend/internal/config"
	"github.com/0xsj/alya.io/backend/pkg/errors"
	"github.com/0xsj/alya.io/backend/pkg/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// NewDB creates a new database connection
func NewDB(config *config.Config, logger logger.Logger) (*sqlx.DB, error) {
	log := logger.WithLayer("database")
	
	// Build connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.SSLMode,
	)
	
	// Connect to database
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, errors.WrapWith(err, "failed to connect to database", 
			errors.NewDatabaseError("database connection error", errors.ErrDatabase))
	}
	
	// Configure connection pool
	db.SetMaxOpenConns(config.Database.MaxConns)
	db.SetMaxIdleConns(config.Database.MaxConns / 2)
	db.SetConnMaxLifetime(time.Hour)
	
	// Test connection
	if err := db.Ping(); err != nil {
		return nil, errors.WrapWith(err, "failed to ping database",
			errors.NewDatabaseError("database connection error", errors.ErrDatabase))
	}
	
	log.Info("Connected to database successfully")
	
	return db, nil
}