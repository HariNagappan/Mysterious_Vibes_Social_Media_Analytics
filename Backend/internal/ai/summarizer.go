package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Summarizer produces the analytical report. It prefers the Python
// summarizer-service and falls back to a deterministic in-process template —
// both paths only explain metrics that were already computed.
type Summarizer struct {
	endpoint string
	http     *http.Client
	logger   *zap.Logger
}

// NewSummarizer builds the summarizer. An empty endpoint disables the ML call
// path and always uses the local template.
func NewSummarizer(endpoint string, timeout time.Duration, logger *zap.Logger) *Summarizer {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Summarizer{
		endpoint: strings.TrimRight(endpoint, "/"),
		http:     &http.Client{Timeout: timeout},
		logger:   logger,
	}
}

// Generate produces the report for the given pre-computed analytics.
func (s *Summarizer) Generate(ctx context.Context, in SummaryInput) *Report {
	source := in.Numbers()

	if s.endpoint != "" {
		report, err := s.callService(ctx, in)
		if err != nil {
			s.logger.Warn("ai: summarizer service unavailable; using template report", zap.Error(err))
		} else {
			report.Mode = "ml-service"
			if report.ModelVersion == "" {
				report.ModelVersion = "summarizer-service"
			}
			if report.GeneratedAt.IsZero() {
				report.GeneratedAt = time.Now().UTC()
			}
			if err := ValidateReport(report, source); err != nil {
				s.logger.Warn("ai: guardrail rejected service report; using template report", zap.Error(err))
			} else {
				return report
			}
		}
	}
	return BuildTemplateReport(in)
}

type summaryResponse struct {
	Title            string    `json:"title"`
	ExecutiveSummary string    `json:"executive_summary"`
	KeyFindings      []string  `json:"key_findings"`
	Sections         []Section `json:"sections"`
	ModelVersion     string    `json:"model_version"`
	GeneratedAt      time.Time `json:"generated_at"`
}

func (s *Summarizer) callService(ctx context.Context, in SummaryInput) (*Report, error) {
	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("ai: marshal summary input: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint+"/generate-summary", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ai: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ai: call summarizer service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ai: summarizer service returned HTTP %d", resp.StatusCode)
	}

	var out summaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("ai: decode summarizer response: %w", err)
	}
	return &Report{
		Title:            out.Title,
		ExecutiveSummary: out.ExecutiveSummary,
		KeyFindings:      out.KeyFindings,
		Sections:         out.Sections,
		GeneratedAt:      out.GeneratedAt,
		ModelVersion:     out.ModelVersion,
	}, nil
}

// BuildTemplateReport renders a deterministic narrative from computed metrics
// only. It is used when the ML service is unavailable and as the guardrail
// fallback for ungrounded service output.
func BuildTemplateReport(in SummaryInput) *Report {
	topEmotion := "neutral"
	topEmotionShare := 0.0
	for _, name := range in.TopEmotions {
		if share, ok := in.EmotionShares[name]; ok {
			topEmotion = name
			topEmotionShare = share
			break
		}
	}
	dominantLanguage, dominantShare := topShareOf(in.Languages)

	sentimentLabel := describeSentiment(in.AvgSentiment)
	growthLabel := describeGrowth(in.GrowthPercent)

	findings := make([]string, 0, 5)
	findings = append(findings, fmt.Sprintf(
		"Analysed %d posts from %d distinct accounts carrying %d total engagements.",
		in.TotalPosts, in.DistinctAuthors, in.TotalEngagement))
	findings = append(findings, fmt.Sprintf(
		"Overall sentiment is %s (%.2f mean score); %s dominates the emotional mix at %.0f%% share.",
		sentimentLabel, in.AvgSentiment, topEmotion, topEmotionShare*100))
	if len(in.TopKeywords) > 0 {
		findings = append(findings, fmt.Sprintf(
			"The fastest-rising keyword is %q — it anchors the current narrative cluster.", in.TopKeywords[0]))
	}
	if len(in.TopInfluencers) > 0 {
		findings = append(findings, fmt.Sprintf(
			"Influence concentrates around %q, with the graph holding %d nodes and %d edges.",
			in.TopInfluencers[0], in.NetworkNodes, in.NetworkEdges))
	}
	if dominantLanguage != "" {
		findings = append(findings, fmt.Sprintf(
			"The conversation is led by %s speakers (%.0f%% of analysed posts).",
			dominantLanguage, dominantShare*100))
	}

	sections := []Section{
		{
			Heading: "Sentiment landscape",
			Body: fmt.Sprintf(
				"Across the %d-hour window, average sentiment is %.2f (%s). The dominant emotion is %s, present in %.0f%% of scored posts.",
				in.WindowHours, in.AvgSentiment, sentimentLabel, topEmotion, topEmotionShare*100),
		},
		{
			Heading: "Discussion dynamics",
			Body: fmt.Sprintf(
				"Volume %s over the window. The discussion reached %d posts, %d engagements and %d distinct participants, with the interaction graph carrying %d edges across %d nodes.",
				growthLabel, in.TotalPosts, in.TotalEngagement, in.DistinctAuthors, in.NetworkEdges, in.NetworkNodes),
		},
		{
			Heading: "Narratives and keywords",
			Body:    keywordSection(in),
		},
		{
			Heading: "Audience composition",
			Body: fmt.Sprintf(
				"The largest linguistic community speaks %s (%.0f%%). Regional and interest patterns are available in the demographics section of the dashboard.",
				dominantLanguage, dominantShare*100),
		},
	}

	return &Report{
		Title:            "PulseGraph intelligence report — " + in.TopicName,
		ExecutiveSummary: fmt.Sprintf("PulseGraph analysed the %q conversation over the last %d hours. Sentiment is %s and discussion volume %s.", in.TopicName, in.WindowHours, sentimentLabel, growthLabel),
		KeyFindings:      findings,
		Sections:         sections,
		GeneratedAt:      time.Now().UTC(),
		ModelVersion:     "template-v1",
		Mode:             "template",
	}
}

func keywordSection(in SummaryInput) string {
	if len(in.TopKeywords) == 0 {
		return "No dominant keyword cluster was detected in the current window."
	}
	quoted := make([]string, 0, len(in.TopKeywords))
	for _, kw := range in.TopKeywords {
		quoted = append(quoted, fmt.Sprintf("%q", kw))
	}
	if len(quoted) > 5 {
		quoted = quoted[:5]
	}
	return fmt.Sprintf("Rising keywords in the conversation: %s. These terms mark the fastest-moving narrative clusters and are tracked for growth and velocity.", strings.Join(quoted, ", "))
}

func topShareOf(m map[string]float64) (string, float64) {
	bestKey := ""
	best := -1.0
	for key, value := range m {
		if value > best || (value == best && key < bestKey) {
			bestKey = key
			best = value
		}
	}
	if best < 0 {
		return "", 0
	}
	return bestKey, best
}

func describeSentiment(score float64) string {
	switch {
	case score >= 0.35:
		return "strongly positive"
	case score >= 0.1:
		return "leaning positive"
	case score > -0.1:
		return "mixed"
	case score > -0.35:
		return "leaning negative"
	default:
		return "strongly negative"
	}
}

func describeGrowth(percent float64) string {
	switch {
	case percent >= 50:
		return "rose sharply"
	case percent >= 10:
		return "trended upward"
	case percent > -10:
		return "held steady"
	default:
		return "declined"
	}
}
