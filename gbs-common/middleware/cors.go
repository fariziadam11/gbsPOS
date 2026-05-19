package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{
			"https://cms.gbs.com",
			"http://localhost:5173",
			"http://localhost:3000",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Client-Type",
		},
		ExposeHeaders:    []string{"Content-Length", "X-Last-Sync"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
