package graph

import (
	"time"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// NodeInput is one node to upsert into the influence graph.
type NodeInput struct {
	Reference   string
	DisplayName string
	Platform    string
}

// EdgeInput is one interaction edge from the builders perspective.
type EdgeInput struct {
	SourceRef       string
	TargetRef       string
	InteractionType string
	Weight          int32
}

// EdgeRef is a resolved numeric edge used by the analysis algorithms.
type EdgeRef struct {
	Source int64
	Target int64
	Weight float64
}

// Influencer is one ranked account in the influence network.
type Influencer struct {
	NodeID         int64   `json:"node_id"`
	Reference      string  `json:"reference"`
	DisplayName    string  `json:"display_name"`
	Platform       string  `json:"platform"`
	InfluenceScore float64 `json:"influence_score"`
	CommunityID    int32   `json:"community_id"`
}

// Network is the GET /topics/:id/network payload.
type Network struct {
	TopicID        int64                 `json:"topic_id"`
	Nodes          []generated.GraphNode `json:"nodes"`
	Edges          []generated.GraphEdge `json:"edges"`
	TopInfluencers []Influencer          `json:"top_influencers"`
	CommunityCount int                   `json:"community_count"`
	GeneratedAt    time.Time             `json:"generated_at"`
}
