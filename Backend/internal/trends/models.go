package trends

import (
	"time"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Bucket is one time bucket of post volume.
type Bucket struct {
	Start time.Time
	Count float64
}

// Detection is the spike/growth analysis for one signal (a keyword or the
// topic as a whole).
type Detection struct {
	Current  float64 `json:"current_volume"`
	Baseline float64 `json:"baseline_volume"`
	Growth   float64 `json:"growth_rate"`
	Velocity float64 `json:"velocity"`
	Spike    bool    `json:"spike"`
}

// Point is one ranked trend measurement.
type Point struct {
	Keyword    string  `json:"keyword"`
	TrendScore float64 `json:"trend_score"`
	GrowthRate float64 `json:"growth_rate"`
	Velocity   float64 `json:"velocity"`
	IsOverall  bool    `json:"-"`
}

// KeywordCount is a raw keyword frequency.
type KeywordCount struct {
	Keyword string
	Count   int
}

// Window is the resolved analysis window of a trends response.
type Window struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// TopicTrends is the GET /topics/:id/trends payload.
type TopicTrends struct {
	TopicID        int64                     `json:"topic_id"`
	Window         Window                    `json:"window"`
	Overall        *Point                    `json:"overall,omitempty"`
	RisingKeywords []Point                   `json:"rising_keywords"`
	History        []generated.TrendAnalytic `json:"history"`
	DetectedAt     *time.Time                `json:"detected_at,omitempty"`
}

// OverallKeyword is the sentinel keyword used for the topic-level trend row.
const OverallKeyword = "overall"
