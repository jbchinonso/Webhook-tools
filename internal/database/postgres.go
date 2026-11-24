// Package database owns the Postgres connection pools — primary (writes)
// and replica (reads) — and exposes them behind a single DB handle.
//
// When ReplicaDSN is empty, Replica aliases Primary so code that reads from
// a replica is correct today and automatically benefits from a real replica
// once one is configured.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"gateway/webhook/internal/config"
)

type DB struct {
	Primary *gorm.DB
	Replica *gorm.DB

	primarySQL     *sql.DB
	replicaSQL     *sql.DB
	replicaIsAlias bool
}

func New(cfg config.DBConfig, logLevel string) (*DB, error) {
	primary, primarySQL, err := open(cfg.PrimaryDSN, cfg, logLevel)
	if err != nil {
		return nil, fmt.Errorf("database: open primary: %w", err)
	}

	if cfg.ReplicaDSN == "" {
		return &DB{
			Primary:        primary,
			Replica:        primary,
			primarySQL:     primarySQL,
			replicaSQL:     primarySQL,
			replicaIsAlias: true,
		}, nil
	}

	replica, replicaSQL, err := open(cfg.ReplicaDSN, cfg, logLevel)
	if err != nil {
		_ = primarySQL.Close()
		return nil, fmt.Errorf("database: open replica: %w", err)
	}

	return &DB{
		Primary:    primary,
		Replica:    replica,
		primarySQL: primarySQL,
		replicaSQL: replicaSQL,
	}, nil
}

func open(dsn string, cfg config.DBConfig, logLevel string) (*gorm.DB, *sql.DB, error) {
	lvl := parseGormLevel(logLevel)

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 gormlogger.Default.LogMode(lvl),
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

func (d *DB) PrimarySQL() *sql.DB { return d.primarySQL }
func (d *DB) ReplicaSQL() *sql.DB { return d.replicaSQL }

func (d *DB) PingPrimary(ctx context.Context) error {
	c, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return d.primarySQL.PingContext(c)
}

func (d *DB) PingReplica(ctx context.Context) error {
	if d.replicaIsAlias {
		// Same pool as primary; no need to ping twice.
		return nil
	}
	c, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return d.replicaSQL.PingContext(c)
}

func (d *DB) Close() error {
	if d.replicaIsAlias {
		return d.primarySQL.Close()
	}
	var firstErr error
	if err := d.primarySQL.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := d.replicaSQL.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}
