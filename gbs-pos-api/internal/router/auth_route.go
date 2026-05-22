package router

import (
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupAuthRoutes(
	rg *gin.RouterGroup,
	authHandler *handler.AuthHandler,
) {

	rg.POST("/login", authHandler.Login)
}