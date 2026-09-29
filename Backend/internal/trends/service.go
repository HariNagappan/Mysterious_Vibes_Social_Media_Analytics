package trends

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

const (
	// WindowHours is the analysis window for trend computation.
	WindowHours = 48
	// keywordLimit is how many rising keywords are tracked.
	keywordLimit = 10
)

// Service implements trend detection over collected posts.
type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService wires the trends service.
func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// ComputeAndStore recomputes the topic trend plus per-keyword trends and
// persists one snapshot batch. Returns the number of stored rows.
func (s *Service) ComputeAndStore(ctx context.Context, topicID int64) (int, error) {
	now := time.Now().UTC()
	start := now.Add(-WindowHours * time.Hour)

	bucketRows, err := s.repo.Buckets(ctx, topicID, start, now, 60)
	if err != nil {
		return 0, err
	}
	series := make([]Bucket, 0, len(bucketRows))
	var totalVolume float64
	for _, b := range bucketRows {
		series = append(series, Bucket{Start: b.Bucket, Count: float64(b.PostCount)})
		totalVolume += float64(b.PostCount)
	}
	overallDetection := Detect(series)
	overallScore := Score(overallDetection.Growth, overallDetection.Velocity, totalVolume)

	stored := 0
	if _, err := s.repo.Insert(ctx, generated.InsertTrendSnapshotParams{
		TopicID:    topicID,
		Keyword:    OverallKeyword,
		TrendScore: overallScore,
		GrowthRate: overallDetection.Growth,
		Velocity:   overallDetection.Velocity,
	}); err != nil {
		return stored, err
	}
	stored++

	posts, err := s.repo.RecentPosts(ctx, topicID, 300)
	if err != nil {
		return stored, err
	}
	texts := make([]string, 0, len(posts))
	for _, p := range posts {
		texts = append(texts, p.Content)
	}

	bucketCount := WindowHours // one bucket per hour
	startUnix := start.Unix()
	bucketSeconds := int64(3600)

	keywords := ExtractKeywords(texts, keywordLimit)
	for _, kw := range keywords {
		var timestamps []int64
		for _, p := range posts {
			if strings.Contains(strings.ToLower(p.Content), kw.Keyword) {
				timestamps = append(timestamps, p.Timestamp.Unix())
			}
		}
		series := Bucketize(timestamps, startUnix, bucketCount, bucketSeconds)
		det := Detect(series)
		score := Score(det.Growth, det.Velocity, float64(kw.Count))
		if _, err := s.repo.Insert(ctx, generated.InsertTrendSnapshotParams{
			TopicID:    topicID,
			Keyword:    kw.Keyword,
			TrendScore: score,
			GrowthRate: det.Growth,
			Velocity:   det.Velocity,
		}); err != nil {
			return stored, err
		}
		stored++
	}

	s.logger.Debug("trends: snapshot stored", zap.Int64("topic_id", topicID), zap.Int("rows", stored))
	return stored, nil
}

// Get returns the latest computed trends for the API.
func (s *Service) Get(ctx context.Context, topicID int64) (*TopicTrends, error) {
	now := time.Now().UTC()
	res := &TopicTrends{
		TopicID:        topicID,
		Window:         Window{From: now.Add(-WindowHours * time.Hour), To: now},
		RisingKeywords: []Point{},
		History:        []generated.TrendAnalytic{},
	}

	rows, err := s.repo.Latest(ctx, topicID, 60)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		point := Point{
			Keyword:    row.Keyword,
			TrendScore: row.TrendScore,
			GrowthRate: row.GrowthRate,
			Velocity:   row.Velocity,
		}
		detected := row.DetectedAt
		if row.Keyword == OverallKeyword {
			overall := point
			res.Overall = &overall
			if res.DetectedAt == nil || detected.After(*res.DetectedAt) {
				res.DetectedAt = &detected
			}
			continue
		}
		res.RisingKeywords = append(res.RisingKeywords, point)
		if res.DetectedAt == nil || detected.After(*res.DetectedAt) {
			res.DetectedAt = &detected
		}
	}

	history, err := s.repo.History(ctx, topicID, 100)
	if err != nil {
		return nil, err
	}
	res.History = history
	res.RisingKeywords = Rank(res.RisingKeywords, keywordLimit)
	return res, nil
}
