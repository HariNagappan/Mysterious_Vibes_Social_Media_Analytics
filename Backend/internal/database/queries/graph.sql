-- name: UpsertGraphNode :one
INSERT INTO graph_nodes (external_user_reference, display_name, platform)
VALUES ($1, $2, $3)
ON CONFLICT (external_user_reference) DO UPDATE
SET display_name = CASE WHEN EXCLUDED.display_name <> '' THEN EXCLUDED.display_name ELSE graph_nodes.display_name END,
    platform = CASE WHEN EXCLUDED.platform <> '' THEN EXCLUDED.platform ELSE graph_nodes.platform END,
    updated_at = now()
RETURNING id, external_user_reference, display_name, platform, influence_score, community_id, updated_at;

-- name: UpsertGraphEdge :one
INSERT INTO graph_edges (topic_id, source_node, target_node, interaction_type, weight)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (topic_id, source_node, target_node, interaction_type) DO UPDATE
SET weight = graph_edges.weight + EXCLUDED.weight
RETURNING id, topic_id, source_node, target_node, interaction_type, weight;

-- name: UpdateNodeInfluence :exec
UPDATE graph_nodes
SET influence_score = $2, community_id = $3, updated_at = now()
WHERE id = $1;

-- name: ListGraphNodesByTopic :many
SELECT n.id, n.external_user_reference, n.display_name, n.platform,
    n.influence_score, n.community_id, n.updated_at
FROM graph_nodes n
WHERE n.id IN (
    SELECT source_node FROM graph_edges WHERE topic_id = $1
    UNION
    SELECT target_node FROM graph_edges WHERE topic_id = $1
)
ORDER BY n.influence_score DESC, n.id ASC
LIMIT $2;

-- name: ListGraphEdgesByTopic :many
SELECT id, topic_id, source_node, target_node, interaction_type, weight
FROM graph_edges
WHERE topic_id = $1
ORDER BY weight DESC
LIMIT $2;
