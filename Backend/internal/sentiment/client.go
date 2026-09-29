package sentiment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Client talks to the Python sentiment-service over HTTP.
type Client struct {
	baseURL string
	http    *http.Client
	logger  *zap.Logger
}

// NewClient builds a sentiment-service client.
func NewClient(baseURL string, timeout time.Duration, logger *zap.Logger) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
		logger:  logger,
	}
}

type predictRequest struct {
	Text string `json:"text"`
}

type predictResponse struct {
	Emotion        string  `json:"emotion"`
	SentimentScore float64 `json:"sentiment_score"`
	Confidence     float64 `json:"confidence"`
	ModelVersion   string  `json:"model_version"`
}

// Predict asks the ML microservice to score a piece of text.
func (c *Client) Predict(ctx context.Context, text string) (Result, error) {
	body, err := json.Marshal(predictRequest{Text: text})
	if err != nil {
		return Result{}, fmt.Errorf("sentiment: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/predict", bytes.NewReader(body))
	if err != nil {
		return Result{}, fmt.Errorf("sentiment: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("sentiment: call ML service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("sentiment: ML service returned HTTP %d", resp.StatusCode)
	}

	var out predictResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Result{}, fmt.Errorf("sentiment: decode ML response: %w", err)
	}
	modelVersion := out.ModelVersion
	if modelVersion == "" {
		modelVersion = "sentiment-service"
	}
	return Result{
		Emotion:      out.Emotion,
		Score:        clamp(out.SentimentScore, -1, 1),
		Confidence:   clamp(out.Confidence, 0, 1),
		ModelVersion: modelVersion,
	}, nil
}
