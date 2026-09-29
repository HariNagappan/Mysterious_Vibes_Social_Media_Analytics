package demographics

import "time"

// Snapshot is the aggregated audience picture for a topic. Only aggregate
// information is ever stored — never per-user records.
type Snapshot struct {
	TopicID              int64              `json:"topic_id"`
	SnapshotAt           time.Time          `json:"snapshot_at"`
	LanguageDistribution map[string]float64 `json:"language_distribution"`
	RegionDistribution   map[string]float64 `json:"region_distribution"`
	InterestDistribution map[string]float64 `json:"interest_distribution"`
	AudienceGroups       []AudienceGroup    `json:"audience_groups"`
	SampleSize           int                `json:"sample_size"`
}

// AudienceGroup is one coarse behavioural segment of the audience.
type AudienceGroup struct {
	Label       string  `json:"label"`
	Share       float64 `json:"share"`
	Description string  `json:"description"`
}
