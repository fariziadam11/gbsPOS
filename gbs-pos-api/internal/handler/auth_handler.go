package handler

import (
	"gbs-common/pkg/response"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login godoc
//
//	@Summary		User login
//	@Description	Authenticate user with username and password. Returns JWT token for API access.
//	@Tags			Authentication
//	@Accept		json
//	@Produce		json
//	@Param			request	body	dto.LoginRequest	true	"Login credentials"
//	@Success		200
//	@Failure		401
//	@Failure		422
//	@Router			/v1/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusUnprocessableEntity,
			response.ValidationError("Invalid request body", nil),
		)
		return
	}
	result, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		if err.Error() == "INVALID_CREDENTIALS" {
			c.JSON(
				http.StatusUnauthorized,
				response.Error("INVALID_CREDENTIALS", "Username or password is incorrect"),
			)
			return
		}
		c.JSON(
			http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", "Login service unavailable"),
		)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{
		"user": gin.H{
			"id":       result.User.ID,
			"username": result.User.Username,
			"name":     result.User.Name,
			"role":     result.User.Role,
		},
		"token": result.Token,
	}))
}
