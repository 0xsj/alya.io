package config

import (
	"strings"
	"time"

	"github.com/0xsj/alya.io/backend/pkg/logger"
)

type Config struct {
	Server struct {
		Port       int
		Host       string
		Timeout    time.Duration
		Cors       CorsConfig
		TLSEnabled bool
		TLSCert    string
		TLSKey     string
	}

	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
		SSLMode  string
		MaxConns int
		Timeout  time.Duration
	}

	YouTube struct {
		APIKey         string
		MaxRetries     int
		RequestTimeout time.Duration
	}

	AI struct {
		Provider      string  // openai, anthropic, etc.
		APIKey        string
		Model         string
		MaxTokens     int
		Temperature   float64
		RequestTimeout time.Duration
	}

	Cache struct {
		Type     string // memory, redis
		Address  string
		Password string
		TTL      time.Duration
	}

	Logger logger.Config

	Env string
}

type CorsConfig struct {
	AllowedOrigins  []string
	AllowedMethods  []string
	AllowedHeaders  []string
	AllowCredentials bool
	MaxAge          int
}

func Load() {}

func validate() error {

	return nil
}

func getLogLevel(level string) int {
	switch strings.ToLower(level) {
	case "debug":
		return logger.DebugLevel
	case "info":
		return logger.InfoLevel
	case "warn", "warning":
		return logger.WarnLevel
	case "error":
		return logger.ErrorLevel
	case "fatal":
		return logger.FatalLevel
	case "panic":
		return logger.PanicLevel
	default:
		return logger.InfoLevel
	}
}

func StringMap() {}

func getLogLevelName(level int) string {
	switch level {
	case logger.DebugLevel:
		return "debug"
	case logger.InfoLevel:
		return "info"
	case logger.WarnLevel:
		return "warn"
	case logger.ErrorLevel:
		return "error"
	case logger.FatalLevel:
		return "fatal"
	case logger.PanicLevel:
		return "panic"
	default:
		return "unknown"
	}
}