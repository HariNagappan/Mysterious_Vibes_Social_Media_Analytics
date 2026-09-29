// Package pipeline implements the asynchronous processing backbone of
// PulseGraph: a typed event envelope, a producer/consumer abstraction over
// Redis Streams, and the deterministic building blocks the workers use
// (language detection, stage handlers).
//
// The transport is deliberately hidden behind Producer/Consumer interfaces so
// Redis Streams can be swapped for Kafka without touching call sites.
package pipeline

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// EventType enumerates the logical stages of the processing pipeline.
type EventType string

const (
	// EventPostIngested is emitted when a normalized post is stored.
	EventPostIngested EventType = "post.ingested"
	// EventSentimentCompleted is emitted when sentiment scoring finished.
	EventSentimentCompleted EventType = "sentiment.completed"
	// EventEmbeddingCompleted is emitted when the post embedding is stored.
	EventEmbeddingCompleted EventType = "embedding.completed"
	// EventTrendUpdated is emitted by the scheduler after trend recomputation.
	EventTrendUpdated EventType = "trend.updated"
	// EventGraphUpdated is emitted by the scheduler after graph recomputation.
	EventGraphUpdated EventType = "graph.updated"
	// EventReportGenerated is emitted when an AI summary is produced.
	EventReportGenerated EventType = "report.generated"
)

// Event is the canonical envelope for every message that travels through the
// pipeline.
type Event struct {
	ID         string          `json:"id"`
	Type       EventType       `json:"type"`
	Version    int             `json:"version"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
}

// NewEvent wraps a typed payload into an envelope with a random ID.
func NewEvent(eventType EventType, payload any) (Event, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Event{}, fmt.Errorf("pipeline: marshal payload for %s: %w", eventType, err)
	}
	return Event{
		ID:         randomID(),
		Type:       eventType,
		Version:    1,
		OccurredAt: time.Now().UTC(),
		Payload:    raw,
	}, nil
}

// Decode unmarshals the event payload into v.
func (e Event) Decode(v any) error {
	if err := json.Unmarshal(e.Payload, v); err != nil {
		return fmt.Errorf("pipeline: decode %s payload: %w", e.Type, err)
	}
	return nil
}

func randomID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

// ── Typed payloads ──────────────────────────────────────────────────────────

// PostIngestedPayload accompanies EventPostIngested.
type PostIngestedPayload struct {
	PostID   int64  `json:"post_id"`
	TopicID  int64  `json:"topic_id"`
	Platform string `json:"platform"`
}

// SentimentCompletedPayload accompanies EventSentimentCompleted.
type SentimentCompletedPayload struct {
	PostID  int64   `json:"post_id"`
	TopicID int64   `json:"topic_id"`
	Emotion string  `json:"emotion"`
	Score   float64 `json:"score"`
}

// EmbeddingCompletedPayload accompanies EventEmbeddingCompleted.
type EmbeddingCompletedPayload struct {
	PostID  int64 `json:"post_id"`
	TopicID int64 `json:"topic_id"`
	Dims    int   `json:"dims"`
}

// TrendUpdatedPayload accompanies EventTrendUpdated.
type TrendUpdatedPayload struct {
	TopicID int64 `json:"topic_id"`
	Points  int   `json:"points"`
}

// GraphUpdatedPayload accompanies EventGraphUpdated.
type GraphUpdatedPayload struct {
	TopicID int64 `json:"topic_id"`
	Nodes   int   `json:"nodes"`
	Edges   int   `json:"edges"`
}
