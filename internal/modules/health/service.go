package health

import (
	"context"

	"gateway/webhook/internal/database"

	apperr "gateway/webhook/internal/commons/errors"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service { return &Service{db: db} }

type ReadinessReport struct {
	Database string `json:"database"`
}

func (s *Service) Readiness(ctx context.Context) (*ReadinessReport, error) {
	report := &ReadinessReport{Database: "up"}

	if err := s.db.PingPrimary(ctx); err != nil {
		report.Database = "down"
		return report, apperr.NewService(
			"health", apperr.KindUnavailable,
			"health.database_down", "database unavailable", err,
		)
	}
	return report, nil
}
