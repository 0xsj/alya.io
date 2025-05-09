// internal/config/config.go
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
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
		Provider       string // openai, anthropic, etc.
		APIKey         string
		Model          string
		MaxTokens      int
		Temperature    float64
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
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// Load loads the configuration from file and environment variables
func Load() (*Config, error) {
	var cfg Config
	
	// Set default configuration
	setDefaults(&cfg)
	
	// Try to load config from JSON file
	configPaths := []string{"./config.json", "./config/config.json", "/etc/alya/config.json"}
	for _, path := range configPaths {
		if err := loadConfigFile(path, &cfg); err == nil {
			break
		}
	}
	
	// Override with environment variables
	loadFromEnv(&cfg)
	
	// Validate configuration
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	
	return &cfg, nil
}

// loadConfigFile loads configuration from a JSON file
func loadConfigFile(path string, cfg *Config) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	
	// Read file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}
	
	// Parse JSON
	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}
	
	return nil
}

// setDefaults sets default configuration values
func setDefaults(cfg *Config) {
	// Server defaults
	cfg.Server.Host = "0.0.0.0"
	cfg.Server.Port = 8080
	cfg.Server.Timeout = 30 * time.Second
	cfg.Server.Cors.AllowedOrigins = []string{"*"}
	cfg.Server.Cors.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	cfg.Server.Cors.AllowedHeaders = []string{"Content-Type", "Authorization"}
	cfg.Server.Cors.AllowCredentials = false
	cfg.Server.Cors.MaxAge = 300
	cfg.Server.TLSEnabled = false

	// Database defaults
	cfg.Database.Host = "localhost"
	cfg.Database.Port = 5432
	cfg.Database.User = "postgres"
	cfg.Database.Name = "alya"
	cfg.Database.SSLMode = "disable"
	cfg.Database.MaxConns = 20
	cfg.Database.Timeout = 5 * time.Second

	// YouTube defaults
	cfg.YouTube.MaxRetries = 3
	cfg.YouTube.RequestTimeout = 10 * time.Second

	// AI defaults
	cfg.AI.Provider = "openai"
	cfg.AI.Model = "gpt-4"
	cfg.AI.MaxTokens = 1000
	cfg.AI.Temperature = 0.7
	cfg.AI.RequestTimeout = 60 * time.Second

	// Cache defaults
	cfg.Cache.Type = "memory"
	cfg.Cache.TTL = 24 * time.Hour

	// Logger defaults
	cfg.Logger = logger.Config{
		Level:        logger.InfoLevel,
		EnableJSON:   false,
		EnableTime:   true,
		EnableCaller: true,
		CallerSkip:   1,
		CallerDepth:  10,
	}

	// Environment
	cfg.Env = "development"
}

// loadFromEnv overrides configuration with environment variables
func loadFromEnv(cfg *Config) {
	env := NewEnvProvider("ALYA")
	
	// Server
	if host := env.Get("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	cfg.Server.Port = env.GetIntDefault("SERVER_PORT", cfg.Server.Port)
	if timeout, err := env.GetDuration("SERVER_TIMEOUT"); err == nil {
		cfg.Server.Timeout = timeout
	}
	cfg.Server.TLSEnabled = env.GetBoolDefault("SERVER_TLS_ENABLED", cfg.Server.TLSEnabled)
	if cert := env.Get("SERVER_TLS_CERT"); cert != "" {
		cfg.Server.TLSCert = cert
	}
	if key := env.Get("SERVER_TLS_KEY"); key != "" {
		cfg.Server.TLSKey = key
	}
	if origins := env.GetArray("SERVER_CORS_ALLOWED_ORIGINS"); len(origins) > 0 {
		cfg.Server.Cors.AllowedOrigins = origins
	}
	if methods := env.GetArray("SERVER_CORS_ALLOWED_METHODS"); len(methods) > 0 {
		cfg.Server.Cors.AllowedMethods = methods
	}
	if headers := env.GetArray("SERVER_CORS_ALLOWED_HEADERS"); len(headers) > 0 {
		cfg.Server.Cors.AllowedHeaders = headers
	}
	cfg.Server.Cors.AllowCredentials = env.GetBoolDefault("SERVER_CORS_ALLOW_CREDENTIALS", cfg.Server.Cors.AllowCredentials)
	cfg.Server.Cors.MaxAge = env.GetIntDefault("SERVER_CORS_MAX_AGE", cfg.Server.Cors.MaxAge)
	
	// Database
	if host := env.Get("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	cfg.Database.Port = env.GetIntDefault("DB_PORT", cfg.Database.Port)
	if user := env.Get("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := env.Get("DB_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}
	if name := env.Get("DB_NAME"); name != "" {
		cfg.Database.Name = name
	}
	if sslMode := env.Get("DB_SSLMODE"); sslMode != "" {
		cfg.Database.SSLMode = sslMode
	}
	cfg.Database.MaxConns = env.GetIntDefault("DB_MAX_CONNS", cfg.Database.MaxConns)
	if timeout, err := env.GetDuration("DB_TIMEOUT"); err == nil {
		cfg.Database.Timeout = timeout
	}
	
	// YouTube
	if apiKey := env.Get("YOUTUBE_API_KEY"); apiKey != "" {
		cfg.YouTube.APIKey = apiKey
	}
	cfg.YouTube.MaxRetries = env.GetIntDefault("YOUTUBE_MAX_RETRIES", cfg.YouTube.MaxRetries)
	if timeout, err := env.GetDuration("YOUTUBE_REQUEST_TIMEOUT"); err == nil {
		cfg.YouTube.RequestTimeout = timeout
	}
	
	// AI
	if provider := env.Get("AI_PROVIDER"); provider != "" {
		cfg.AI.Provider = provider
	}
	if apiKey := env.Get("AI_API_KEY"); apiKey != "" {
		cfg.AI.APIKey = apiKey
	}
	if model := env.Get("AI_MODEL"); model != "" {
		cfg.AI.Model = model
	}
	cfg.AI.MaxTokens = env.GetIntDefault("AI_MAX_TOKENS", cfg.AI.MaxTokens)
	if tempStr := env.Get("AI_TEMPERATURE"); tempStr != "" {
		if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
			cfg.AI.Temperature = temp
		}
	}
	if timeout, err := env.GetDuration("AI_REQUEST_TIMEOUT"); err == nil {
		cfg.AI.RequestTimeout = timeout
	}
	
	// Cache
	if cacheType := env.Get("CACHE_TYPE"); cacheType != "" {
		cfg.Cache.Type = cacheType
	}
	if address := env.Get("CACHE_ADDRESS"); address != "" {
		cfg.Cache.Address = address
	}
	if password := env.Get("CACHE_PASSWORD"); password != "" {
		cfg.Cache.Password = password
	}
	if ttl, err := env.GetDuration("CACHE_TTL"); err == nil {
		cfg.Cache.TTL = ttl
	}
	
	// Logger
	if level := env.Get("LOG_LEVEL"); level != "" {
		cfg.Logger.Level = getLogLevel(level)
	}
	cfg.Logger.EnableJSON = env.GetBoolDefault("LOG_JSON", cfg.Logger.EnableJSON)
	cfg.Logger.EnableTime = env.GetBoolDefault("LOG_TIME", cfg.Logger.EnableTime)
	cfg.Logger.EnableCaller = env.GetBoolDefault("LOG_CALLER", cfg.Logger.EnableCaller)
	cfg.Logger.DisableColors = env.GetBoolDefault("LOG_NO_COLORS", cfg.Logger.DisableColors)
	
	// Environment
	if envName := env.Get("ENV"); envName != "" {
		cfg.Env = envName
	}
}

// validate checks if the configuration is valid
func validate(cfg *Config) error {
	// Server validation
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return errors.New("invalid server port")
	}
	
	if cfg.Server.TLSEnabled {
		if cfg.Server.TLSCert == "" {
			return errors.New("TLS certificate file path is required when TLS is enabled")
		}
		if cfg.Server.TLSKey == "" {
			return errors.New("TLS key file path is required when TLS is enabled")
		}
	}

	// Database validation
	if cfg.Database.Host == "" {
		return errors.New("database host is required")
	}
	if cfg.Database.Port <= 0 || cfg.Database.Port > 65535 {
		return errors.New("invalid database port")
	}
	if cfg.Database.User == "" {
		return errors.New("database user is required")
	}
	if cfg.Database.Name == "" {
		return errors.New("database name is required")
	}

	// YouTube validation
	if cfg.YouTube.APIKey == "" {
		return errors.New("YouTube API key is required")
	}

	// AI validation
	if cfg.AI.Provider == "" {
		return errors.New("AI provider is required")
	}
	if cfg.AI.APIKey == "" {
		return errors.New("AI API key is required")
	}
	if cfg.AI.Model == "" {
		return errors.New("AI model is required")
	}

	// Cache validation
	if cfg.Cache.Type != "memory" && cfg.Cache.Type != "redis" {
		return errors.New("invalid cache type, must be 'memory' or 'redis'")
	}
	if cfg.Cache.Type == "redis" && cfg.Cache.Address == "" {
		return errors.New("redis address is required when cache type is 'redis'")
	}

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

// StringMap returns the configuration as a string map for logging
func StringMap(cfg *Config) map[string]string {
	result := make(map[string]string)
	
	// Server settings
	result["server.host"] = cfg.Server.Host
	result["server.port"] = fmt.Sprintf("%d", cfg.Server.Port)
	result["server.timeout"] = cfg.Server.Timeout.String()
	result["server.tls_enabled"] = fmt.Sprintf("%t", cfg.Server.TLSEnabled)
	
	// Database settings (mask password)
	result["database.host"] = cfg.Database.Host
	result["database.port"] = fmt.Sprintf("%d", cfg.Database.Port)
	result["database.user"] = cfg.Database.User
	result["database.name"] = cfg.Database.Name
	result["database.sslmode"] = cfg.Database.SSLMode
	
	// YouTube settings (mask API key)
	result["youtube.max_retries"] = fmt.Sprintf("%d", cfg.YouTube.MaxRetries)
	result["youtube.request_timeout"] = cfg.YouTube.RequestTimeout.String()
	
	// AI settings (mask API key)
	result["ai.provider"] = cfg.AI.Provider
	result["ai.model"] = cfg.AI.Model
	result["ai.max_tokens"] = fmt.Sprintf("%d", cfg.AI.MaxTokens)
	result["ai.temperature"] = fmt.Sprintf("%.2f", cfg.AI.Temperature)
	
	// Cache settings
	result["cache.type"] = cfg.Cache.Type
	if cfg.Cache.Type == "redis" {
		result["cache.address"] = cfg.Cache.Address
	}
	result["cache.ttl"] = cfg.Cache.TTL.String()
	
	// Logger settings
	result["logger.level"] = getLogLevelName(cfg.Logger.Level)
	result["logger.enable_json"] = fmt.Sprintf("%t", cfg.Logger.EnableJSON)
	
	// Environment
	result["env"] = cfg.Env
	
	return result
}

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