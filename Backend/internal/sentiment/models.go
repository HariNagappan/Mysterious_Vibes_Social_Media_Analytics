package sentiment

import "time"

// Result is one post's sentiment verdict.
type Result struct {
	Emotion      string  `json:"emotion"`
	Score        float64 `json:"sentiment_score"`
	Confidence   float64 `json:"confidence"`
	ModelVersion string  `json:"model_version"`
}

// DistributionItem is one slice of the emotion distribution.
type DistributionItem struct {
	Emotion string  `json:"emotion"`
	Count   int64   `json:"count"`
	Share   float64 `json:"share"`
}

// TimelinePoint is one bucket of the sentiment timeline.
type TimelinePoint struct {
	Bucket    time.Time `json:"bucket"`
	AvgScore  float64   `json:"avg_score"`
	PostCount int64     `json:"post_count"`
	MinScore  float64   `json:"min_score"`
	MaxScore  float64   `json:"max_score"`
}

// TopicSentiment is the GET /topics/:id/sentiment payload.
type TopicSentiment struct {
	TopicID       int64              `json:"topic_id"`
	Window        Window             `json:"window"`
	OverallScore  float64            `json:"overall_score"`
	AnalyzedPosts int64              `json:"analyzed_posts"`
	TotalPosts    int64              `json:"total_posts"`
	Distribution  []DistributionItem `json:"distribution"`
	Timeline      []TimelinePoint    `json:"timeline"`
}

// Window is the resolved time window of a sentiment response.
type Window struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}
