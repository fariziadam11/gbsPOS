package handler

import (
	"gbs-cms-api/internal/service"
	"gbs-common/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler for authentication
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	User  UserInfo `json:"user"`
	Token string   `json:"token"`
}

// UserInfo represents user information
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

// Login godoc
//
//	@Summary		User login
//	@Description	Authenticate user and receive JWT token. Only available when ENABLE_DEMO_AUTH=true or Keycloak not configured.
//	@Tags			Authentication
//	@Accept		json
//	@Produce		json
//	@Param		request	body	LoginRequest	true	"Login credentials"
//	@Success		200	{object}	LoginResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		422	{object}	ErrorResponse
//	@Router			/v1/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusUnprocessableEntity,
			response.ValidationError("Invalid request body", nil),
		)
		return
	}
	result, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(
			http.StatusUnauthorized,
			response.Error("INVALID_CREDENTIALS", "Username or password is incorrect"),
		)
		return
	}
	c.JSON(http.StatusOK, response.Success(LoginResponse{
		User: UserInfo{
			ID:       result.User.ID,
			Username: result.User.Username,
			Name:     result.User.Name,
			Role:     result.User.Role,
		},
		Token: result.Token,
	}))
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
