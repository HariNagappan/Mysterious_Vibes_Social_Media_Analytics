package generated

import (
	"context"
)

// CreateSourceParams carries the insert arguments for CreateSource.
type CreateSourceParams struct {
	Platform      string
	SourceName    string
	Configuration []byte
	Status        string
}

const createSource = `INSERT INTO social_sources (platform, source_name, configuration, status)
VALUES ($1, $2, $3, $4)
RETURNING id, platform, source_name, configuration, status, created_at`

// CreateSource registers a connected platform source.
func (q *Queries) CreateSource(ctx context.Context, arg CreateSourceParams) (SocialSource, error) {
	row := q.db.QueryRow(ctx, createSource, arg.Platform, arg.SourceName, arg.Configuration, arg.Status)
	var i SocialSource
	err := row.Scan(&i.ID, &i.Platform, &i.SourceName, &i.Configuration, &i.Status, &i.CreatedAt)
	return i, err
}

const listSources = `SELECT id, platform, source_name, configuration, status, created_at
FROM social_sources
ORDER BY created_at DESC`

// ListSources returns every registered source.
func (q *Queries) ListSources(ctx context.Context) ([]SocialSource, error) {
	rows, err := q.db.Query(ctx, listSources)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SocialSource{}
	for rows.Next() {
		var i SocialSource
		if err := rows.Scan(&i.ID, &i.Platform, &i.SourceName, &i.Configuration, &i.Status, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const updateSourceStatus = `UPDATE social_sources
SET status = $2
WHERE id = $1
RETURNING id, platform, source_name, configuration, status, created_at`

// UpdateSourceStatus changes the lifecycle status of a source.
func (q *Queries) UpdateSourceStatus(ctx context.Context, id int64, status string) (SocialSource, error) {
	row := q.db.QueryRow(ctx, updateSourceStatus, id, status)
	var i SocialSource
	err := row.Scan(&i.ID, &i.Platform, &i.SourceName, &i.Configuration, &i.Status, &i.CreatedAt)
	return i, err
}
