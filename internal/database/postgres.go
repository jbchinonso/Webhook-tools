// Package database owns the Postgres connection pool and exposes it behind
// a single DB handle.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"gateway/webhook/internal/config"
)

type DB struct {
	Primary *gorm.DB

	primarySQL *sql.DB
}

// New opens the primary connection pool.
// log is the application's single *slog.Logger — GORM logs its SQL activity
// through it too, rather than through a separate logger of its own, so every
// log line (HTTP, app, SQL) shares one output format and one request_id.
func New(cfg config.DBConfig, log *slog.Logger, logLevel string) (*DB, error) {
	primary, primarySQL, err := open(cfg.PrimaryDSN, cfg, log, logLevel)
	if err != nil {
		return nil, fmt.Errorf("database: open primary: %w", err)
	}

	return &DB{
		Primary:    primary,
		primarySQL: primarySQL,
	}, nil
}

func open(dsn string, cfg config.DBConfig, log *slog.Logger, logLevel string) (*gorm.DB, *sql.DB, error) {
	gormLogger := gormlogger.NewSlogLogger(log, gormlogger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      parseGormLevel(logLevel),
	})

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return gdb, sqlDB, nil
}

func parseGormLevel(level string) gormlogger.LogLevel {
	switch level {
	case "debug":
		return gormlogger.Info
	case "warn":
		return gormlogger.Warn
	case "error":
		return gormlogger.Error
	default:
		return gormlogger.Warn
	}
}

func (d *DB) PingPrimary(ctx context.Context) error {
	c, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return d.primarySQL.PingContext(c)
}

func (d *DB) Close() error {
	return d.primarySQL.Close()
}
