package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"gbs-common/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func NewCompositeAuthMiddleware(jwksURL, jwtSecret string) (gin.HandlerFunc, error) {
	keycloakHandler, err := NewKeycloakMiddleware(jwksURL)
	if err != nil {
		return nil, err
	}

	legacyHandler := NewAuthMiddleware(jwtSecret)

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("[AUTH] missing authorization header | path=%s", path)
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("UNAUTHORIZED", "Missing authorization header"),
			)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		parser := jwt.NewParser()
		token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			log.Printf("[AUTH] invalid token format | path=%s error=%v", path, err)
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("INVALID_TOKEN", "Invalid token format"),
			)
			return
		}

		alg, _ := token.Header["alg"].(string)
		log.Printf("[AUTH] token algorithm | path=%s alg=%s", path, alg)
		if alg == "RS256" {
			keycloakHandler(c)
			return
		}

		if jwtSecret != "" {
			legacyHandler(c)
			return
		}

		log.Printf("[AUTH] unsupported token algorithm | path=%s alg=%s", path, alg)
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			response.Error("INVALID_TOKEN", "Unsupported token algorithm"),
		)
	}, nil
}

func NewAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	secret := []byte(jwtSecret)
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("[AUTH] missing authorization header | path=%s", path)
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("UNAUTHORIZED", "Missing authorization header"),
			)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		},
			jwt.WithValidMethods([]string{"HS256"}),
			jwt.WithExpirationRequired(),
			jwt.WithLeeway(5*time.Second),
		)

		if err != nil || !token.Valid {
			log.Printf("[AUTH] invalid or expired legacy token | path=%s error=%v", path, err)
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("INVALID_TOKEN", "Invalid or expired token"),
			)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("[AUTH] invalid legacy token claims | path=%s", path)
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("INVALID_TOKEN", "Invalid token claims"),
			)
			return
		}

		c.Set("userID", claims["sub"])
		c.Set("username", claims["username"])
		c.Set("role", claims["role"])
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(
			http.StatusForbidden,
			response.Error(
				"INSUFFICIENT_PERMISSIONS",
				"You don't have permission to access this resource",
			),
		)
	}
}
