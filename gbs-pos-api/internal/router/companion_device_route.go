package router

import (
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupCompanionDeviceRoutes(auth *gin.RouterGroup, h *handler.CompanionDeviceHandler) {
	auth.POST("/companion/register", h.Register)
}
