package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Subject is the logical stream an event belongs to. Subjects make the
// pipeline topology explicit and keep transport topics decoupled from event
// types (one subject can carry several event types).
type Subject string

const (
	// SubjectPosts carries raw ingestion events.
	SubjectPosts Subject = "posts"
	// SubjectSentiment carries post-sentiment-completed events.
	SubjectSentiment Subject = "sentiment"
	// SubjectAnalytics carries downstream analytics/report events.
	SubjectAnalytics Subject = "analytics"
)

// SubjectForEvent maps an event type onto its stream subject.
func SubjectForEvent(t EventType) Subject {
	switch t {
	case EventPostIngested:
		return SubjectPosts
	case EventSentimentCompleted:
		return SubjectSentiment
	default:
		return SubjectAnalytics
	}
}

// Producer publishes events onto the message queue. This interface is the
// Kafka seam: a Kafka producer can replace Redis Streams without touching any
// caller.
type Producer interface {
	Publish(ctx context.Context, subject Subject, evt Event) error
	Close() error
}

// RedisProducer implements Producer on top of Redis Streams.
type RedisProducer struct {
	client       *redis.Client
	streamPrefix string
	logger       *zap.Logger
}

// NewRedisProducer builds a Redis Streams producer.
func NewRedisProducer(client *redis.Client, streamPrefix string, logger *zap.Logger) *RedisProducer {
	return &RedisProducer{client: client, streamPrefix: streamPrefix, logger: logger}
}

// StreamName resolves the physical Redis stream name for a subject.
func (p *RedisProducer) StreamName(s Subject) string {
	return fmt.Sprintf("%s:stream:%s", p.streamPrefix, s)
}

// Publish appends the event to the stream for its subject. Streams are capped
// (approximate trim) so local environments do not grow unbounded.
func (p *RedisProducer) Publish(ctx context.Context, subject Subject, evt Event) error {
	values := map[string]any{
		"id":          evt.ID,
		"type":        string(evt.Type),
		"version":     evt.Version,
		"occurred_at": evt.OccurredAt.Format(time.RFC3339Nano),
		"payload":     string(evt.Payload),
	}
	_, err := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: p.StreamName(subject),
		MaxLen: 20000,
		Approx: true,
		Values: values,
	}).Result()
	if err != nil {
		return fmt.Errorf("pipeline: publish %s: %w", evt.Type, err)
	}
	return nil
}

// Close releases producer resources (Redis Streams needs none).
func (p *RedisProducer) Close() error { return nil }
