// internal/config/env.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type EnvProvider interface {
	Get(key string) string
	GetDefault(key, defaultValue string) string
	GetBool(key string) (bool, error)
	GetBoolDefault(key string, defaultValue bool) bool
	GetInt(key string) (int, error)
	GetIntDefault(key string, defaultValue int) int
	GetDuration(key string) (time.Duration, error)
	GetDurationDefault(key string, defaultValue time.Duration) time.Duration
	GetArray(key string) []string
}

type OsEnvProvider struct {
	prefix string
}

// Get retrieves the value of an environment variable
func (p *OsEnvProvider) Get(key string) string {
	if p.prefix != "" {
		key = fmt.Sprintf("%s_%s", p.prefix, key)
	}
	return os.Getenv(strings.ToUpper(key))
}

// GetDefault retrieves the value of an environment variable or returns a default value
func (p *OsEnvProvider) GetDefault(key, defaultValue string) string {
	value := p.Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetBool retrieves a boolean value from an environment variable
func (p *OsEnvProvider) GetBool(key string) (bool, error) {
	value := p.Get(key)
	if value == "" {
		return false, fmt.Errorf("environment variable %s not found", key)
	}
	
	// Check for string representations
	value = strings.ToLower(value)
	switch value {
	case "true", "yes", "y", "1", "on":
		return true, nil
	case "false", "no", "n", "0", "off":
		return false, nil
	default:
		return false, fmt.Errorf("cannot parse %s as bool: %s", key, value)
	}
}

// GetBoolDefault retrieves a boolean value from an environment variable with a default
func (p *OsEnvProvider) GetBoolDefault(key string, defaultValue bool) bool {
	value, err := p.GetBool(key)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetInt retrieves an integer value from an environment variable
func (p *OsEnvProvider) GetInt(key string) (int, error) {
	value := p.Get(key)
	if value == "" {
		return 0, fmt.Errorf("environment variable %s not found", key)
	}
	return strconv.Atoi(value)
}

// GetIntDefault retrieves an integer value from an environment variable with a default
func (p *OsEnvProvider) GetIntDefault(key string, defaultValue int) int {
	value, err := p.GetInt(key)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetDuration retrieves a duration value from an environment variable
func (p *OsEnvProvider) GetDuration(key string) (time.Duration, error) {
	value := p.Get(key)
	if value == "" {
		return 0, fmt.Errorf("environment variable %s not found", key)
	}
	return time.ParseDuration(value)
}

// GetDurationDefault retrieves a duration value from an environment variable with a default
func (p *OsEnvProvider) GetDurationDefault(key string, defaultValue time.Duration) time.Duration {
	value, err := p.GetDuration(key)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetArray retrieves an array of values from an environment variable
// Array values are comma-separated
func (p *OsEnvProvider) GetArray(key string) []string {
	value := p.Get(key)
	if value == "" {
		return []string{}
	}
	
	parts := strings.Split(value, ",")
	result := make([]string, len(parts))
	
	for i, part := range parts {
		result[i] = strings.TrimSpace(part)
	}
	
	return result
}

// NewEnvProvider creates a new environment variable provider
func NewEnvProvider(prefix string) EnvProvider {
	return &OsEnvProvider{
		prefix: prefix,
	}
}