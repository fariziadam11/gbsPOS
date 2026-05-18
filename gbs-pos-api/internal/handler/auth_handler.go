package handler

import (
	"net/http"
	"gbs-pos-api/internal/service"
	"gbs-pos-api/pkg/response"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	result, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error("INVALID_CREDENTIALS", "Username or password is incorrect"))
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
