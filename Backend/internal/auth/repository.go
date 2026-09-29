package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// store abstracts persistence so the service layer is unit-testable with a
// fake, while production uses the PostgreSQL repository below.
type store interface {
	CreateUser(ctx context.Context, name, email, passwordHash, role string) (generated.User, error)
	GetUserByEmail(ctx context.Context, email string) (generated.User, error)
	GetUserByID(ctx context.Context, id int64) (generated.User, error)
}

// Repository implements store on PostgreSQL via the generated query layer.
type Repository struct {
	queries *generated.Queries
}

// NewRepository wraps the generated queries.
func NewRepository(queries *generated.Queries) *Repository {
	return &Repository{queries: queries}
}

// CreateUser inserts a new user row.
func (r *Repository) CreateUser(ctx context.Context, name, email, passwordHash, role string) (generated.User, error) {
	return r.queries.CreateUser(ctx, name, email, passwordHash, role)
}

// GetUserByEmail loads a user by email.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (generated.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

// GetUserByID loads a user by primary key.
func (r *Repository) GetUserByID(ctx context.Context, id int64) (generated.User, error) {
	return r.queries.GetUserByID(ctx, id)
}

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
