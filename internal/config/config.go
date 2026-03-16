// Package config provides configuration management for the application.
package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration.
type Config struct {
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Encryption EncryptionConfig
	Log        LogConfig
	Cache      CacheConfig // Nuevo
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	URL string
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Addr     string
	Password string
}

// JWTConfig holds JWT authentication configuration.
type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

// EncryptionConfig holds encryption configuration.
type EncryptionConfig struct {
	Key string
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string
}

// CacheConfig configura TTLs de cache en segundos.
type CacheConfig struct {
	ContactsTTL int // TTL para contacts (default 300 = 5 min)
	ChatsTTL    int // TTL para lista de chats (default 120 = 2 min)
	ChatInfoTTL int // TTL para info de chat individual (default 300 = 5 min)
	ResolveTTL  int // TTL para resolve peer (default 600 = 10 min)
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	expiry, _ := strconv.Atoi(os.Getenv("JWT_EXPIRY_HOURS"))
	if expiry == 0 {
		expiry = 24
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		Database: DatabaseConfig{
			URL: os.Getenv("DB_URL"),
		},
		Redis: RedisConfig{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: os.Getenv("REDIS_PASSWORD"),
		},
		JWT: JWTConfig{
			Secret:      os.Getenv("JWT_SECRET"),
			ExpiryHours: expiry,
		},
		Encryption: EncryptionConfig{
			Key: os.Getenv("ENCRYPTION_KEY"),
		},
		Log: LogConfig{
			Level: logLevel,
		},
		Cache: loadCacheConfig(),
	}, nil
}

func loadCacheConfig() CacheConfig {
	return CacheConfig{
		ContactsTTL: getEnvInt("CACHE_CONTACTS_TTL", 300),  // 5 min default
		ChatsTTL:    getEnvInt("CACHE_CHATS_TTL", 120),     // 2 min default
		ChatInfoTTL: getEnvInt("CACHE_CHAT_INFO_TTL", 300), // 5 min default
		ResolveTTL:  getEnvInt("CACHE_RESOLVE_TTL", 600),   // 10 min default
	}
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}
