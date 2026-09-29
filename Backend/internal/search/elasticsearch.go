package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/config"
)

// esIndexer is a minimal Elasticsearch/OpenSearch REST client covering the
// document index and multi-match search operations PulseGraph needs.
type esIndexer struct {
	baseURL  string
	index    string
	username string
	password string
	http     *http.Client
	logger   *zap.Logger
}

func newESIndexer(cfg config.SearchConfig, logger *zap.Logger) *esIndexer {
	return &esIndexer{
		baseURL:  strings.TrimRight(cfg.URL, "/"),
		index:    cfg.Index,
		username: cfg.Username,
		password: cfg.Password,
		http:     &http.Client{Timeout: 10 * time.Second},
		logger:   logger,
	}
}

// Enabled reports whether a real backend is configured.
func (e *esIndexer) Enabled() bool { return true }

// Index stores one post document (idempotent per post id).
func (e *esIndexer) Index(ctx context.Context, doc Doc) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	endpoint := fmt.Sprintf("%s/%s/_doc/%d", e.baseURL, e.index, doc.ID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	e.authorize(req)

	resp, err := e.http.Do(req)
	if err != nil {
		return fmt.Errorf("search: index post %d: %w", doc.ID, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("search: index post %d: HTTP %d: %s", doc.ID, resp.StatusCode, string(payload))
	}
	return nil
}

// Search runs a multi-match query over content and author.
func (e *esIndexer) Search(ctx context.Context, query string, limit int) ([]Hit, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	payload := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  query,
				"fields": []string{"content^2", "author"},
			},
		},
		"size": limit,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("%s/%s/_search", e.baseURL, e.index)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	e.authorize(req)

	resp, err := e.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search: query: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("search: query: HTTP %d: %s", resp.StatusCode, string(payload))
	}

	var out struct {
		Hits struct {
			Hits []struct {
				Score  float64 `json:"_score"`
				Source Doc     `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("search: decode response: %w", err)
	}

	hits := make([]Hit, 0, len(out.Hits.Hits))
	for _, h := range out.Hits.Hits {
		hits = append(hits, Hit{
			ID:        h.Source.ID,
			Score:     h.Score,
			Content:   h.Source.Content,
			Author:    h.Source.Author,
			Timestamp: h.Source.Timestamp,
		})
	}
	return hits, nil
}

// Close releases resources (none for the HTTP client).
func (e *esIndexer) Close() error { return nil }

func (e *esIndexer) authorize(req *http.Request) {
	if e.username != "" {
		req.SetBasicAuth(e.username, e.password)
	}
}
