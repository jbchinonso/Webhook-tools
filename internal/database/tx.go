package database

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// WithTx runs fn inside a single database transaction against the primary.
// Any error returned by fn — or any panic inside — rolls the transaction back.
// The wrapped error is returned untouched so the service layer keeps its
// layered error type.
//
// This is the ONLY sanctioned way to mutate inventory, confirm bookings,
// issue tickets, and record payments.
func WithTx(ctx context.Context, db *DB, fn func(tx *gorm.DB) error) error {
	tx := db.Primary.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback().Error
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}

	return tx.Commit().Error
}