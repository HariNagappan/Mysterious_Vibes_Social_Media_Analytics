package timeline

import (
	"time"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Query carries the filter parameters for a timeline request.
type Query struct {
	TopicID int64
	From    time.Time
	HasFrom bool
	To      time.Time
	HasTo   bool
	Limit   int32
	Offset  int32
}

// ImportantEvent is a high-engagement post surfaced as a discussion marker
// (spikes, viral posts, key moments).
type ImportantEvent struct {
	PostID     int64     `json:"post_id"`
	Author     string    `json:"author"`
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
	Engagement int32     `json:"engagement_count"`
	Platform   string    `json:"platform"`
}

// Window is the resolved time window of a timeline response.
type Window struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// Timeline is the GET /topics/:id/timeline payload: chronological posts,
// discussion growth and important events.
type Timeline struct {
	TopicID         int64                   `json:"topic_id"`
	Window          Window                  `json:"window"`
	Posts           []generated.SocialPost  `json:"posts"`
	Growth          []generated.PostsBucket `json:"growth"`
	ImportantEvents []ImportantEvent        `json:"important_events"`
}
