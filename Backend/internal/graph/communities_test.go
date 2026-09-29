package graph

import "testing"

func TestLabelPropagationFindsTwoCommunities(t *testing.T) {
	ids := []int64{1, 2, 3, 4}
	edges := []EdgeRef{
		{Source: 1, Target: 2, Weight: 2},
		{Source: 2, Target: 1, Weight: 2},
		{Source: 3, Target: 4, Weight: 2},
		{Source: 4, Target: 3, Weight: 2},
		{Source: 2, Target: 3, Weight: 0.1},
	}

	communities := LabelPropagation(ids, edges, 20)
	if communities[1] != communities[2] {
		t.Fatalf("nodes 1 and 2 should share a community: %+v", communities)
	}
	if communities[3] != communities[4] {
		t.Fatalf("nodes 3 and 4 should share a community: %+v", communities)
	}
	if communities[1] == communities[3] {
		t.Fatalf("weakly linked clusters should differ: %+v", communities)
	}
}
