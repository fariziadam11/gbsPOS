package handler

import (
	"errors"
	"gbs-cms-api/internal/dto"
	"gbs-cms-api/internal/service"
	"gbs-common/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// List godoc
//
//	@Summary		List all users
//	@Description	Get list of all users. Only ADMIN role.
//	@Tags			Users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200
//	@Failure		401
//	@Failure		403
//	@Router			/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	users, err := h.userService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(users))
}

// Get godoc
//
//	@Summary		Get user by ID
//	@Description	Get a single user by ID. Only ADMIN role.
//	@Tags			Users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"User ID"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		404
//	@Router			/v1/users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid user ID"))
		return
	}
	user, err := h.userService.Get(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.Error("USER_NOT_FOUND", "User not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(user))
}

// Create godoc
//
//	@Summary		Create new user
//	@Description	Create a new user account. Only ADMIN role.
//	@Tags			Users
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	dto.CreateUserRequest	true	"User data"
//	@Success		201
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		422
//	@Router			/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	user, err := h.userService.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success(user))
}

// Update godoc
//
//	@Summary		Update user
//	@Description	Update an existing user. Only ADMIN role.
//	@Tags			Users
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int					true	"User ID"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		404
//	@Failure		422
//	@Router			/v1/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid user ID"))
		return
	}
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	user, err := h.userService.Update(uint(id), req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.Error("USER_NOT_FOUND", "User not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(user))
}

// Delete godoc
//
//	@Summary		Delete user
//	@Description	Delete a user by ID. Only ADMIN role.
//	@Tags			Users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"User ID"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		404
//	@Router			/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid user ID"))
		return
	}
	if err := h.userService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"message": "User deleted successfully"}))
}
