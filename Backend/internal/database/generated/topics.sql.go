package generated

import (
	"context"
)

// CreateTopicParams carries the insert arguments for CreateTopic.
type CreateTopicParams struct {
	Name        string
	Keywords    []string
	Description string
}

const createTopic = `INSERT INTO topics (name, keywords, description)
VALUES ($1, $2, $3)
RETURNING id, name, keywords, description, created_at, updated_at`

// CreateTopic registers a monitored conversation.
func (q *Queries) CreateTopic(ctx context.Context, arg CreateTopicParams) (Topic, error) {
	row := q.db.QueryRow(ctx, createTopic, arg.Name, arg.Keywords, arg.Description)
	var i Topic
	err := row.Scan(&i.ID, &i.Name, &i.Keywords, &i.Description, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

const getTopicByID = `SELECT id, name, keywords, description, created_at, updated_at
FROM topics
WHERE id = $1`

// GetTopicByID loads one topic.
func (q *Queries) GetTopicByID(ctx context.Context, id int64) (Topic, error) {
	row := q.db.QueryRow(ctx, getTopicByID, id)
	var i Topic
	err := row.Scan(&i.ID, &i.Name, &i.Keywords, &i.Description, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

// ListTopicsParams carries pagination arguments for ListTopics.
type ListTopicsParams struct {
	Limit  int32
	Offset int32
}

const listTopics = `SELECT id, name, keywords, description, created_at, updated_at
FROM topics
ORDER BY created_at DESC
LIMIT $1 OFFSET $2`

// ListTopics returns topics in reverse chronological order.
func (q *Queries) ListTopics(ctx context.Context, arg ListTopicsParams) ([]Topic, error) {
	rows, err := q.db.Query(ctx, listTopics, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Topic{}
	for rows.Next() {
		var i Topic
		if err := rows.Scan(&i.ID, &i.Name, &i.Keywords, &i.Description, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const countTopics = `SELECT COUNT(*)::bigint FROM topics`

// CountTopics returns the total number of monitored topics.
func (q *Queries) CountTopics(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, countTopics)
	var count int64
	err := row.Scan(&count)
	return count, err
}
