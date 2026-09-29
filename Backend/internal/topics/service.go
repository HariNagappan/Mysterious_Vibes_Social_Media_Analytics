package topics

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Service implements topic management use-cases.
type Service struct {
	repo *Repository
}

// NewService wires the topic service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create registers a monitored conversation after normalizing keywords.
func (s *Service) Create(ctx context.Context, req CreateRequest) (generated.Topic, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return generated.Topic{}, ErrInvalidInput
	}
	keywords := normalizeKeywords(req.Keywords)
	return s.repo.Create(ctx, generated.CreateTopicParams{
		Name:        name,
		Keywords:    keywords,
		Description: strings.TrimSpace(req.Description),
	})
}

// List returns a page of topics plus the total count.
func (s *Service) List(ctx context.Context, limit, offset int32) (ListResponse, error) {
	items, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return ListResponse{}, err
	}
	total, err := s.repo.Count(ctx)
	if err != nil {
		return ListResponse{}, err
	}
	return ListResponse{Topics: items, Total: total}, nil
}

// Get loads one topic, mapping missing rows onto ErrNotFound.
func (s *Service) Get(ctx context.Context, id int64) (generated.Topic, error) {
	topic, err := s.repo.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return generated.Topic{}, ErrNotFound
		}
		return generated.Topic{}, err
	}
	return topic, nil
}

// normalizeKeywords lowercases, trims and de-duplicates keywords.
func normalizeKeywords(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, k := range in {
		k = strings.ToLower(strings.TrimSpace(k))
		if k == "" {
			continue
		}
		if _, dup := seen[k]; dup {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
		if len(out) >= 50 {
			break
		}
	}
	return out
}

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
