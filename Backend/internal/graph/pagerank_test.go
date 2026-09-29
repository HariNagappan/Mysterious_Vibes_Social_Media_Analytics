package graph

import (
	"math"
	"testing"
)

func TestPageRankRanksLinkedNodeHighest(t *testing.T) {
	ids := []int64{1, 2, 3}
	edges := []EdgeRef{
		{Source: 1, Target: 3, Weight: 1},
		{Source: 2, Target: 3, Weight: 1},
	}
	ranks := PageRank(ids, edges, 0.85, 50)

	if !(ranks[3] > ranks[1] && ranks[3] > ranks[2]) {
		t.Fatalf("expected node 3 to rank highest: %+v", ranks)
	}
	var sum float64
	for _, v := range ranks {
		sum += v
	}
	if math.Abs(sum-1) > 0.05 {
		t.Fatalf("ranks should sum to ~1, got %.4f", sum)
	}
}

func TestPageRankEmpty(t *testing.T) {
	if got := PageRank(nil, nil, 0.85, 10); len(got) != 0 {
		t.Fatalf("expected empty ranks, got %+v", got)
	}
}
