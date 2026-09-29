// Package telegram implements the Telegram ingestion connector.
package telegram

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/ingestion"
)

// Platform identifies posts produced by this connector.
const Platform = "telegram"

// Client fetches messages from Telegram. In demo mode (no bot token) it
// synthesizes plausible channel chatter; with a token it polls the Bot API
// getUpdates endpoint and maps incoming messages.
type Client struct {
	botToken string
	http     *http.Client
	logger   *zap.Logger
}

// NewClient builds the connector.
func NewClient(botToken string, logger *zap.Logger) *Client {
	return &Client{
		botToken: strings.TrimSpace(botToken),
		http:     &http.Client{Timeout: 20 * time.Second},
		logger:   logger,
	}
}

// Name implements ingestion.Connector.
func (c *Client) Name() string { return "telegram-bot-updates" }

// Platform implements ingestion.Connector.
func (c *Client) Platform() string { return Platform }

// Fetch pulls messages for the given topic keywords.
func (c *Client) Fetch(ctx context.Context, opts ingestion.FetchOptions) ([]ingestion.RawPost, error) {
	if c.botToken == "" {
		return c.fetchDemo(opts), nil
	}
	return c.fetchLive(ctx, opts)
}

// fetchLive polls the Bot API for recent updates.
//
// Note: real-world channel monitoring normally uses a dedicated collector
// account; the Bot API only sees chats the bot belongs to. That trade-off is
// documented in docs/architecture.md.
func (c *Client) fetchLive(ctx context.Context, opts ingestion.FetchOptions) ([]ingestion.RawPost, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("timeout", "0")
	params.Set("allowed_updates", `["message","channel_post"]`)

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?%s", c.botToken, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("telegram: build request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("telegram: getUpdates: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("telegram: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram: getUpdates returned HTTP %d", resp.StatusCode)
	}

	messages, err := ParseUpdates(body)
	if err != nil {
		return nil, err
	}
	return ToRawPosts(messages), nil
}

// fetchDemo synthesizes a small batch of plausible channel messages.
func (c *Client) fetchDemo(opts ingestion.FetchOptions) []ingestion.RawPost {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	count := 4 + rng.Intn(6)
	now := time.Now().UTC()

	out := make([]ingestion.RawPost, 0, count)
	for i := 0; i < count; i++ {
		channel := demoChannels[rng.Intn(len(demoChannels))]
		template := demoTemplates[rng.Intn(len(demoTemplates))]
		keyword := "the situation"
		if len(opts.Keywords) > 0 {
			keyword = opts.Keywords[rng.Intn(len(opts.Keywords))]
		}
		content := strings.ReplaceAll(template, "{kw}", keyword)
		ts := now.Add(-time.Duration(rng.Intn(40)) * time.Minute)

		var mentions []string
		if rng.Intn(2) == 0 {
			other := demoChannels[rng.Intn(len(demoChannels))]
			if other != channel {
				mentions = []string{"tg:user:@" + other}
			}
		}

		out = append(out, ingestion.RawPost{
			Platform:        Platform,
			ExternalID:      fmt.Sprintf("tg-%d-%d", ts.UnixNano(), i),
			AuthorRef:       "tg:user:@" + channel,
			AuthorName:      channel,
			Content:         content,
			Timestamp:       ts,
			EngagementCount: 1 + rng.Intn(60),
			MentionRefs:     mentions,
		})
	}
	return out
}

var demoChannels = []string{
	"assam_relief_net", "city_events_hub", "metro_updates", "monsoon_watch",
	"product_deals_in", "techbharat_official",
}

var demoTemplates = []string{
	"Forwarded from a field contact: situation around {kw} is developing. Verify locally.",
	"Updates on {kw}: relief material being organised at the community centre.",
	"Please ignore the viral audio about {kw} — it is unverified.",
	"{kw}: traffic diversions in place, take alternate routes.",
	"Event volunteers for {kw} meetup check-in opens at 4 PM.",
	"Pre-orders for the new launch tied to {kw} are live now.",
	"Rumours about {kw} are circulating without sources. Check the pinned post.",
	"Someone with verified information on {kw}, please post here.",
}
