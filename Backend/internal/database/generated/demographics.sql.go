package generated

import (
	"context"
)

// InsertDemographicSnapshotParams carries the insert arguments for
// InsertDemographicSnapshot. JSONB columns are passed as pre-encoded JSON.
type InsertDemographicSnapshotParams struct {
	TopicID              int64
	LanguageDistribution []byte
	RegionDistribution   []byte
	InterestDistribution []byte
	AudienceGroups       []byte
}

const insertDemographicSnapshot = `INSERT INTO demographic_analytics (
	topic_id, language_distribution, region_distribution, interest_distribution, audience_groups
)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, topic_id, snapshot_at, language_distribution, region_distribution, interest_distribution, audience_groups`

// InsertDemographicSnapshot stores one aggregated audience snapshot. Only
// aggregated information is ever stored — never per-user data.
func (q *Queries) InsertDemographicSnapshot(ctx context.Context, arg InsertDemographicSnapshotParams) (DemographicAnalytics, error) {
	row := q.db.QueryRow(ctx, insertDemographicSnapshot,
		arg.TopicID,
		arg.LanguageDistribution,
		arg.RegionDistribution,
		arg.InterestDistribution,
		arg.AudienceGroups,
	)
	var i DemographicAnalytics
	err := row.Scan(&i.ID, &i.TopicID, &i.SnapshotAt, &i.LanguageDistribution, &i.RegionDistribution, &i.InterestDistribution, &i.AudienceGroups)
	return i, err
}

const getLatestDemographicSnapshot = `SELECT id, topic_id, snapshot_at, language_distribution, region_distribution, interest_distribution, audience_groups
FROM demographic_analytics
WHERE topic_id = $1
ORDER BY snapshot_at DESC
LIMIT 1`

// GetLatestDemographicSnapshot loads the freshest audience snapshot.
func (q *Queries) GetLatestDemographicSnapshot(ctx context.Context, topicID int64) (DemographicAnalytics, error) {
	row := q.db.QueryRow(ctx, getLatestDemographicSnapshot, topicID)
	var i DemographicAnalytics
	err := row.Scan(&i.ID, &i.TopicID, &i.SnapshotAt, &i.LanguageDistribution, &i.RegionDistribution, &i.InterestDistribution, &i.AudienceGroups)
	return i, err
}
