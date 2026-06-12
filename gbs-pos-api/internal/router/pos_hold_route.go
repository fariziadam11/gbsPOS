package router

import (
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupPosHoldRoutes(
	rg *gin.RouterGroup,
	posHoldHandler *handler.PosHoldHandler,
) {
	pos := rg.Group("/pos/hold")

	pos.POST("", posHoldHandler.Hold)
	pos.GET("", posHoldHandler.List)
	pos.GET("/:id", posHoldHandler.Get)
	pos.POST("/:id/resume", posHoldHandler.Resume)
	pos.DELETE("/:id", posHoldHandler.Delete)
}