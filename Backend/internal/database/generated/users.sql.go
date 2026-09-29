package generated

import (
	"context"
)

const createUser = `INSERT INTO users (name, email, password_hash, role)
VALUES ($1, $2, $3, $4)
RETURNING id, name, email, password_hash, role, created_at, updated_at`

// CreateUser inserts a new dashboard user and returns the stored row.
func (q *Queries) CreateUser(ctx context.Context, name string, email string, passwordHash string, role string) (User, error) {
	row := q.db.QueryRow(ctx, createUser, name, email, passwordHash, role)
	var i User
	err := row.Scan(&i.ID, &i.Name, &i.Email, &i.PasswordHash, &i.Role, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

const getUserByEmail = `SELECT id, name, email, password_hash, role, created_at, updated_at
FROM users
WHERE email = $1`

// GetUserByEmail loads a user by unique email (login flow).
func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(&i.ID, &i.Name, &i.Email, &i.PasswordHash, &i.Role, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

const getUserByID = `SELECT id, name, email, password_hash, role, created_at, updated_at
FROM users
WHERE id = $1`

// GetUserByID loads a user by primary key.
func (q *Queries) GetUserByID(ctx context.Context, id int64) (User, error) {
	row := q.db.QueryRow(ctx, getUserByID, id)
	var i User
	err := row.Scan(&i.ID, &i.Name, &i.Email, &i.PasswordHash, &i.Role, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}
