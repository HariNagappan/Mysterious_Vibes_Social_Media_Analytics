// Package generated is a hand-maintained mirror of what `sqlc generate`
// produces for the queries in ../queries (the sqlc CLI is not available in
// every build environment). It is kept byte-for-byte compatible with the
// generated style: typed row structs + methods over a DBTX interface, so the
// package can be regenerated at any time with:
//
//	sqlc generate
//
// after which this header should be removed.
package generated

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX is the minimal database interface required by Queries. Both
// *pgxpool.Pool and pgx.Tx satisfy it.
type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// New builds a Queries bound to db.
func New(db DBTX) *Queries {
	return &Queries{db: db}
}

// Queries bundles every generated query method.
type Queries struct {
	db DBTX
}

// WithTx returns a copy of Queries bound to the provided transaction,
// allowing queries to participate in a caller-managed transaction.
func (q *Queries) WithTx(tx pgx.Tx) *Queries {
	return &Queries{db: tx}
}
