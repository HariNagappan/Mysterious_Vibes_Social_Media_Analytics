package demographics

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Service implements audience analytics.
type Service struct {
	repo *Repository
}

// NewService wires the demographics service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ComputeAndStore rebuilds the aggregated snapshot from the collected posts
// (scheduler job).
func (s *Service) ComputeAndStore(ctx context.Context, topicID int64) (Snapshot, error) {
	signals, err := s.repo.Signals(ctx, topicID, 5000)
	if err != nil {
		return Snapshot{}, err
	}
	snap := Estimate(signals)

	languageJSON, err := json.Marshal(snap.LanguageDistribution)
	if err != nil {
		return Snapshot{}, err
	}
	regionJSON, err := json.Marshal(snap.RegionDistribution)
	if err != nil {
		return Snapshot{}, err
	}
	interestJSON, err := json.Marshal(snap.InterestDistribution)
	if err != nil {
		return Snapshot{}, err
	}
	groupsJSON, err := json.Marshal(snap.AudienceGroups)
	if err != nil {
		return Snapshot{}, err
	}

	row, err := s.repo.Insert(ctx, generated.InsertDemographicSnapshotParams{
		TopicID:              topicID,
		LanguageDistribution: languageJSON,
		RegionDistribution:   regionJSON,
		InterestDistribution: interestJSON,
		AudienceGroups:       groupsJSON,
	})
	if err != nil {
		return Snapshot{}, err
	}
	snap.TopicID = topicID
	snap.SnapshotAt = row.SnapshotAt
	return snap, nil
}

// Get returns the latest stored snapshot, or an empty well-formed one when
// nothing has been computed yet.
func (s *Service) Get(ctx context.Context, topicID int64) (Snapshot, error) {
	row, err := s.repo.Latest(ctx, topicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Snapshot{
				TopicID:              topicID,
				LanguageDistribution: map[string]float64{},
				RegionDistribution:   map[string]float64{},
				InterestDistribution: map[string]float64{},
				AudienceGroups:       []AudienceGroup{},
			}, nil
		}
		return Snapshot{}, err
	}

	snap := Snapshot{
		TopicID:              row.TopicID,
		SnapshotAt:           row.SnapshotAt,
		LanguageDistribution: decodeFloatMap(row.LanguageDistribution),
		RegionDistribution:   decodeFloatMap(row.RegionDistribution),
		InterestDistribution: decodeFloatMap(row.InterestDistribution),
		AudienceGroups:       decodeGroups(row.AudienceGroups),
	}
	return snap, nil
}

func decodeFloatMap(raw []byte) map[string]float64 {
	out := map[string]float64{}
	if len(raw) == 0 {
		return out
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]float64{}
	}
	return out
}

func decodeGroups(raw []byte) []AudienceGroup {
	out := []AudienceGroup{}
	if len(raw) == 0 {
		return out
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return []AudienceGroup{}
	}
	return out
}
