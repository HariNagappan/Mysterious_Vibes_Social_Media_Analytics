package trends

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// Detect analyses a bucket series: growth of the recent half versus the
// baseline half, least-squares velocity, and a z-score spike test on the last
// bucket.
func Detect(buckets []Bucket) Detection {
	if len(buckets) == 0 {
		return Detection{}
	}

	split := len(buckets) / 2
	if split == 0 {
		split = 1
	}
	baseline := avgOf(buckets[:split])
	current := avgOf(buckets[split:])

	growth := 0.0
	if baseline > 0 {
		growth = (current - baseline) / baseline
	}
	velocity := slope(buckets)

	spike := false
	if len(buckets) >= 4 {
		counts := make([]float64, len(buckets))
		for i, b := range buckets {
			counts[i] = b.Count
		}
		mean, std := meanStd(counts)
		if std > 0 {
			z := (counts[len(counts)-1] - mean) / std
			spike = z >= 2
		}
	}

	return Detection{Current: current, Baseline: baseline, Growth: growth, Velocity: velocity, Spike: spike}
}

// ExtractKeywords returns the most frequent meaningful words across the given
// texts, filtering URLs, numbers, short tokens and stop words.
func ExtractKeywords(texts []string, topN int) []KeywordCount {
	counts := map[string]int{}
	for _, text := range texts {
		for _, word := range tokenize(text) {
			if len(word) < 3 || isNumeric(word) {
				continue
			}
			if _, stop := stopwords[word]; stop {
				continue
			}
			counts[word]++
		}
	}

	out := make([]KeywordCount, 0, len(counts))
	for word, n := range counts {
		out = append(out, KeywordCount{Keyword: word, Count: n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Keyword < out[j].Keyword
	})
	if topN > 0 && len(out) > topN {
		out = out[:topN]
	}
	return out
}

// Bucketize distributes the given timestamps into fixed-size buckets covering
// [start, start+bucketCount*bucketSize).
func Bucketize(timestamps []int64, start int64, bucketCount int, bucketSize int64) []Bucket {
	if bucketCount <= 0 || bucketSize <= 0 {
		return nil
	}
	series := make([]Bucket, bucketCount)
	for i := range series {
		series[i] = Bucket{Count: 0}
	}
	for _, ts := range timestamps {
		if ts < start {
			continue
		}
		idx := int((ts - start) / bucketSize)
		if idx >= bucketCount {
			idx = bucketCount - 1
		}
		if idx < 0 {
			continue
		}
		series[idx].Count++
	}
	return series
}

func avgOf(buckets []Bucket) float64 {
	if len(buckets) == 0 {
		return 0
	}
	var sum float64
	for _, b := range buckets {
		sum += b.Count
	}
	return sum / float64(len(buckets))
}

// slope is the least-squares slope of (index, count).
func slope(buckets []Bucket) float64 {
	n := len(buckets)
	if n < 2 {
		return 0
	}
	var sx, sy, sxy, sxx float64
	for i, b := range buckets {
		x := float64(i)
		y := b.Count
		sx += x
		sy += y
		sxy += x * y
		sxx += x * x
	}
	den := float64(n)*sxx - sx*sx
	if den == 0 {
		return 0
	}
	return (float64(n)*sxy - sx*sy) / den
}

func meanStd(values []float64) (float64, float64) {
	n := float64(len(values))
	if n == 0 {
		return 0, 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / n
	var sq float64
	for _, v := range values {
		sq += (v - mean) * (v - mean)
	}
	return mean, math.Sqrt(sq / n)
}

func tokenize(text string) []string {
	lower := strings.ToLower(text)
	fields := strings.FieldsFunc(lower, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if strings.HasPrefix(f, "http") || strings.HasPrefix(f, "www") {
			continue
		}
		out = append(out, f)
	}
	return out
}

func isNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// stopwords keeps trend extraction focused on topical terms.
var stopwords = map[string]struct{}{
	"the": {}, "and": {}, "for": {}, "with": {}, "that": {}, "this": {},
	"from": {}, "are": {}, "was": {}, "were": {}, "has": {}, "have": {},
	"had": {}, "you": {}, "your": {}, "our": {}, "their": {}, "its": {},
	"not": {}, "but": {}, "who": {}, "what": {}, "when": {}, "where": {},
	"will": {}, "would": {}, "could": {}, "should": {}, "about": {},
	"they": {}, "them": {}, "then": {}, "than": {}, "into": {}, "over": {},
	"under": {}, "here": {}, "there": {}, "been": {}, "being": {}, "just": {},
	"also": {}, "after": {}, "before": {}, "between": {}, "during": {},
	"https": {}, "http": {}, "www": {}, "com": {}, "amp": {}, "via": {},
	"can": {}, "may": {}, "still": {}, "now": {}, "new": {},
	// Hinglish
	"hai": {}, "hain": {}, "aur": {}, "ke": {}, "ki": {}, "ka": {}, "se": {},
	"ko": {}, "kya": {}, "nahi": {}, "mein": {}, "me": {}, "par": {},
	"yeh": {}, "wo": {}, "kar": {}, "koi": {}, "bhi": {}, "tha": {},
	"thi": {}, "raha": {}, "rahi": {}, "rahe": {}, "hoga": {}, "liye": {},
	"bahut": {}, "kuch": {}, "sab": {}, "ab": {}, "tak": {}, "de": {},
	"na": {}, "is": {}, "to": {}, "of": {}, "in": {}, "on": {}, "at": {},
}
