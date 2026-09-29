package twitter

import (
	"github.com/pulsegraph/pulsegraph-backend/internal/ingestion"
)

// ToRawPosts maps parsed tweets onto the canonical connector format. The
// external id prefix ("tw-") is kept consistent so reply chains can be
// resolved against stored posts.
func ToRawPosts(tweets []ParsedTweet) []ingestion.RawPost {
	out := make([]ingestion.RawPost, 0, len(tweets))
	for _, t := range tweets {
		authorRef := "x:user:" + t.Username
		if t.Username == "" {
			authorRef = "x:author:" + t.AuthorID
		}

		mentions := make([]string, 0, len(t.Mentions))
		for _, username := range t.Mentions {
			mentions = append(mentions, "x:user:"+username)
		}

		replyTo := ""
		if t.ReplyToID != "" {
			replyTo = "tw-" + t.ReplyToID
		}

		engagement := t.LikeCount + t.ReplyCount + t.RetweetCount*2

		out = append(out, ingestion.RawPost{
			Platform:          Platform,
			ExternalID:        "tw-" + t.ID,
			AuthorRef:         authorRef,
			AuthorName:        t.DisplayName,
			Content:           t.Text,
			Timestamp:         t.CreatedAt,
			EngagementCount:   engagement,
			ReplyToExternalID: replyTo,
			MentionRefs:       mentions,
		})
	}
	return out
}
