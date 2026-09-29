package ai

import "time"

// SummaryInput is the deterministic analytics payload shared with the
// summarizer service. It carries only metrics that were already computed by
// the analytics engine.
type SummaryInput struct {
	TopicName       string             `json:"topic_name"`
	GeneratedAt     time.Time          `json:"generated_at"`
	WindowHours     int                `json:"window_hours"`
	TotalPosts      int64              `json:"total_posts"`
	TotalEngagement int64              `json:"total_engagement"`
	DistinctAuthors int64              `json:"distinct_authors"`
	AvgSentiment    float64            `json:"avg_sentiment"`
	EmotionShares   map[string]float64 `json:"emotion_shares"`
	TopEmotions     []string           `json:"top_emotions"`
	TopKeywords     []string           `json:"top_keywords"`
	TopInfluencers  []string           `json:"top_influencers"`
	Languages       map[string]float64 `json:"language_distribution"`
	GrowthPercent   float64            `json:"growth_percent"`
	NetworkNodes    int                `json:"network_nodes"`
	NetworkEdges    int                `json:"network_edges"`
	PeakActivity    string             `json:"peak_activity"`
}

// Numbers extracts every grounded numeric value for the guardrail.
func (in SummaryInput) Numbers() map[string]float64 {
	out := map[string]float64{
		"window_hours":     float64(in.WindowHours),
		"total_posts":      float64(in.TotalPosts),
		"total_engagement": float64(in.TotalEngagement),
		"distinct_authors": float64(in.DistinctAuthors),
		"avg_sentiment":    in.AvgSentiment,
		"growth_percent":   in.GrowthPercent,
		"network_nodes":    float64(in.NetworkNodes),
		"network_edges":    float64(in.NetworkEdges),
	}
	for key, value := range in.EmotionShares {
		out["emotion_share_"+key] = value
	}
	for key, value := range in.Languages {
		out["language_share_"+key] = value
	}
	return out
}
