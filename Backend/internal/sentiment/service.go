package sentiment

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Service implements the sentiment engine.
type Service struct {
	repo   *Repository
	client *Client
	logger *zap.Logger
}

// NewService wires the sentiment service.
func NewService(repo *Repository, client *Client, logger *zap.Logger) *Service {
	return &Service{repo: repo, client: client, logger: logger}
}

// Analyze scores text, preferring the Python ML service and falling back to
// the built-in lexicon when it is unreachable (graceful degradation keeps the
// pipeline flowing during demos).
func (s *Service) Analyze(ctx context.Context, text string) Result {
	if s.client != nil {
		res, err := s.client.Predict(ctx, text)
		if err == nil {
			return res
		}
		s.logger.Warn("sentiment: ML service unavailable, using lexicon fallback", zap.Error(err))
	}
	return AnalyzeLexicon(text)
}

// ProcessPost is invoked by the worker pipeline for EventPostIngested: it
// scores the post and persists the verdict.
func (s *Service) ProcessPost(ctx context.Context, postID int64) (Result, error) {
	post, err := s.repo.GetPost(ctx, postID)
	if err != nil {
		return Result{}, err
	}
	res := s.Analyze(ctx, post.Content)
	_, err = s.repo.UpsertResult(ctx, generated.UpsertSentimentResultParams{
		PostID:         postID,
		Emotion:        res.Emotion,
		SentimentScore: res.Score,
		Confidence:     res.Confidence,
		ModelVersion:   res.ModelVersion,
	})
	if err != nil {
		return Result{}, err
	}
	return res, nil
}

// GetTopicSentiment assembles the emotion distribution and sentiment timeline
// for the API.
func (s *Service) GetTopicSentiment(ctx context.Context, topicID int64, since, until time.Time) (*TopicSentiment, error) {
	distribution, err := s.repo.Distribution(ctx, topicID, since, until)
	if err != nil {
		return nil, err
	}
	timelineRows, err := s.repo.Timeline(ctx, topicID, since, until, sentimentBucketMinutes(since, until))
	if err != nil {
		return nil, err
	}
	overall, err := s.repo.Average(ctx, topicID, since, until)
	if err != nil {
		return nil, err
	}
	analyzed, err := s.repo.CountByTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}
	total, err := s.repo.CountPostsByTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}

	items := make([]DistributionItem, 0, len(distribution))
	for _, d := range distribution {
		share := 0.0
		if analyzed > 0 {
			share = float64(d.Count) / float64(analyzed)
		}
		items = append(items, DistributionItem{Emotion: d.Emotion, Count: d.Count, Share: share})
	}

	points := make([]TimelinePoint, 0, len(timelineRows))
	for _, t := range timelineRows {
		points = append(points, TimelinePoint{
			Bucket:    t.Bucket,
			AvgScore:  t.AvgScore,
			PostCount: t.PostCount,
			MinScore:  t.MinScore,
			MaxScore:  t.MaxScore,
		})
	}

	return &TopicSentiment{
		TopicID:       topicID,
		Window:        Window{From: since, To: until},
		OverallScore:  overall,
		AnalyzedPosts: analyzed,
		TotalPosts:    total,
		Distribution:  items,
		Timeline:      points,
	}, nil
}

func sentimentBucketMinutes(from, to time.Time) int32 {
	span := to.Sub(from)
	switch {
	case span <= 48*time.Hour:
		return 60
	case span <= 10*24*time.Hour:
		return 720
	default:
		return 1440
	}
}
