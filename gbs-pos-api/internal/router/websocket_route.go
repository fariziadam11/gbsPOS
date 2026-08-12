package router

import (
	"net/http"

	ws "gbs-pos-api/internal/websocket"

	"github.com/gin-gonic/gin"
)

func setupWebSocketRoute(r *gin.Engine, hub *ws.Hub, authMiddleware gin.HandlerFunc) {
	r.GET("/ws", authMiddleware, func(c *gin.Context) {
		clientType := c.Query("client_type")
		id := c.Query("terminal_id")
		if clientType == ws.ClientCompanion {
			id = c.Query("device_id")
		}
		if (clientType != ws.ClientPOS && clientType != ws.ClientCompanion) || id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "INVALID_CLIENT"})
			return
		}
		hub.ServeHTTP(c.Writer, c.Request, clientType, id)
	})
}
