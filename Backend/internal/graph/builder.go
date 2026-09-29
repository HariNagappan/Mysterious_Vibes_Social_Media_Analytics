package graph

import (
	"context"

	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Builder reconstructs the interaction graph for a topic from collected
// posts: reply and mention relationships become weighted edges, authors and
// mentioned accounts become nodes.
type Builder struct {
	queries *generated.Queries
	logger  *zap.Logger
}

// NewBuilder wires the graph builder.
func NewBuilder(queries *generated.Queries, logger *zap.Logger) *Builder {
	return &Builder{queries: queries, logger: logger}
}

// BuildForTopic rebuilds the graph and refreshes influence scores and
// community assignments. Returns the number of nodes and edges processed in
// this pass.
func (b *Builder) BuildForTopic(ctx context.Context, topicID int64) (int, int, error) {
	posts, err := b.queries.ListPostsByTopic(ctx, generated.ListPostsByTopicParams{
		TopicID: topicID,
		Limit:   5000,
	})
	if err != nil {
		return 0, 0, err
	}
	if len(posts) == 0 {
		return 0, 0, nil
	}

	authorByExternal := make(map[string]string, len(posts))
	displayByRef := map[string]string{}
	platformByRef := map[string]string{}

	for _, p := range posts {
		authorByExternal[p.ExternalID] = p.AuthorReference
		if _, seen := displayByRef[p.AuthorReference]; !seen {
			displayByRef[p.AuthorReference] = p.AuthorDisplayName
			platformByRef[p.AuthorReference] = p.Platform
		}
	}

	type edgeKey struct {
		source string
		target string
		kind   string
	}
	weights := map[edgeKey]int32{}

	for _, p := range posts {
		source := p.AuthorReference
		if p.ReplyToExternalID != "" {
			if target, ok := authorByExternal[p.ReplyToExternalID]; ok && target != source {
				weights[edgeKey{source, target, "reply"}]++
			}
		}
		for _, mention := range p.MentionRefs {
			if mention == "" || mention == source {
				continue
			}
			weights[edgeKey{source, mention, "mention"}]++
		}
	}

	// Resolve every referenced account to a node id.
	refs := map[string]struct{}{}
	for ref := range displayByRef {
		refs[ref] = struct{}{}
	}
	for k := range weights {
		refs[k.source] = struct{}{}
		refs[k.target] = struct{}{}
	}

	nodeIDs := make(map[string]int64, len(refs))
	var ids []int64
	for ref := range refs {
		node, err := b.queries.UpsertGraphNode(ctx, generated.UpsertGraphNodeParams{
			ExternalUserReference: ref,
			DisplayName:           displayByRef[ref],
			Platform:              platformByRef[ref],
		})
		if err != nil {
			return 0, 0, err
		}
		nodeIDs[ref] = node.ID
		ids = append(ids, node.ID)
	}

	// Persist edges and keep a resolved copy for the analysis algorithms.
	var edgeRefs []EdgeRef
	for k, w := range weights {
		sourceID, okSource := nodeIDs[k.source]
		targetID, okTarget := nodeIDs[k.target]
		if !okSource || !okTarget {
			continue
		}
		if _, err := b.queries.UpsertGraphEdge(ctx, generated.UpsertGraphEdgeParams{
			TopicID:         topicID,
			SourceNode:      sourceID,
			TargetNode:      targetID,
			InteractionType: k.kind,
			Weight:          w,
		}); err != nil {
			return 0, 0, err
		}
		edgeRefs = append(edgeRefs, EdgeRef{Source: sourceID, Target: targetID, Weight: float64(w)})
	}

	// Influence + communities.
	ranks := PageRank(ids, edgeRefs, 0.85, 30)
	communities := LabelPropagation(ids, edgeRefs, 20)
	for _, id := range ids {
		if err := b.queries.UpdateNodeInfluence(ctx, generated.UpdateNodeInfluenceParams{
			ID:             id,
			InfluenceScore: ranks[id],
			CommunityID:    int32(communities[id]),
		}); err != nil {
			return 0, 0, err
		}
	}

	b.logger.Debug("graph: rebuilt",
		zap.Int64("topic_id", topicID),
		zap.Int("nodes", len(nodeIDs)),
		zap.Int("edges", len(edgeRefs)))
	return len(nodeIDs), len(edgeRefs), nil
}
