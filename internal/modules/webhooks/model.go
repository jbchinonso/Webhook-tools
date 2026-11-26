package webhooks

import (
	"time"

	"github.com/google/uuid"
)

// Endpoint is a capture point that a third-party system sends webhooks to,
// reached at a stable public path keyed by Slug (see init.go).
type Endpoint struct {
	ID        uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	Name      string    `gorm:"column:name"`
	Slug      string    `gorm:"column:slug"`
	Secret    string    `gorm:"column:secret"`
	Status    string    `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Endpoint) TableName() string { return "webhook_endpoints" }

const (
	EndpointStatusActive   = "active"
	EndpointStatusDisabled = "disabled"
)

// Event is one captured inbound delivery for an Endpoint.
//
// Forwarding a captured Event to a destination, and replaying it, is
// deliberately out of scope for this skeleton — see the module-level
// TODO in repository.go.
type Event struct {
	ID         uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	EndpointID uuid.UUID `gorm:"column:endpoint_id;type:uuid"`
	Method     string    `gorm:"column:method"`
	Headers    []byte    `gorm:"column:headers;type:jsonb"`
	Body       string    `gorm:"column:body"`
	SourceIP   string    `gorm:"column:source_ip"`
	ReceivedAt time.Time `gorm:"column:received_at"`
}

func (Event) TableName() string { return "webhook_events" }
