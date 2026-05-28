package router

import (
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupSettlementRoutes(
	rg *gin.RouterGroup,
	settlementHandler *handler.SettlementHandler,
) {

	rg.GET("/settlements", settlementHandler.List)
	rg.GET("/settlements/:id", settlementHandler.Get)
}