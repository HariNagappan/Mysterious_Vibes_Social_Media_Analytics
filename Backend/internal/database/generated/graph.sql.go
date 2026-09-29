package generated

import (
	"context"
)

// UpsertGraphNodeParams carries the insert arguments for UpsertGraphNode.
type UpsertGraphNodeParams struct {
	ExternalUserReference string
	DisplayName           string
	Platform              string
}

const upsertGraphNode = `INSERT INTO graph_nodes (external_user_reference, display_name, platform)
VALUES ($1, $2, $3)
ON CONFLICT (external_user_reference) DO UPDATE
SET display_name = CASE WHEN EXCLUDED.display_name <> '' THEN EXCLUDED.display_name ELSE graph_nodes.display_name END,
    platform = CASE WHEN EXCLUDED.platform <> '' THEN EXCLUDED.platform ELSE graph_nodes.platform END,
    updated_at = now()
RETURNING id, external_user_reference, display_name, platform, influence_score, community_id, updated_at`

// UpsertGraphNode creates or refreshes an author node and returns it.
func (q *Queries) UpsertGraphNode(ctx context.Context, arg UpsertGraphNodeParams) (GraphNode, error) {
	row := q.db.QueryRow(ctx, upsertGraphNode, arg.ExternalUserReference, arg.DisplayName, arg.Platform)
	var i GraphNode
	err := row.Scan(&i.ID, &i.ExternalUserReference, &i.DisplayName, &i.Platform, &i.InfluenceScore, &i.CommunityID, &i.UpdatedAt)
	return i, err
}

// UpsertGraphEdgeParams carries the insert arguments for UpsertGraphEdge.
type UpsertGraphEdgeParams struct {
	TopicID         int64
	SourceNode      int64
	TargetNode      int64
	InteractionType string
	Weight          int32
}

const upsertGraphEdge = `INSERT INTO graph_edges (topic_id, source_node, target_node, interaction_type, weight)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (topic_id, source_node, target_node, interaction_type) DO UPDATE
SET weight = graph_edges.weight + EXCLUDED.weight
RETURNING id, topic_id, source_node, target_node, interaction_type, weight`

// UpsertGraphEdge accumulates interaction weight between two nodes.
func (q *Queries) UpsertGraphEdge(ctx context.Context, arg UpsertGraphEdgeParams) (GraphEdge, error) {
	row := q.db.QueryRow(ctx, upsertGraphEdge,
		arg.TopicID,
		arg.SourceNode,
		arg.TargetNode,
		arg.InteractionType,
		arg.Weight,
	)
	var i GraphEdge
	err := row.Scan(&i.ID, &i.TopicID, &i.SourceNode, &i.TargetNode, &i.InteractionType, &i.Weight)
	return i, err
}

// UpdateNodeInfluenceParams carries the update arguments for
// UpdateNodeInfluence.
type UpdateNodeInfluenceParams struct {
	ID             int64
	InfluenceScore float64
	CommunityID    int32
}

const updateNodeInfluence = `UPDATE graph_nodes
SET influence_score = $2, community_id = $3, updated_at = now()
WHERE id = $1`

// UpdateNodeInfluence persists pagerank + community assignment for a node.
func (q *Queries) UpdateNodeInfluence(ctx context.Context, arg UpdateNodeInfluenceParams) error {
	_, err := q.db.Exec(ctx, updateNodeInfluence, arg.ID, arg.InfluenceScore, arg.CommunityID)
	return err
}

const listGraphNodesByTopic = `SELECT n.id, n.external_user_reference, n.display_name, n.platform,
	n.influence_score, n.community_id, n.updated_at
FROM graph_nodes n
WHERE n.id IN (
	SELECT source_node FROM graph_edges WHERE topic_id = $1
	UNION
	SELECT target_node FROM graph_edges WHERE topic_id = $1
)
ORDER BY n.influence_score DESC, n.id ASC
LIMIT $2`

// ListGraphNodesByTopic returns the nodes participating in a topic's
// interaction graph, ranked by influence.
func (q *Queries) ListGraphNodesByTopic(ctx context.Context, topicID int64, limit int32) ([]GraphNode, error) {
	rows, err := q.db.Query(ctx, listGraphNodesByTopic, topicID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GraphNode{}
	for rows.Next() {
		var i GraphNode
		if err := rows.Scan(&i.ID, &i.ExternalUserReference, &i.DisplayName, &i.Platform, &i.InfluenceScore, &i.CommunityID, &i.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const listGraphEdgesByTopic = `SELECT id, topic_id, source_node, target_node, interaction_type, weight
FROM graph_edges
WHERE topic_id = $1
ORDER BY weight DESC
LIMIT $2`

// ListGraphEdgesByTopic returns the strongest interaction edges for a topic.
func (q *Queries) ListGraphEdgesByTopic(ctx context.Context, topicID int64, limit int32) ([]GraphEdge, error) {
	rows, err := q.db.Query(ctx, listGraphEdgesByTopic, topicID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GraphEdge{}
	for rows.Next() {
		var i GraphEdge
		if err := rows.Scan(&i.ID, &i.TopicID, &i.SourceNode, &i.TargetNode, &i.InteractionType, &i.Weight); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}
