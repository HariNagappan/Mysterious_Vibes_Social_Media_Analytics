package demographics

import (
	"testing"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

func TestEstimateDistributions(t *testing.T) {
	signals := []generated.PostSignal{
		{Language: "en", RegionHint: "Assam", Content: "flood rescue needed urgently in the district", EngagementCount: 600, SentimentScore: -0.5},
		{Language: "hi", RegionHint: "", Content: "बाढ़ relief madad जरूरी है", EngagementCount: 50, SentimentScore: 0.1},
		{Language: "en", RegionHint: "Assam", Content: "support the relief teams working overnight", EngagementCount: 150, SentimentScore: 0.4},
	}

	snap := Estimate(signals)
	if snap.SampleSize != 3 {
		t.Fatalf("sample size = %d", snap.SampleSize)
	}
	if snap.LanguageDistribution["en"] < 0.6 {
		t.Fatalf("expected english majority, got %+v", snap.LanguageDistribution)
	}
	if _, ok := snap.RegionDistribution["Assam"]; !ok {
		t.Fatalf("expected Assam region share, got %+v", snap.RegionDistribution)
	}
	if snap.InterestDistribution["Disaster response"] <= 0 {
		t.Fatalf("expected disaster interest, got %+v", snap.InterestDistribution)
	}

	var hasBroadcaster bool
	for _, group := range snap.AudienceGroups {
		if group.Label == "Broadcasters" && group.Share > 0 {
			hasBroadcaster = true
		}
	}
	if !hasBroadcaster {
		t.Fatalf("expected broadcaster audience group, got %+v", snap.AudienceGroups)
	}
}

func TestEstimateEmptyInput(t *testing.T) {
	snap := Estimate(nil)
	if len(snap.LanguageDistribution) != 0 || len(snap.AudienceGroups) != 0 {
		t.Fatalf("empty input should yield an empty snapshot: %+v", snap)
	}
}
