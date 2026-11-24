package health

import (
	"context"
	"sync"

	"gateway/webhook/internal/database"

	apperr "gateway/webhook/internal/commons/errors"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service { return &Service{db: db} }

type ReadinessReport struct {
	Primary string `json:"primary"`
	Replica string `json:"replica"`
}

// Readiness checks both database handles concurrently. When no replica is
// configured, PingReplica is a no-op and the replica is reported as "up"
// via the primary alias.
func (s *Service) Readiness(ctx context.Context) (*ReadinessReport, error) {
	report := &ReadinessReport{Primary: "up", Replica: "up"}

	var (
		wg         sync.WaitGroup
		primaryErr error
		replicaErr error
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		primaryErr = s.db.PingPrimary(ctx)
	}()
	go func() {
		defer wg.Done()
		replicaErr = s.db.PingReplica(ctx)
	}()
	wg.Wait()

	if primaryErr != nil {
		report.Primary = "down"
		return report, apperr.NewService(
			"health", apperr.KindUnavailable,
			"health.primary_down", "primary database unavailable", primaryErr,
		)
	}
	if replicaErr != nil {
		report.Replica = "down"
		return report, apperr.NewService(
			"health", apperr.KindUnavailable,
			"health.replica_down", "replica database unavailable", replicaErr,
		)
	}
	return report, nil
}
