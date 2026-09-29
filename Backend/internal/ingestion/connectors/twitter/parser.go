package twitter

import (
	"encoding/json"
	"fmt"
	"time"
)

// ParsedTweet is the subset of the Twitter API v2 recent-search payload the
// connector consumes.
type ParsedTweet struct {
	ID           string
	Text         string
	AuthorID     string
	Username     string
	DisplayName  string
	CreatedAt    time.Time
	LikeCount    int
	ReplyCount   int
	RetweetCount int
	ReplyToID    string
	Mentions     []string
}

// ParseSearchResponse decodes a Twitter API v2 recent-search response into
// ParsedTweet values.
func ParseSearchResponse(data []byte) ([]ParsedTweet, error) {
	var payload struct {
		Data []struct {
			ID               string `json:"id"`
			Text             string `json:"text"`
			AuthorID         string `json:"author_id"`
			CreatedAt        string `json:"created_at"`
			ReferencedTweets []struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			} `json:"referenced_tweets"`
			PublicMetrics struct {
				LikeCount    int `json:"like_count"`
				ReplyCount   int `json:"reply_count"`
				RetweetCount int `json:"retweet_count"`
			} `json:"public_metrics"`
			Entities struct {
				Mentions []struct {
					Username string `json:"username"`
				} `json:"mentions"`
			} `json:"entities"`
		} `json:"data"`
		Includes struct {
			Users []struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Username string `json:"username"`
			} `json:"users"`
		} `json:"includes"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("twitter: decode search response: %w", err)
	}

	type twitterUser struct {
		name     string
		username string
	}
	userByID := map[string]twitterUser{}
	for _, u := range payload.Includes.Users {
		userByID[u.ID] = twitterUser{name: u.Name, username: u.Username}
	}

	out := make([]ParsedTweet, 0, len(payload.Data))
	for _, t := range payload.Data {
		createdAt := time.Now().UTC()
		if t.CreatedAt != "" {
			if parsed, err := time.Parse(time.RFC3339, t.CreatedAt); err == nil {
				createdAt = parsed
			}
		}

		replyTo := ""
		for _, ref := range t.ReferencedTweets {
			if ref.Type == "replied_to" {
				replyTo = ref.ID
				break
			}
		}

		mentions := make([]string, 0, len(t.Entities.Mentions))
		for _, m := range t.Entities.Mentions {
			if m.Username != "" {
				mentions = append(mentions, m.Username)
			}
		}

		user := userByID[t.AuthorID]
		out = append(out, ParsedTweet{
			ID:           t.ID,
			Text:         t.Text,
			AuthorID:     t.AuthorID,
			Username:     user.username,
			DisplayName:  user.name,
			CreatedAt:    createdAt,
			LikeCount:    t.PublicMetrics.LikeCount,
			ReplyCount:   t.PublicMetrics.ReplyCount,
			RetweetCount: t.PublicMetrics.RetweetCount,
			ReplyToID:    replyTo,
			Mentions:     mentions,
		})
	}
	return out, nil
}
