package generated

import "time"

// User mirrors the users table.
type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// SocialSource mirrors the social_sources table.
type SocialSource struct {
	ID            int64
	Platform      string
	SourceName    string
	Configuration []byte
	Status        string
	CreatedAt     time.Time
}

// Topic mirrors the topics table.
type Topic struct {
	ID          int64
	Name        string
	Keywords    []string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SocialPost mirrors the social_posts table.
type SocialPost struct {
	ID                int64
	TopicID           int64
	Platform          string
	ExternalID        string
	AuthorReference   string
	AuthorDisplayName string
	Content           string
	Language          string
	RegionHint        string
	Timestamp         time.Time
	EngagementCount   int32
	ReplyToExternalID string
	MentionRefs       []string
	IngestedAt        time.Time
}

// SentimentResult mirrors the sentiment_results table.
type SentimentResult struct {
	ID             int64
	PostID         int64
	Emotion        string
	SentimentScore float64
	Confidence     float64
	ModelVersion   string
	CreatedAt      time.Time
}

// DemographicAnalytics mirrors the demographic_analytics table. JSONB
// columns are carried as raw bytes and decoded by the service layer.
type DemographicAnalytics struct {
	ID                   int64
	TopicID              int64
	SnapshotAt           time.Time
	LanguageDistribution []byte
	RegionDistribution   []byte
	InterestDistribution []byte
	AudienceGroups       []byte
}

// TrendAnalytic mirrors the trend_analytics table.
type TrendAnalytic struct {
	ID         int64
	TopicID    int64
	Keyword    string
	TrendScore float64
	GrowthRate float64
	Velocity   float64
	DetectedAt time.Time
}

// GraphNode mirrors the graph_nodes table.
type GraphNode struct {
	ID                    int64
	ExternalUserReference string
	DisplayName           string
	Platform              string
	InfluenceScore        float64
	CommunityID           int32
	UpdatedAt             time.Time
}

// GraphEdge mirrors the graph_edges table.
type GraphEdge struct {
	ID              int64
	TopicID         int64
	SourceNode      int64
	TargetNode      int64
	InteractionType string
	Weight          int32
}

// PostEmbedding mirrors the post_embeddings table. The vector column is
// selected as text and carries the canonical pgvector literal form.
type PostEmbedding struct {
	PostID       int64
	ModelVersion string
	Embedding    string
	CreatedAt    time.Time
}

// ── Aggregate / join result rows ────────────────────────────────────────────

// PostsBucket is one time bucket of the post-count growth series.
type PostsBucket struct {
	Bucket    time.Time
	PostCount int64
}

// EmotionCount is one row of the emotion distribution.
type EmotionCount struct {
	Emotion string
	Count   int64
}

// SentimentBucket is one time bucket of the sentiment timeline.
type SentimentBucket struct {
	Bucket    time.Time
	AvgScore  float64
	PostCount int64
	MinScore  float64
	MaxScore  float64
}

// TopicOverview aggregates headline statistics for a topic.
type TopicOverview struct {
	TotalPosts      int64
	TotalEngagement int64
	DistinctAuthors int64
	FirstPostAt     time.Time
	LastPostAt      time.Time
	AvgSentiment    float64
}

// PlatformCount is one row of the platform breakdown.
type PlatformCount struct {
	Platform string
	Count    int64
}

// TopAuthor is one row of the most-engaging authors ranking.
type TopAuthor struct {
	AuthorReference string
	PostCount       int64
	TotalEngagement int64
}

// DayCount is one day of the daily post volume series.
type DayCount struct {
	Day       time.Time
	PostCount int64
}

// RecentPostBrief is a compact post projection used for keyword extraction
// and important-event detection.
type RecentPostBrief struct {
	ID              int64
	Content         string
	Language        string
	Timestamp       time.Time
	EngagementCount int32
}

// PostSignal is the compact projection used by the demographics estimator.
type PostSignal struct {
	Language        string
	RegionHint      string
	Content         string
	AuthorReference string
	EngagementCount int32
	SentimentScore  float64
}

// SimilarPost is one row of the semantic similarity search.
type SimilarPost struct {
	PostID    int64
	Distance  float64
	Content   string
	Timestamp time.Time
}
