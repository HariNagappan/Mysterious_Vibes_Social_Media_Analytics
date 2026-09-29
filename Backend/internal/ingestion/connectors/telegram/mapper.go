package telegram

import (
	"fmt"

	"github.com/pulsegraph/pulsegraph-backend/internal/ingestion"
)

// ToRawPosts maps parsed messages onto the canonical connector format.
func ToRawPosts(messages []ParsedMessage) []ingestion.RawPost {
	out := make([]ingestion.RawPost, 0, len(messages))
	for _, m := range messages {
		authorRef := fmt.Sprintf("tg:user:%d", m.AuthorID)
		if m.AuthorID == 0 {
			authorRef = "tg:channel:" + m.ChatTitle
		}

		mentions := make([]string, 0, len(m.Mentions))
		for _, username := range m.Mentions {
			mentions = append(mentions, "tg:user:@"+username)
		}

		replyTo := ""
		if m.ReplyToID != 0 {
			replyTo = fmt.Sprintf("tg-%d-%d", m.ChatID, m.ReplyToID)
		}

		out = append(out, ingestion.RawPost{
			Platform:          Platform,
			ExternalID:        fmt.Sprintf("tg-%d-%d", m.ChatID, m.MessageID),
			AuthorRef:         authorRef,
			AuthorName:        m.AuthorName,
			Content:           m.Text,
			Timestamp:         m.Date,
			EngagementCount:   0,
			ReplyToExternalID: replyTo,
			MentionRefs:       mentions,
		})
	}
	return out
}
