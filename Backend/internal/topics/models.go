package topics

import (
	"errors"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Domain errors surfaced to handlers.
var (
	// ErrNotFound is returned when a topic does not exist.
	ErrNotFound = errors.New("topics: not found")
	// ErrInvalidInput is returned for semantically invalid requests.
	ErrInvalidInput = errors.New("topics: invalid input")
)

// CreateRequest is the POST /topics payload.
type CreateRequest struct {
	Name        string   `json:"name" binding:"required,min=3,max=200"`
	Keywords    []string `json:"keywords" binding:"omitempty,max=50"`
	Description string   `json:"description" binding:"omitempty,max=1000"`
}

// ListResponse is the GET /topics payload.
type ListResponse struct {
	Topics []generated.Topic `json:"topics"`
	Total  int64             `json:"total"`
}
