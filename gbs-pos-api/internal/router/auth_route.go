package router

import (
	"gbs-pos-api/internal/config"
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupAuthRoutes(
	rg *gin.RouterGroup,
	authHandler *handler.AuthHandler,
	cfg *config.Config,
) {
	if !cfg.UseKeycloak() || cfg.EnableDemoAuth {
		rg.POST("/login", authHandler.Login)
	}
}