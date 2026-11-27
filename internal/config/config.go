package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env  string
	HTTP HTTPConfig
	DB   DBConfig
	Log  LogConfig
}

type HTTPConfig struct {
	Addr           string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	RequestTimeout time.Duration
}

type DBConfig struct {
	PrimaryDSN      string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env: getEnv("APP_ENV", "development"),
		HTTP: HTTPConfig{
			Addr:           getEnv("HTTP_ADDR", ":8080"),
			ReadTimeout:    getDuration("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:   getDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:    getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			RequestTimeout: getDuration("HTTP_REQUEST_TIMEOUT", 30*time.Second),
		},
		DB: DBConfig{
			PrimaryDSN:      getEnv("DB_PRIMARY_DSN", ""),
			MaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnMaxIdleTime: getDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	if cfg.DB.PrimaryDSN == "" {
		return nil, fmt.Errorf("config: DB_PRIMARY_DSN is required")
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}