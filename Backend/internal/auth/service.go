package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Service implements registration and login.
type Service struct {
	store  store
	tokens *Manager
	logger *zap.Logger
}

// NewService wires the auth service.
func NewService(store store, tokens *Manager, logger *zap.Logger) *Service {
	return &Service{store: store, tokens: tokens, logger: logger}
}

// Register creates a user and immediately issues a token.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	role := req.Role
	if role == "" {
		role = "analyst"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("auth: hash password: %w", err)
	}
	user, err := s.store.CreateUser(ctx, name, email, string(hash), role)
	if err != nil {
		if isUniqueViolation(err) {
			return AuthResponse{}, ErrEmailTaken
		}
		return AuthResponse{}, fmt.Errorf("auth: create user: %w", err)
	}
	return s.issue(user)
}

// Login verifies credentials and issues a token.
func (s *Service) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if isNotFound(err) {
			return AuthResponse{}, ErrInvalidCredentials
		}
		return AuthResponse{}, fmt.Errorf("auth: load user: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}
	return s.issue(user)
}

func (s *Service) issue(user generated.User) (AuthResponse, error) {
	principal := User{ID: user.ID, Name: user.Name, Email: user.Email, Role: user.Role}
	token, expiresAt, err := s.tokens.GenerateToken(principal)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("auth: issue token: %w", err)
	}
	return AuthResponse{Token: token, ExpiresAt: expiresAt, User: principal}, nil
}

// ErrUnauthorized is a sentinel used by tests.
var ErrUnauthorized = errors.New("auth: unauthorized")
