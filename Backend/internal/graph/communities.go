package graph

import "sort"

// LabelPropagation finds communities with a weighted, deterministic label
// propagation pass: nodes iterate in id order and adopt the heaviest label
// among their neighbours. Returns a compressed community id (0..k-1) per node.
func LabelPropagation(nodeIDs []int64, edges []EdgeRef, iterations int) map[int64]int {
	result := make(map[int64]int, len(nodeIDs))
	if len(nodeIDs) == 0 {
		return result
	}
	if iterations <= 0 {
		iterations = 20
	}

	sortedIDs := make([]int64, len(nodeIDs))
	copy(sortedIDs, nodeIDs)
	sort.Slice(sortedIDs, func(i, j int) bool { return sortedIDs[i] < sortedIDs[j] })

	// Treat interactions as undirected weighted adjacency.
	neighbors := make(map[int64]map[int64]float64, len(sortedIDs))
	ensure := func(id int64) map[int64]float64 {
		if neighbors[id] == nil {
			neighbors[id] = map[int64]float64{}
		}
		return neighbors[id]
	}
	for _, id := range sortedIDs {
		ensure(id)
	}
	for _, e := range edges {
		if e.Source == e.Target {
			continue
		}
		w := e.Weight
		if w <= 0 {
			w = 1
		}
		ensure(e.Source)[e.Target] += w
		ensure(e.Target)[e.Source] += w
	}

	label := make(map[int64]int, len(sortedIDs))
	for _, id := range sortedIDs {
		label[id] = int(id)
	}

	for iter := 0; iter < iterations; iter++ {
		changed := false
		for _, id := range sortedIDs {
			if len(neighbors[id]) == 0 {
				continue
			}
			weightByLabel := map[int]float64{}
			for neighbor, w := range neighbors[id] {
				weightByLabel[label[neighbor]] += w
			}
			best := label[id]
			bestWeight := -1.0
			for lbl, w := range weightByLabel {
				if w > bestWeight || (w == bestWeight && lbl < best) {
					best = lbl
					bestWeight = w
				}
			}
			if best != label[id] {
				label[id] = best
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	// Compress labels into 0..k-1.
	seen := map[int]int{}
	next := 0
	for _, id := range sortedIDs {
		lbl := label[id]
		compressed, ok := seen[lbl]
		if !ok {
			compressed = next
			seen[lbl] = compressed
			next++
		}
		result[id] = compressed
	}
	return result
}
