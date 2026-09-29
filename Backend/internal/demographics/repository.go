package demographics

import (
	"context"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Repository persists and reads aggregated audience snapshots.
type Repository struct {
	queries *generated.Queries
}

// NewRepository wraps the generated queries.
func NewRepository(queries *generated.Queries) *Repository {
	return &Repository{queries: queries}
}

// Signals loads the compact post projections the estimator consumes.
func (r *Repository) Signals(ctx context.Context, topicID int64, limit int32) ([]generated.PostSignal, error) {
	return r.queries.ListPostSignals(ctx, topicID, limit)
}

// Insert stores one aggregated snapshot.
func (r *Repository) Insert(ctx context.Context, arg generated.InsertDemographicSnapshotParams) (generated.DemographicAnalytics, error) {
	return r.queries.InsertDemographicSnapshot(ctx, arg)
}

// Latest loads the freshest stored snapshot.
func (r *Repository) Latest(ctx context.Context, topicID int64) (generated.DemographicAnalytics, error) {
	return r.queries.GetLatestDemographicSnapshot(ctx, topicID)
}
