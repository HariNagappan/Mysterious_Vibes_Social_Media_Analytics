package analytics

import (
	"context"
	"time"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Snapshot bundles the headline aggregates behind a dashboard response.
type Snapshot struct {
	Overview   generated.TopicOverview   `json:"overview"`
	Platforms  []generated.PlatformCount `json:"platforms"`
	TopAuthors []generated.TopAuthor     `json:"top_authors"`
	PerDay     []generated.DayCount      `json:"per_day"`
}

// Aggregator runs the aggregate queries for the dashboard.
type Aggregator struct {
	queries *generated.Queries
}

// NewAggregator wraps the generated queries.
func NewAggregator(queries *generated.Queries) *Aggregator {
	return &Aggregator{queries: queries}
}

// Snapshot computes all headline aggregates for a topic.
func (a *Aggregator) Snapshot(ctx context.Context, topicID int64, sinceDays int32) (Snapshot, error) {
	overview, err := a.queries.TopicOverview(ctx, topicID)
	if err != nil {
		return Snapshot{}, err
	}
	platforms, err := a.queries.PlatformBreakdown(ctx, topicID)
	if err != nil {
		return Snapshot{}, err
	}
	topAuthors, err := a.queries.TopAuthors(ctx, topicID, 10)
	if err != nil {
		return Snapshot{}, err
	}
	since := time.Now().UTC().AddDate(0, 0, -int(sinceDays))
	perDay, err := a.queries.PostsPerDay(ctx, topicID, since)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{
		Overview:   overview,
		Platforms:  platforms,
		TopAuthors: topAuthors,
		PerDay:     perDay,
	}, nil
}
