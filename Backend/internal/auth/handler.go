package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/api/validators"
)

// Handler exposes the authentication endpoints.
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler wires the auth HTTP handler.
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// RegisterRoutes mounts /auth endpoints on the given router group.
func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/auth/register", h.Register)
	group.POST("/auth/login", h.Login)
}

// Register handles POST /api/v1/auth/register.
//
// @Summary      Register a dashboard user
// @Description  Creates a user account and returns a JWT.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body      RegisterRequest  true  "Registration payload"
// @Success      201      {object}  AuthResponse
// @Failure      400      {object}  validators.ErrorResponse
// @Failure      409      {object}  validators.ErrorResponse
// @Router       /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validators.RespondValidationError(c, err)
		return
	}
	res, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			validators.RespondError(c, http.StatusConflict, "email_taken", "an account with this email already exists")
			return
		}
		h.logger.Error("auth: register failed", zap.Error(err))
		validators.RespondError(c, http.StatusInternalServerError, "internal_error", "registration failed")
		return
	}
	c.JSON(http.StatusCreated, res)
}

// Login handles POST /api/v1/auth/login.
//
// @Summary      Log in
// @Description  Verifies credentials and returns a JWT.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body      LoginRequest  true  "Login payload"
// @Success      200      {object}  AuthResponse
// @Failure      400      {object}  validators.ErrorResponse
// @Failure      401      {object}  validators.ErrorResponse
// @Router       /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validators.RespondValidationError(c, err)
		return
	}
	res, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			validators.RespondError(c, http.StatusUnauthorized, "invalid_credentials", "email or password is incorrect")
			return
		}
		h.logger.Error("auth: login failed", zap.Error(err))
		validators.RespondError(c, http.StatusInternalServerError, "internal_error", "login failed")
		return
	}
	c.JSON(http.StatusOK, res)
}
