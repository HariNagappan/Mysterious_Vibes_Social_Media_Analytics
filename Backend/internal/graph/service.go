package graph

import (
	"context"
	"time"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

const topInfluencerLimit = 10

// Service implements the network read model.
type Service struct {
	queries *generated.Queries
}

// NewService wires the graph read service.
func NewService(queries *generated.Queries) *Service {
	return &Service{queries: queries}
}

// Network returns nodes, edges, influence scores and community structure for
// a topic's interaction graph.
func (s *Service) Network(ctx context.Context, topicID int64, nodeLimit, edgeLimit int32) (*Network, error) {
	nodes, err := s.queries.ListGraphNodesByTopic(ctx, topicID, nodeLimit)
	if err != nil {
		return nil, err
	}
	edges, err := s.queries.ListGraphEdgesByTopic(ctx, topicID, edgeLimit)
	if err != nil {
		return nil, err
	}

	top := make([]Influencer, 0, topInfluencerLimit)
	for _, n := range nodes {
		if len(top) >= topInfluencerLimit {
			break
		}
		top = append(top, Influencer{
			NodeID:         n.ID,
			Reference:      n.ExternalUserReference,
			DisplayName:    n.DisplayName,
			Platform:       n.Platform,
			InfluenceScore: n.InfluenceScore,
			CommunityID:    n.CommunityID,
		})
	}

	communities := map[int32]struct{}{}
	for _, n := range nodes {
		communities[n.CommunityID] = struct{}{}
	}

	return &Network{
		TopicID:        topicID,
		Nodes:          nodes,
		Edges:          edges,
		TopInfluencers: top,
		CommunityCount: len(communities),
		GeneratedAt:    time.Now().UTC(),
	}, nil
}
