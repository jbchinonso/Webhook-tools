package users

import (
	"time"

	"github.com/google/uuid"
)

// Role and Status are typed so illegal values can't come from Go code.
type Role string

const (
	RoleAdmin Role = "admin"
	RoleSuperAdmin     Role = "super-admin"
)

func (r Role) Valid() bool {
	switch r {
	case RoleAdmin, RoleSuperAdmin:

		return true
	}
	return false
}

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusDeleted   Status = "deleted"
)

// User is the persistence model. It carries password_hash, which never
// leaves this module except to the auth service's verifier.
type User struct {
	ID              uuid.UUID  `gorm:"column:id;primaryKey;type:uuid"`
	Email           string     `gorm:"column:email"`
	PasswordHash    string     `gorm:"column:password_hash"`
	Role            Role       `gorm:"column:role"`
	Status          Status     `gorm:"column:status"`
	FullName        string     `gorm:"column:full_name"`
	Phone           string     `gorm:"column:phone"`
	EmailVerifiedAt *time.Time `gorm:"column:email_verified_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (User) TableName() string { return "users" }