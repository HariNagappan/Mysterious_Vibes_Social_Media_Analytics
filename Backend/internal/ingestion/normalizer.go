package ingestion

import (
	"fmt"
	"strings"
	"time"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
	"github.com/pulsegraph/pulsegraph-backend/internal/pipeline"
)

const maxMentionsPerPost = 20

// NormalizedPost is a validated, canonical post ready for persistence.
type NormalizedPost struct {
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
}

// Normalize validates and canonicalizes a raw connector post: required fields
// are enforced, timestamps default to now, engagement is clamped and the
// language is detected when the connector could not supply it.
func Normalize(raw RawPost, topicID int64) (NormalizedPost, error) {
	platform := strings.ToLower(strings.TrimSpace(raw.Platform))
	if platform == "" {
		return NormalizedPost{}, fmt.Errorf("ingestion: post is missing the platform")
	}
	externalID := strings.TrimSpace(raw.ExternalID)
	if externalID == "" {
		return NormalizedPost{}, fmt.Errorf("ingestion: post is missing the external id")
	}
	authorRef := strings.TrimSpace(raw.AuthorRef)
	if authorRef == "" {
		return NormalizedPost{}, fmt.Errorf("ingestion: post is missing the author reference")
	}
	content := strings.TrimSpace(raw.Content)
	if content == "" {
		return NormalizedPost{}, fmt.Errorf("ingestion: post has empty content")
	}

	language := strings.TrimSpace(raw.Language)
	if language == "" {
		language, _ = pipeline.DetectLanguage(content)
	}

	timestamp := raw.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}

	engagement := raw.EngagementCount
	if engagement < 0 {
		engagement = 0
	}

	return NormalizedPost{
		TopicID:           topicID,
		Platform:          platform,
		ExternalID:        externalID,
		AuthorReference:   authorRef,
		AuthorDisplayName: strings.TrimSpace(raw.AuthorName),
		Content:           content,
		Language:          language,
		RegionHint:        strings.TrimSpace(raw.RegionHint),
		Timestamp:         timestamp,
		EngagementCount:   int32(engagement),
		ReplyToExternalID: strings.TrimSpace(raw.ReplyToExternalID),
		MentionRefs:       normalizeMentions(raw.MentionRefs, authorRef),
	}, nil
}

func normalizeMentions(refs []string, self string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "" || ref == self {
			continue
		}
		if _, dup := seen[ref]; dup {
			continue
		}
		seen[ref] = struct{}{}
		out = append(out, ref)
		if len(out) >= maxMentionsPerPost {
			break
		}
	}
	return out
}

// ToInsertParams converts the normalized post to the persistence-layer params.
func (n NormalizedPost) ToInsertParams() generated.InsertPostParams {
	return generated.InsertPostParams{
		TopicID:           n.TopicID,
		Platform:          n.Platform,
		ExternalID:        n.ExternalID,
		AuthorReference:   n.AuthorReference,
		AuthorDisplayName: n.AuthorDisplayName,
		Content:           n.Content,
		Language:          n.Language,
		RegionHint:        n.RegionHint,
		Timestamp:         n.Timestamp,
		EngagementCount:   n.EngagementCount,
		ReplyToExternalID: n.ReplyToExternalID,
		MentionRefs:       n.MentionRefs,
	}
}
