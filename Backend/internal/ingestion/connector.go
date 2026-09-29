// Package ingestion defines the plugin-based connector contract and the data
// normalizer that turns platform payloads into canonical PulseGraph posts.
package ingestion

import (
	"context"
	"sort"
	"time"
)

// RawPost is the connector-neutral wire format for a fetched post. Connectors
// keep platform specifics in Raw and fill the canonical fields they can.
type RawPost struct {
	Platform          string
	ExternalID        string
	AuthorRef         string
	AuthorName        string
	Content           string
	Language          string
	RegionHint        string
	Timestamp         time.Time
	EngagementCount   int
	ReplyToExternalID string
	MentionRefs       []string
	Raw               map[string]any
}

// FetchOptions parameterizes a connector pull.
type FetchOptions struct {
	TopicID  int64
	Keywords []string
	Since    time.Time
	Limit    int
}

// Connector is the plugin interface implemented by each platform connector.
// Supporting a new platform means implementing this interface and registering
// it — nothing else in the codebase changes.
type Connector interface {
	Name() string
	Platform() string
	Fetch(ctx context.Context, opts FetchOptions) ([]RawPost, error)
}

// Registry holds the available connectors keyed by platform.
type Registry struct {
	connectors map[string]Connector
}

// NewRegistry builds an empty connector registry.
func NewRegistry() *Registry {
	return &Registry{connectors: map[string]Connector{}}
}

// Register adds a connector (last registration for a platform wins).
func (r *Registry) Register(connector Connector) {
	r.connectors[connector.Platform()] = connector
}

// Get returns the connector for a platform.
func (r *Registry) Get(platform string) (Connector, bool) {
	connector, ok := r.connectors[platform]
	return connector, ok
}

// Platforms lists the registered platforms in deterministic order.
func (r *Registry) Platforms() []string {
	out := make([]string, 0, len(r.connectors))
	for platform := range r.connectors {
		out = append(out, platform)
	}
	sort.Strings(out)
	return out
}
