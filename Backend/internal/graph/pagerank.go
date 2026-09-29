package graph

// PageRank runs the classic power iteration over the weighted interaction
// graph and returns scores keyed by node id. Weights represent interaction
// strength (reply/mention counts).
func PageRank(nodeIDs []int64, edges []EdgeRef, damping float64, iterations int) map[int64]float64 {
	n := len(nodeIDs)
	ranks := make(map[int64]float64, n)
	if n == 0 {
		return ranks
	}
	if damping <= 0 || damping >= 1 {
		damping = 0.85
	}
	if iterations <= 0 {
		iterations = 30
	}

	outWeight := make(map[int64]float64, n)
	for _, e := range edges {
		if e.Source == e.Target {
			continue
		}
		w := e.Weight
		if w <= 0 {
			w = 1
		}
		outWeight[e.Source] += w
	}

	initial := 1.0 / float64(n)
	for _, id := range nodeIDs {
		ranks[id] = initial
	}

	for iter := 0; iter < iterations; iter++ {
		// Dangling mass (nodes without out-edges) is redistributed uniformly;
		// it is computed from the previous iteration so ranks keep summing to 1.
		dangling := 0.0
		for _, id := range nodeIDs {
			if outWeight[id] <= 0 {
				dangling += ranks[id]
			}
		}
		baseRank := (1-damping)/float64(n) + damping*dangling/float64(n)

		next := make(map[int64]float64, n)
		for _, id := range nodeIDs {
			next[id] = baseRank
		}
		for _, e := range edges {
			if e.Source == e.Target {
				continue
			}
			ow := outWeight[e.Source]
			if ow <= 0 {
				continue
			}
			w := e.Weight
			if w <= 0 {
				w = 1
			}
			next[e.Target] += damping * ranks[e.Source] * (w / ow)
		}
		ranks = next
	}
	return ranks
}
