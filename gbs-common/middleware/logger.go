package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		if statusCode >= 500 {
			log.Error().
				Str("method", method).
				Str("path", path).
				Int("status", statusCode).
				Dur("latency", latency).
				Str("ip", clientIP).
				Str("client_type", c.GetHeader("X-Client-Type")).
				Msg("request")
		}
	}
}
