package generated

import (
	"context"
)

// InsertTrendSnapshotParams carries the insert arguments for
// InsertTrendSnapshot.
type InsertTrendSnapshotParams struct {
	TopicID    int64
	Keyword    string
	TrendScore float64
	GrowthRate float64
	Velocity   float64
}

const insertTrendSnapshot = `INSERT INTO trend_analytics (topic_id, keyword, trend_score, growth_rate, velocity)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, topic_id, keyword, trend_score, growth_rate, velocity, detected_at`

// InsertTrendSnapshot records one computed trend measurement.
func (q *Queries) InsertTrendSnapshot(ctx context.Context, arg InsertTrendSnapshotParams) (TrendAnalytic, error) {
	row := q.db.QueryRow(ctx, insertTrendSnapshot,
		arg.TopicID,
		arg.Keyword,
		arg.TrendScore,
		arg.GrowthRate,
		arg.Velocity,
	)
	var i TrendAnalytic
	err := row.Scan(&i.ID, &i.TopicID, &i.Keyword, &i.TrendScore, &i.GrowthRate, &i.Velocity, &i.DetectedAt)
	return i, err
}

const getLatestTrends = `SELECT id, topic_id, keyword, trend_score, growth_rate, velocity, detected_at
FROM trend_analytics
WHERE topic_id = $1
  AND detected_at >= (SELECT COALESCE(MAX(detected_at), to_timestamp(0)) FROM trend_analytics WHERE topic_id = $1) - interval '10 minutes'
ORDER BY trend_score DESC
LIMIT $2`

// GetLatestTrends returns rows from the most recent detection batch.
func (q *Queries) GetLatestTrends(ctx context.Context, topicID int64, limit int32) ([]TrendAnalytic, error) {
	rows, err := q.db.Query(ctx, getLatestTrends, topicID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TrendAnalytic{}
	for rows.Next() {
		var i TrendAnalytic
		if err := rows.Scan(&i.ID, &i.TopicID, &i.Keyword, &i.TrendScore, &i.GrowthRate, &i.Velocity, &i.DetectedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const trendHistory = `SELECT id, topic_id, keyword, trend_score, growth_rate, velocity, detected_at
FROM trend_analytics
WHERE topic_id = $1
ORDER BY detected_at DESC
LIMIT $2`

// TrendHistory returns the raw trend measurement history (newest first).
func (q *Queries) TrendHistory(ctx context.Context, topicID int64, limit int32) ([]TrendAnalytic, error) {
	rows, err := q.db.Query(ctx, trendHistory, topicID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TrendAnalytic{}
	for rows.Next() {
		var i TrendAnalytic
		if err := rows.Scan(&i.ID, &i.TopicID, &i.Keyword, &i.TrendScore, &i.GrowthRate, &i.Velocity, &i.DetectedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}
