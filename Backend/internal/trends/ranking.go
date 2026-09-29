package trends

import (
	"math"
	"sort"
)

// Score combines growth, velocity and volume into a single trend score in
// [0, 100]. The weights are deliberately simple and explainable:
//
//	45% volume (log-scaled), 30% growth, 25% velocity.
func Score(growth, velocity, volume float64) float64 {
	score := 30*normalizeGrowth(growth) + 25*normalizeVelocity(velocity) + 45*normalizeVolume(volume)
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return math.Round(score*100) / 100
}

// Rank sorts points by score (desc) and truncates to topN (topN <= 0 keeps
// everything).
func Rank(points []Point, topN int) []Point {
	out := make([]Point, len(points))
	copy(out, points)
	sort.Slice(out, func(i, j int) bool {
		if out[i].TrendScore != out[j].TrendScore {
			return out[i].TrendScore > out[j].TrendScore
		}
		return out[i].Keyword < out[j].Keyword
	})
	if topN > 0 && len(out) > topN {
		out = out[:topN]
	}
	return out
}

func normalizeGrowth(g float64) float64 {
	if g <= 0 {
		return 0
	}
	return math.Min(1, g/2) // +200% saturates
}

func normalizeVelocity(v float64) float64 {
	if v <= 0 {
		return 0
	}
	return math.Min(1, v/10) // +10 posts/bucket saturates
}

func normalizeVolume(volume float64) float64 {
	if volume <= 0 {
		return 0
	}
	return math.Min(1, math.Log1p(volume)/math.Log1p(1000))
}
