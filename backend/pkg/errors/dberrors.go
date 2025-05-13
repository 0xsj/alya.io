// pkg/errors/dberrors.go
package errors

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
)

const (
	PgUniqueViolation      = "23505" // Unique constraint violation
	PgForeignKeyViolation  = "23503" // Foreign key violation
	PgCheckViolation       = "23514" // Check constraint violation
	PgNotNullViolation     = "23502" // Not null constraint violation
	PgConnectionFailure    = "08006" // Connection failure
	PgInvalidCatalogName   = "3D000" // Invalid database name
	PgInvalidAuthSpec      = "28000" // Invalid authentication
	PgInsufficientPrivilege = "42501" // Insufficient privilege
	PgUndefinedTable       = "42P01" // Undefined table
	PgUndefinedColumn      = "42703" // Undefined column
	PgDuplicateTable       = "42P07" // Duplicate table
	PgDuplicateColumn      = "42701" // Duplicate column
	PgTooManyConnections   = "53300" // Too many connections
	PgSerializationFailure = "40001" // Serialization failure (deadlock)
	PgDiskFull             = "53100" // Disk full
	PgOutOfMemory          = "53200" // Out of memory
)

// Common database error types
var (
	ErrNoRows              = sql.ErrNoRows
	ErrDuplicateKey        = errors.New("duplicate key violation")
	ErrForeignKeyViolation = errors.New("foreign key violation")
	ErrConstraintViolation = errors.New("constraint violation")
	ErrConnectionFailed    = errors.New("database connection failed")
	ErrTransactionFailed   = errors.New("transaction failed")
	ErrQueryFailed         = errors.New("query execution failed")
	ErrStatementPrepareFailed = errors.New("statement preparation failed")
	ErrInvalidParameter    = errors.New("invalid query parameter")
	ErrMigrationFailed     = errors.New("database migration failed")
	ErrTableNotFound       = errors.New("table not found")
	ErrColumnNotFound      = errors.New("column not found")
	ErrInvalidCredentials  = errors.New("invalid database credentials")
	ErrInsufficientPrivileges = errors.New("insufficient database privileges")
	ErrDatabaseCorruption  = errors.New("database corruption detected")
	ErrDeadlockDetected    = errors.New("deadlock detected")
	ErrQuotaExceeded       = errors.New("database quota exceeded")
	ErrSystemResource      = errors.New("database system resource error")
)

// IsNoRows checks if the error is a "no rows in result set" error
func IsNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

// IsDuplicateKey checks if the error is a duplicate key error
func IsDuplicateKey(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == PgUniqueViolation
	}
	return errors.Is(err, ErrDuplicateKey)
}

// IsForeignKeyViolation checks if the error is a foreign key violation
func IsForeignKeyViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == PgForeignKeyViolation
	}
	return errors.Is(err, ErrForeignKeyViolation)
}

// IsConstraintViolation checks if the error is any kind of constraint violation
func IsConstraintViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == PgCheckViolation || 
				pqErr.Code == PgNotNullViolation || 
				pqErr.Code == PgUniqueViolation || 
				pqErr.Code == PgForeignKeyViolation
	}
	return errors.Is(err, ErrConstraintViolation)
}

// IsConnectionFailure checks if the error is a connection failure
func IsConnectionFailure(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == PgConnectionFailure
	}
	return errors.Is(err, ErrConnectionFailed)
}

// IsTableNotFound checks if the error indicates a table does not exist
func IsTableNotFound(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == PgUndefinedTable
	}
	return errors.Is(err, ErrTableNotFound)
}

// IsColumnNotFound checks if the error indicates a column does not exist
func IsColumnNotFound(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == PgUndefinedColumn
	}
	return errors.Is(err, ErrColumnNotFound)
}

// IsDeadlock checks if the error indicates a deadlock was detected
func IsDeadlock(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == PgSerializationFailure
	}
	return errors.Is(err, ErrDeadlockDetected)
}

// IsResourceError checks if the error is related to system resources (memory, disk, etc.)
func IsResourceError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == PgDiskFull || pqErr.Code == PgOutOfMemory || pqErr.Code == PgTooManyConnections
	}
	return errors.Is(err, ErrSystemResource) || errors.Is(err, ErrQuotaExceeded)
}

// IsAuthenticationError checks if the error is an authentication error
func IsAuthenticationError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == PgInvalidAuthSpec || pqErr.Code == PgInsufficientPrivilege
	}
	return errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrInsufficientPrivileges)
}

// ParsePqError parses a PostgreSQL error and returns a standardized error
func ParsePqError(err error) error {
	if err == nil {
		return nil
	}
	
	if pqErr, ok := err.(*pq.Error); ok {
		// Parse the PostgreSQL error code
		switch pqErr.Code {
		case PgUniqueViolation:
			return WrapWith(err, "Duplicate entry", NewConflictError("duplicate key violation", ErrDuplicateKey))
		case PgForeignKeyViolation:
			return WrapWith(err, "Foreign key constraint violation", NewBadRequestError("reference not found", ErrForeignKeyViolation))
		case PgCheckViolation:
			return WrapWith(err, "Check constraint violation", NewBadRequestError("validation failed", ErrConstraintViolation))
		case PgNotNullViolation:
			return WrapWith(err, "Not null constraint violation", NewBadRequestError("required field missing", ErrConstraintViolation))
		case PgConnectionFailure:
			return WrapWith(err, "Database connection failed", NewInternalError("database connection error", ErrConnectionFailed))
		case PgInvalidCatalogName:
			return WrapWith(err, "Invalid database name", NewInternalError("configuration error", ErrConnectionFailed))
		case PgInvalidAuthSpec:
			return WrapWith(err, "Invalid authentication", NewInternalError("database authentication error", ErrInvalidCredentials))
		case PgInsufficientPrivilege:
			return WrapWith(err, "Insufficient privileges", NewInternalError("database permission error", ErrInsufficientPrivileges))
		case PgUndefinedTable:
			return WrapWith(err, "Table not found", NewInternalError("database schema error", ErrTableNotFound))
		case PgUndefinedColumn:
			return WrapWith(err, "Column not found", NewInternalError("database schema error", ErrColumnNotFound))
		case PgSerializationFailure:
			return WrapWith(err, "Transaction conflict (deadlock)", NewInternalError("concurrent transaction error", ErrDeadlockDetected))
		case PgDiskFull, PgOutOfMemory, PgTooManyConnections:
			return WrapWith(err, "Database resource error", NewInternalError("database resource error", ErrSystemResource))
		default:
			// Extract details from the error message for more context
			message := "Database error"
			if pqErr.Detail != "" {
				message = pqErr.Detail
			} else if pqErr.Message != "" {
				message = pqErr.Message
			}
			return WrapWith(err, message, NewDatabaseError(message, ErrDatabase))
		}
	}
	
	// Handle standard SQL errors
	if errors.Is(err, sql.ErrNoRows) {
		return NewNotFoundError("no record found", ErrNoRows)
	}
	
	// Check for common error messages
	errMsg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errMsg, "duplicate"):
		return WrapWith(err, "Duplicate entry", NewConflictError("duplicate entry", ErrDuplicateKey))
	case strings.Contains(errMsg, "foreign key"):
		return WrapWith(err, "Foreign key error", NewBadRequestError("reference not found", ErrForeignKeyViolation))
	case strings.Contains(errMsg, "constraint"):
		return WrapWith(err, "Constraint violation", NewBadRequestError("validation failed", ErrConstraintViolation))
	case strings.Contains(errMsg, "connect"):
		return WrapWith(err, "Connection error", NewInternalError("database connection error", ErrConnectionFailed))
	}
	
	// Default to a generic database error
	return WrapWith(err, "Database operation failed", NewDatabaseError("database error", ErrDatabase))
}

// WithRetry runs a database operation with retries for transient errors
func WithRetry(operation func() error, maxRetries int, retryDelay time.Duration) error {
	var err error
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}
		
		// Only retry for specific types of errors
		if IsConnectionFailure(err) || IsDeadlock(err) || IsResourceError(err) {
			if attempt < maxRetries-1 {
				time.Sleep(retryDelay)
				continue
			}
		}
		
		// Don't retry for other errors
		break
	}
	
	return err
}