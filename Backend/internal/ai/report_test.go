package ai

import (
	"testing"
	"time"
)

func sampleInput() SummaryInput {
	return SummaryInput{
		TopicName:       "Flood misinformation",
		GeneratedAt:     time.Now().UTC(),
		WindowHours:     168,
		TotalPosts:      1234,
		TotalEngagement: 5600,
		DistinctAuthors: 300,
		AvgSentiment:    -0.22,
		EmotionShares:   map[string]float64{"fear": 0.42, "anger": 0.21},
		TopEmotions:     []string{"fear", "anger"},
		TopKeywords:     []string{"flood", "relief"},
		TopInfluencers:  []string{"FloodWatch Assam"},
		Languages:       map[string]float64{"en": 0.6, "hi": 0.4},
		GrowthPercent:   35.5,
		NetworkNodes:    210,
		NetworkEdges:    380,
	}
}

func TestTemplateReportIsGrounded(t *testing.T) {
	in := sampleInput()
	report := BuildTemplateReport(in)
	if err := ValidateReport(report, in.Numbers()); err != nil {
		t.Fatalf("template report must only cite computed metrics: %v", err)
	}
	if report.Mode != "template" {
		t.Fatalf("expected template mode, got %s", report.Mode)
	}
	if len(report.KeyFindings) == 0 || len(report.Sections) == 0 {
		t.Fatal("expected findings and sections in the report")
	}
}

func TestGuardrailCatchesFabricatedNumbers(t *testing.T) {
	source := sampleInput().Numbers()
	text := "Sentiment improved by 87 percent across 987654 posts overnight."
	if got := UnexplainedNumbers(text, source); len(got) == 0 {
		t.Fatal("expected ungrounded numbers to be flagged")
	}
}

func TestGuardrailAllowsPercentForms(t *testing.T) {
	source := map[string]float64{"emotion_share_fear": 0.42}
	if got := UnexplainedNumbers("Fear accounts for 42% of scored posts.", source); len(got) != 0 {
		t.Fatalf("expected 42 to be grounded via share form, got %v", got)
	}
}

func TestGuardrailAllowsStructuralNumbers(t *testing.T) {
	source := map[string]float64{"total_posts": 10}
	if got := UnexplainedNumbers("Top 5 keywords for the 2026 window.", source); len(got) != 0 {
		t.Fatalf("small integers and years should be allowed, got %v", got)
	}
}
