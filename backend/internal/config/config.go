package config

import (
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