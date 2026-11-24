package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env   string
	HTTP  HTTPConfig
	DB    DBConfig
	Log   LogConfig
	Auth  AuthConfig
	Email EmailConfig
	Payments PaymentConfig
}

type HTTPConfig struct {
	Addr           string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	RequestTimeout time.Duration
}

type DBConfig struct {
	PrimaryDSN string
	// ReplicaDSN is optional. When empty, reads fall back to Primary and the
	// DB facade aliases the replica handle to the primary pool.
	ReplicaDSN      string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

type AuthConfig struct {
	JWTSecret       string
	JWTIssuer       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	ResetTokenTTL   time.Duration
	ResetMaxAttempts  int

	LoginRatePerWindow    int
	LoginWindow           time.Duration
	RegisterRatePerWindow int
	RegisterWindow        time.Duration
	RefreshRatePerWindow  int
	RefreshWindow         time.Duration
}

type EmailConfig struct {
	// When SMTPHost is empty, the app uses the Noop sender (dev only).
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
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
			ReplicaDSN:      getEnv("DB_REPLICA_DSN", ""),
			MaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnMaxIdleTime: getDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Auth: AuthConfig{
			JWTSecret:       getEnv("JWT_SECRET", ""),
			JWTIssuer:       getEnv("JWT_ISSUER", "webhook"),
			AccessTokenTTL:  getDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTokenTTL: getDuration("JWT_REFRESH_TTL", 30*24*time.Hour),
			ResetTokenTTL:    getDuration("PASSWORD_RESET_TTL", 10*time.Minute),
			ResetMaxAttempts: getInt("PASSWORD_RESET_MAX_ATTEMPTS", 5),

			LoginRatePerWindow:    getInt("RL_LOGIN_MAX", 10),
			LoginWindow:           getDuration("RL_LOGIN_WINDOW", time.Minute),
			RegisterRatePerWindow: getInt("RL_REGISTER_MAX", 5),
			RegisterWindow:        getDuration("RL_REGISTER_WINDOW", time.Minute),
			RefreshRatePerWindow:  getInt("RL_REFRESH_MAX", 30),
			RefreshWindow:         getDuration("RL_REFRESH_WINDOW", time.Minute),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", ""),
			SMTPPort:     getInt("SMTP_PORT", 587),
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			SMTPFrom:     getEnv("SMTP_FROM", "no-reply@webhook.local"),
		},
		Payments: PaymentConfig{
			Provider:   getEnv("PAYMENT_PROVIDER", "mock"),
			MockSecret: getEnv("PAYMENT_MOCK_SECRET", "dev-mock-secret-change-me"),
		},
	}

	if cfg.DB.PrimaryDSN == "" {
		return nil, fmt.Errorf("config: DB_PRIMARY_DSN is required")
	}

	if cfg.Auth.JWTSecret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required")
	}
	if len(cfg.Auth.JWTSecret) < 32 {
		return nil, fmt.Errorf("config: JWT_SECRET must be at least 32 characters")
	}
	if cfg.Payments.Provider != "mock" {
		return nil, fmt.Errorf("config: PAYMENT_PROVIDER=%q not supported; only 'mock' is currently implemented", cfg.Payments.Provider)
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

type PaymentConfig struct {
	Provider   string // "mock" | "paystack" (only "mock" implemented)
	MockSecret string // HMAC secret for the mock provider's signatures
}