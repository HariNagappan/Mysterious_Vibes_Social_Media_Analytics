package telegram

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

var mentionRe = regexp.MustCompile(`@([A-Za-z0-9_]{3,32})`)

// ParsedMessage is the subset of Telegram update payloads the connector
// consumes (both private messages and channel posts).
type ParsedMessage struct {
	MessageID  int64
	ChatID     int64
	ChatTitle  string
	AuthorID   int64
	AuthorName string
	Text       string
	Date       time.Time
	ReplyToID  int64
	Mentions   []string
}

type tgUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type tgChat struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type tgReply struct {
	MessageID int64 `json:"message_id"`
}

type tgMessage struct {
	MessageID      int64    `json:"message_id"`
	Date           int64    `json:"date"`
	Text           string   `json:"text"`
	From           *tgUser  `json:"from"`
	Chat           tgChat   `json:"chat"`
	ReplyToMessage *tgReply `json:"reply_to_message"`
}

type tgUpdate struct {
	UpdateID    int64      `json:"update_id"`
	Message     *tgMessage `json:"message"`
	ChannelPost *tgMessage `json:"channel_post"`
}

// ParseUpdates decodes a getUpdates response into ParsedMessage values.
func ParseUpdates(data []byte) ([]ParsedMessage, error) {
	var payload struct {
		OK     bool       `json:"ok"`
		Result []tgUpdate `json:"result"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("telegram: decode updates: %w", err)
	}
	if !payload.OK {
		return nil, fmt.Errorf("telegram: API returned ok=false")
	}

	out := make([]ParsedMessage, 0, len(payload.Result))
	for _, update := range payload.Result {
		msg := update.Message
		if msg == nil {
			msg = update.ChannelPost
		}
		if msg == nil || msg.Text == "" {
			continue
		}

		var authorID int64
		authorName := msg.Chat.Title
		if msg.From != nil {
			authorID = msg.From.ID
			authorName = msg.From.FirstName
			if msg.From.Username != "" {
				authorName = msg.From.Username
			}
		}

		var replyTo int64
		if msg.ReplyToMessage != nil {
			replyTo = msg.ReplyToMessage.MessageID
		}

		out = append(out, ParsedMessage{
			MessageID:  msg.MessageID,
			ChatID:     msg.Chat.ID,
			ChatTitle:  msg.Chat.Title,
			AuthorID:   authorID,
			AuthorName: authorName,
			Text:       msg.Text,
			Date:       time.Unix(msg.Date, 0).UTC(),
			ReplyToID:  replyTo,
			Mentions:   extractMentions(msg.Text),
		})
	}
	return out, nil
}

func extractMentions(text string) []string {
	matches := mentionRe.FindAllStringSubmatch(text, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) == 2 {
			out = append(out, m[1])
		}
	}
	return out
}
