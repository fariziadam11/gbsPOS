package router

import (
	"gbs-common/middleware"
	"gbs-pos-api/internal/config"
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func Setup(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	productHandler *handler.ProductHandler,
	orderHandler *handler.OrderHandler,
	settlementHandler *handler.SettlementHandler,
) *gin.Engine {

	r := gin.New()
	r.MaxMultipartMemory = 32 << 20 // 32 MB
	r.Use(middleware.LoggerMiddleware())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	v1 := r.Group("/v1")
	{
		v1.POST("/login", authHandler.Login)

		auth := v1.Group("", middleware.NewAuthMiddleware(cfg.JWTSecret))
		{
			auth.GET("/products", productHandler.List)
			auth.POST("/products", middleware.RequireRole("ADMIN"), productHandler.Create)
			auth.PUT("/products/:id", middleware.RequireRole("ADMIN"), productHandler.Update)
			auth.DELETE("/products/:id", middleware.RequireRole("ADMIN"), productHandler.Delete)

			auth.GET("/orders", orderHandler.List)
			auth.GET("/orders/:id", orderHandler.Get)
			auth.POST("/orders", orderHandler.Create)
			auth.POST("/sync/orders", orderHandler.BulkSync)
			auth.PATCH("/orders/:id/void", middleware.RequireRole("ADMIN"), orderHandler.Void)
			auth.GET("/orders/unsettled/summary", orderHandler.UnsettledSummary)
			auth.POST("/orders/settle", middleware.RequireRole("ADMIN"), orderHandler.Settle)

			auth.GET("/settlements", settlementHandler.List)
			auth.GET("/settlements/:id", settlementHandler.Get)
		}
	}

	return r
}