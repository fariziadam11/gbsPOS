package router

import (
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupHoldRoutes(
	rg *gin.RouterGroup,
	holdHandler *handler.HoldHandler,
) {
	if holdHandler == nil {
		return
	}

	hold := rg.Group("/pos/hold")

	hold.POST("", holdHandler.Create)
	hold.GET("", holdHandler.List)
	hold.GET("/:id", holdHandler.Get)
	hold.POST("/:id/resume", holdHandler.Resume)
	hold.DELETE("/:id", holdHandler.Delete)
}
