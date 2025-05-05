package config

import "time"

type EnvProvider interface {
	Get(key string)	string
	GetDefault(key, defaultValue string) string
	GetBool(key string) (bool, error)
	GetBoolDefault(key string, defaultValue bool) bool
	GetInt(key string) (int error)
	GetIntDefault(key string, defaultValue int) int
	GetDuration(key string) (time.Duration, error)
	GetDurationDefault(key string, defaultValue time.Duration) time.Duration
	GetArray(key string) []string
}

type OsEnvProvider struct {}

func Get(){}

func GetDefault(){}

func GetBool() {}

func GetBoolDefault() {}

func GetInt() {}

func GetIntDefault() {}

func GetDuration() {}

func GetDurationDefault() {}

func GetArray() {}

func NewEnvtProvider() {}