// Package twitter implements the X (Twitter) ingestion connector.
package twitter

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
const Platform = "twitter"

// Client fetches posts from X. In demo mode (no bearer token) it synthesizes
// realistic recent-search payloads so the full pipeline runs without API
// keys; with a token it calls the Twitter API v2 recent-search endpoint.
type Client struct {
	bearerToken string
	http        *http.Client
	logger      *zap.Logger
}

// NewClient builds the connector.
func NewClient(bearerToken string, logger *zap.Logger) *Client {
	return &Client{
		bearerToken: strings.TrimSpace(bearerToken),
		http:        &http.Client{Timeout: 20 * time.Second},
		logger:      logger,
	}
}

// Name implements ingestion.Connector.
func (c *Client) Name() string { return "twitter-recent-search" }

// Platform implements ingestion.Connector.
func (c *Client) Platform() string { return Platform }

// Fetch pulls posts for the given topic keywords.
func (c *Client) Fetch(ctx context.Context, opts ingestion.FetchOptions) ([]ingestion.RawPost, error) {
	if c.bearerToken == "" {
		return c.fetchDemo(opts), nil
	}
	return c.fetchLive(ctx, opts)
}

// fetchLive calls the recent-search endpoint and maps the response.
func (c *Client) fetchLive(ctx context.Context, opts ingestion.FetchOptions) ([]ingestion.RawPost, error) {
	query := strings.Join(opts.Keywords, " OR ")
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("twitter: fetch requires at least one keyword")
	}
	maxResults := opts.Limit
	if maxResults < 10 {
		maxResults = 10
	}
	if maxResults > 100 {
		maxResults = 100
	}

	params := url.Values{}
	params.Set("query", fmt.Sprintf("(%s) -is:retweet", query))
	params.Set("max_results", fmt.Sprintf("%d", maxResults))
	params.Set("tweet.fields", "created_at,public_metrics,referenced_tweets,entities,lang")
	params.Set("expansions", "author_id")
	params.Set("user.fields", "username,name")

	endpoint := "https://api.twitter.com/2/tweets/search/recent?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("twitter: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("twitter: recent search: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("twitter: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("twitter: recent search returned HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	tweets, err := ParseSearchResponse(body)
	if err != nil {
		return nil, err
	}
	return ToRawPosts(tweets), nil
}

// fetchDemo synthesizes a small batch of plausible posts around the topic
// keywords so the pipeline is fully demonstrable without credentials.
func (c *Client) fetchDemo(opts ingestion.FetchOptions) []ingestion.RawPost {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	count := 4 + rng.Intn(6)
	now := time.Now().UTC()

	out := make([]ingestion.RawPost, 0, count)
	for i := 0; i < count; i++ {
		author := demoAuthors[rng.Intn(len(demoAuthors))]
		template := demoTemplates[rng.Intn(len(demoTemplates))]
		keyword := "the topic"
		if len(opts.Keywords) > 0 {
			keyword = opts.Keywords[rng.Intn(len(opts.Keywords))]
		}
		content := strings.ReplaceAll(template, "{kw}", keyword)
		ts := now.Add(-time.Duration(rng.Intn(25)) * time.Minute)

		var mentions []string
		if rng.Intn(3) == 0 {
			other := demoAuthors[rng.Intn(len(demoAuthors))]
			if other != author {
				mentions = []string{"x:user:" + other}
			}
		}

		out = append(out, ingestion.RawPost{
			Platform:        Platform,
			ExternalID:      fmt.Sprintf("tw-%d-%d", ts.UnixNano(), i),
			AuthorRef:       "x:user:" + author,
			AuthorName:      capitalize(strings.ReplaceAll(author, "_", " ")),
			Content:         content,
			Timestamp:       ts,
			EngagementCount: 3 + rng.Intn(180),
			MentionRefs:     mentions,
		})
	}
	return out
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

var demoAuthors = []string{
	"floodwatch_assam", "newsbharat", "citizen_ravi", "metro_maya",
	"techwithravi", "civic_watchdog", "trafficpulse_city", "monsoon_alerts",
}

var demoTemplates = []string{
	"Reports about {kw} are spreading fast in local groups. Stay alert and verify before sharing.",
	"Anyone with updates on {kw}? Authorities say a statement is coming soon.",
	"Honestly impressed by how people are pulling together around {kw}. Community response is strong.",
	"This {kw} situation looks worse than the news is letting on. We need clearer coverage.",
	"Volunteers needed for {kw} support efforts. DM if you can help this weekend.",
	"So much misinformation around {kw}. Please check official sources before posting.",
	"Big update about {kw} today — the detail everyone missed matters.",
	"Panic around {kw} again? Citizens, stay calm and follow official advisories.",
}
