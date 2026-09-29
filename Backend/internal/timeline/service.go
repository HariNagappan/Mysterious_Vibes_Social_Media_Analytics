package timeline

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Domain errors surfaced to handlers.
var (
	// ErrTopicNotFound is returned when the topic id does not exist.
	ErrTopicNotFound = errors.New("timeline: topic not found")
	// ErrInvalidWindow is returned when from >= to.
	ErrInvalidWindow = errors.New("timeline: invalid time window")
)

const (
	defaultWindow = 7 * 24 * time.Hour
	maxPosts      = 500
)

// Service implements the timeline read model.
type Service struct {
	repo *Repository
}

// NewService wires the timeline service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Get assembles the timeline: posts, growth buckets and important events.
func (s *Service) Get(ctx context.Context, q Query) (*Timeline, error) {
	if _, err := s.repo.GetTopic(ctx, q.TopicID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTopicNotFound
		}
		return nil, err
	}

	from, to, err := resolveWindow(q)
	if err != nil {
		return nil, err
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > maxPosts {
		limit = maxPosts
	}

	posts, err := s.repo.ListPosts(ctx, generated.ListPostsByTopicParams{
		TopicID: q.TopicID,
		Since:   pgtype.Timestamptz{Time: from, Valid: true},
		Until:   pgtype.Timestamptz{Time: to, Valid: true},
		Limit:   limit,
		Offset:  q.Offset,
	})
	if err != nil {
		return nil, err
	}

	growth, err := s.repo.PostsPerBucket(ctx, generated.PostsPerBucketParams{
		TopicID:       q.TopicID,
		Since:         from,
		Until:         to,
		BucketMinutes: bucketMinutesFor(from, to),
	})
	if err != nil {
		return nil, err
	}

	top, err := s.repo.TopPosts(ctx, q.TopicID, from, to, 5)
	if err != nil {
		return nil, err
	}
	events := make([]ImportantEvent, 0, len(top))
	for _, p := range top {
		events = append(events, ImportantEvent{
			PostID:     p.ID,
			Author:     p.AuthorReference,
			Content:    p.Content,
			Timestamp:  p.Timestamp,
			Engagement: p.EngagementCount,
			Platform:   p.Platform,
		})
	}

	return &Timeline{
		TopicID:         q.TopicID,
		Window:          Window{From: from, To: to},
		Posts:           posts,
		Growth:          growth,
		ImportantEvents: events,
	}, nil
}

func resolveWindow(q Query) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	to := q.To
	if !q.HasTo {
		to = now
	}
	from := q.From
	if !q.HasFrom {
		from = to.Add(-defaultWindow)
	}
	if !from.Before(to) {
		return time.Time{}, time.Time{}, ErrInvalidWindow
	}
	return from, to, nil
}

// bucketMinutesFor picks a chart-friendly bucket size for the window span.
func bucketMinutesFor(from, to time.Time) int32 {
	span := to.Sub(from)
	switch {
	case span <= 48*time.Hour:
		return 60 // hourly
	case span <= 10*24*time.Hour:
		return 720 // 12-hourly
	default:
		return 1440 // daily
	}
}
