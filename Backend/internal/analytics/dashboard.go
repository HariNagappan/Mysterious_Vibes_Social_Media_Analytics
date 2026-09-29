package analytics

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/ai"
	"github.com/pulsegraph/pulsegraph-backend/internal/cache"
	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
	"github.com/pulsegraph/pulsegraph-backend/internal/demographics"
	"github.com/pulsegraph/pulsegraph-backend/internal/graph"
	"github.com/pulsegraph/pulsegraph-backend/internal/sentiment"
	"github.com/pulsegraph/pulsegraph-backend/internal/timeline"
	"github.com/pulsegraph/pulsegraph-backend/internal/topics"
	"github.com/pulsegraph/pulsegraph-backend/internal/trends"
)

// Domain errors surfaced to handlers.
var (
	// ErrTopicNotFound is returned when the topic id does not exist.
	ErrTopicNotFound = errors.New("analytics: topic not found")
)

const (
	dashboardWindowHours = 168 // 7 days
	networkNodeLimit     = 200
	networkEdgeLimit     = 400
)

// Dashboard is the GET /topics/:id/dashboard payload: complete intelligence
// for one monitored conversation.
type Dashboard struct {
	Topic        generated.Topic           `json:"topic"`
	GeneratedAt  time.Time                 `json:"generated_at"`
	Statistics   Snapshot                  `json:"statistics"`
	Timeline     *timeline.Timeline        `json:"timeline"`
	Sentiment    *sentiment.TopicSentiment `json:"sentiment"`
	Demographics demographics.Snapshot     `json:"demographics"`
	Trends       *trends.TopicTrends       `json:"trends"`
	Network      *graph.Network            `json:"network"`
	Summary      *ai.Report                `json:"summary"`
}

// Service assembles complete intelligence for the dashboard.
type Service struct {
	topics       *topics.Service
	timeline     *timeline.Service
	sentiment    *sentiment.Service
	demographics *demographics.Service
	trends       *trends.Service
	graph        *graph.Service
	summarizer   *ai.Summarizer
	aggregator   *Aggregator
	cache        *cache.Cache
	cacheTTL     time.Duration
	logger       *zap.Logger
}

// NewService wires the dashboard service.
func NewService(
	topicsSvc *topics.Service,
	timelineSvc *timeline.Service,
	sentimentSvc *sentiment.Service,
	demographicsSvc *demographics.Service,
	trendsSvc *trends.Service,
	graphSvc *graph.Service,
	summarizer *ai.Summarizer,
	aggregator *Aggregator,
	cacheClient *cache.Cache,
	cacheTTL time.Duration,
	logger *zap.Logger,
) *Service {
	return &Service{
		topics:       topicsSvc,
		timeline:     timelineSvc,
		sentiment:    sentimentSvc,
		demographics: demographicsSvc,
		trends:       trendsSvc,
		graph:        graphSvc,
		summarizer:   summarizer,
		aggregator:   aggregator,
		cache:        cacheClient,
		cacheTTL:     cacheTTL,
		logger:       logger,
	}
}

// Get assembles the full dashboard payload, served from Redis when warm.
func (s *Service) Get(ctx context.Context, topicID int64) (*Dashboard, error) {
	cacheKey := fmt.Sprintf("dashboard:%d", topicID)
	var cached Dashboard
	if found, err := s.cache.GetJSON(ctx, cacheKey, &cached); err == nil && found {
		return &cached, nil
	}

	topic, err := s.topics.Get(ctx, topicID)
	if err != nil {
		if errors.Is(err, topics.ErrNotFound) {
			return nil, ErrTopicNotFound
		}
		return nil, err
	}

	now := time.Now().UTC()
	since := now.Add(-dashboardWindowHours * time.Hour)

	timelineRes, err := s.timeline.Get(ctx, timeline.Query{
		TopicID: topicID,
		From:    since,
		HasFrom: true,
		To:      now,
		HasTo:   true,
		Limit:   50,
	})
	if err != nil {
		return nil, err
	}
	sentimentRes, err := s.sentiment.GetTopicSentiment(ctx, topicID, since, now)
	if err != nil {
		return nil, err
	}
	demographicsRes, err := s.demographics.Get(ctx, topicID)
	if err != nil {
		return nil, err
	}
	trendsRes, err := s.trends.Get(ctx, topicID)
	if err != nil {
		return nil, err
	}
	network, err := s.graph.Network(ctx, topicID, networkNodeLimit, networkEdgeLimit)
	if err != nil {
		return nil, err
	}
	stats, err := s.aggregator.Snapshot(ctx, topicID, 30)
	if err != nil {
		return nil, err
	}

	input := buildSummaryInput(topic, stats, sentimentRes, demographicsRes, trendsRes, network, now)
	summary := s.summarizer.Generate(ctx, input)

	dash := &Dashboard{
		Topic:        topic,
		GeneratedAt:  now,
		Statistics:   stats,
		Timeline:     timelineRes,
		Sentiment:    sentimentRes,
		Demographics: demographicsRes,
		Trends:       trendsRes,
		Network:      network,
		Summary:      summary,
	}
	if err := s.cache.SetJSON(ctx, cacheKey, dash, s.cacheTTL); err != nil {
		s.logger.Warn("analytics: dashboard cache write failed", zap.Error(err))
	}
	return dash, nil
}

// buildSummaryInput folds the computed sections into the deterministic
// payload the summarizer consumes (and the guardrail validates against).
func buildSummaryInput(
	topic generated.Topic,
	stats Snapshot,
	sent *sentiment.TopicSentiment,
	demo demographics.Snapshot,
	trendsRes *trends.TopicTrends,
	network *graph.Network,
	now time.Time,
) ai.SummaryInput {
	emotionShares := map[string]float64{}
	topEmotions := []string{}
	for _, item := range sent.Distribution {
		emotionShares[item.Emotion] = ai.Round(item.Share, 4)
		if len(topEmotions) < 3 {
			topEmotions = append(topEmotions, item.Emotion)
		}
	}

	topKeywords := []string{}
	for _, p := range trendsRes.RisingKeywords {
		if len(topKeywords) >= 5 {
			break
		}
		topKeywords = append(topKeywords, p.Keyword)
	}

	topInfluencers := []string{}
	for _, inf := range network.TopInfluencers {
		if len(topInfluencers) >= 5 {
			break
		}
		name := inf.DisplayName
		if name == "" {
			name = inf.Reference
		}
		topInfluencers = append(topInfluencers, name)
	}

	growthPercent := 0.0
	if trendsRes.Overall != nil {
		growthPercent = ai.Round(trendsRes.Overall.GrowthRate*100, 2)
	}

	return ai.SummaryInput{
		TopicName:       topic.Name,
		GeneratedAt:     now,
		WindowHours:     dashboardWindowHours,
		TotalPosts:      stats.Overview.TotalPosts,
		TotalEngagement: stats.Overview.TotalEngagement,
		DistinctAuthors: stats.Overview.DistinctAuthors,
		AvgSentiment:    ai.Round(sent.OverallScore, 4),
		EmotionShares:   emotionShares,
		TopEmotions:     topEmotions,
		TopKeywords:     topKeywords,
		TopInfluencers:  topInfluencers,
		Languages:       demo.LanguageDistribution,
		GrowthPercent:   growthPercent,
		NetworkNodes:    len(network.Nodes),
		NetworkEdges:    len(network.Edges),
	}
}
