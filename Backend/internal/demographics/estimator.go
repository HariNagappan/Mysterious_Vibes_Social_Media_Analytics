package demographics

import (
	"math"
	"strings"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Estimate builds an aggregated audience snapshot from post signals.
//
// The estimator is deliberately explainable: distributions come from explicit
// signals (language codes, region hints, topic lexicons) and behavioural
// segments from engagement tiers plus sentiment lean. It never touches
// personal data — only aggregate shares leave this function.
func Estimate(signals []generated.PostSignal) Snapshot {
	empty := Snapshot{
		LanguageDistribution: map[string]float64{},
		RegionDistribution:   map[string]float64{},
		InterestDistribution: map[string]float64{},
		AudienceGroups:       []AudienceGroup{},
	}

	total := len(signals)
	if total == 0 {
		return empty
	}

	language := map[string]int{}
	region := map[string]int{}
	interest := map[string]int{}
	var broadcasters, amplifiers, participants int
	var advocates, critics int

	for _, sg := range signals {
		lang := sg.Language
		if lang == "" {
			lang = "und"
		}
		language[lang]++

		reg := sg.RegionHint
		if reg == "" {
			reg = inferRegion(sg.Language)
		}
		region[reg]++

		for _, label := range inferInterests(sg.Content) {
			interest[label]++
		}

		switch {
		case sg.EngagementCount >= 500:
			broadcasters++
		case sg.EngagementCount >= 100:
			amplifiers++
		default:
			participants++
		}

		switch {
		case sg.SentimentScore >= 0.2:
			advocates++
		case sg.SentimentScore <= -0.2:
			critics++
		}
	}

	toShares := func(m map[string]int) map[string]float64 {
		out := make(map[string]float64, len(m))
		for k, v := range m {
			out[k] = round3(float64(v) / float64(total))
		}
		return out
	}

	groups := []AudienceGroup{
		{
			Label:       "Broadcasters",
			Share:       round3(float64(broadcasters) / float64(total)),
			Description: "High-reach accounts (500+ engagements) shaping the dominant narrative.",
		},
		{
			Label:       "Amplifiers",
			Share:       round3(float64(amplifiers) / float64(total)),
			Description: "Mid-reach accounts (100-499 engagements) spreading content between communities.",
		},
		{
			Label:       "Participants",
			Share:       round3(float64(participants) / float64(total)),
			Description: "Grass-roots accounts (<100 engagements) replying, quoting and reacting.",
		},
	}
	if share := float64(critics) / float64(total); share >= 0.1 {
		groups = append(groups, AudienceGroup{
			Label:       "Critical voices",
			Share:       round3(share),
			Description: "Accounts leaning negative — the early-warning segment for misinformation spread.",
		})
	}
	if share := float64(advocates) / float64(total); share >= 0.1 {
		groups = append(groups, AudienceGroup{
			Label:       "Advocates",
			Share:       round3(share),
			Description: "Accounts leaning positive — counter-narrative and support mobilisation.",
		})
	}

	return Snapshot{
		LanguageDistribution: toShares(language),
		RegionDistribution:   toShares(region),
		InterestDistribution: toShares(interest),
		AudienceGroups:       groups,
		SampleSize:           total,
	}
}

// inferRegion maps a language onto a coarse regional bucket when a post does
// not carry an explicit region hint.
func inferRegion(lang string) string {
	switch lang {
	case "hi":
		return "India (Hindi belt)"
	case "bn":
		return "India (East)"
	case "ta", "te":
		return "India (South)"
	case "en":
		return "English-speaking (global)"
	default:
		return "Unknown"
	}
}

type interestRule struct {
	name     string
	keywords []string
}

// interestRules is an ordered lexicon so estimation output is deterministic.
var interestRules = []interestRule{
	{"Disaster response", []string{"flood", "rescue", "relief", "evacuation", "emergency", "landslide", "monsoon", "shelter", "stranded", "damage"}},
	{"Public safety", []string{"police", "safety", "alert", "warning", "accident", "fire", "traffic", "crime", "danger"}},
	{"Politics & policy", []string{"election", "minister", "government", "parliament", "vote", "campaign", "policy", "opposition", "bill"}},
	{"Technology", []string{"launch", "ai", "chip", "device", "app", "upgrade", "features", "battery", "software", "camera"}},
	{"Sports", []string{"match", "tournament", "marathon", "score", "team", "cricket", "football", "medal", "coach"}},
	{"Health", []string{"hospital", "health", "vaccine", "disease", "medical", "doctor", "outbreak", "symptom"}},
	{"Environment", []string{"climate", "pollution", "weather", "river", "forest", "heatwave", "emissions"}},
	{"Business", []string{"price", "market", "stock", "brand", "product", "startup", "revenue", "sale", "discount"}},
	{"Media & virality", []string{"viral", "rumour", "rumor", "misinformation", "fake", "breaking", "shared", "trending", "screenshot"}},
}

// inferInterests returns up to three interest labels detected in the content.
func inferInterests(content string) []string {
	lower := strings.ToLower(content)
	out := make([]string, 0, 3)
	for _, rule := range interestRules {
		for _, kw := range rule.keywords {
			if strings.Contains(lower, kw) {
				out = append(out, rule.name)
				break
			}
		}
		if len(out) >= 3 {
			break
		}
	}
	return out
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}
