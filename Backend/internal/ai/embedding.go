package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// EmbeddingClient talks to the Python embedding-service.
type EmbeddingClient struct {
	baseURL string
	http    *http.Client
	logger  *zap.Logger
}

// NewEmbeddingClient builds an embedding-service client.
func NewEmbeddingClient(baseURL string, timeout time.Duration, logger *zap.Logger) *EmbeddingClient {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &EmbeddingClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
		logger:  logger,
	}
}

type embedRequest struct {
	Text string `json:"text"`
}

type embedResponse struct {
	Embedding    []float64 `json:"embedding"`
	ModelVersion string    `json:"model_version"`
	Dim          int       `json:"dim"`
}

// Embed returns the text embedding and the producing model version.
func (c *EmbeddingClient) Embed(ctx context.Context, text string) ([]float64, string, error) {
	body, err := json.Marshal(embedRequest{Text: text})
	if err != nil {
		return nil, "", fmt.Errorf("embedding: marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embed", bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("embedding: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("embedding: call service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("embedding: service returned HTTP %d", resp.StatusCode)
	}

	var out embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, "", fmt.Errorf("embedding: decode response: %w", err)
	}
	if len(out.Embedding) == 0 {
		return nil, "", fmt.Errorf("embedding: service returned an empty vector")
	}
	modelVersion := out.ModelVersion
	if modelVersion == "" {
		modelVersion = "embedding-service"
	}
	return out.Embedding, modelVersion, nil
}

// FormatVector renders a float vector as the canonical pgvector literal,
// e.g. "[0.013000,0.221000,...]".
func FormatVector(vec []float64) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, v := range vec {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(v, 'f', 6, 64))
	}
	b.WriteByte(']')
	return b.String()
}
