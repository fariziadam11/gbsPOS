package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"gbs-common/pkg/response"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func NewKeycloakMiddleware(jwksURL string) (gin.HandlerFunc, error) {
	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS keyfunc: %w", err)
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("UNAUTHORIZED", "Missing authorization header"),
			)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, jwks.Keyfunc,
			jwt.WithValidMethods([]string{"RS256"}),
			jwt.WithExpirationRequired(),
			jwt.WithLeeway(5*time.Second),
		)

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("INVALID_TOKEN", "Invalid or expired token"),
			)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("INVALID_TOKEN", "Invalid token claims"),
			)
			return
		}

		c.Set("userID", claims["sub"])
		c.Set("username", keycloakUsername(claims))
		c.Set("role", keycloakRole(claims))
		c.Next()
	}, nil
}

func keycloakUsername(claims jwt.MapClaims) string {
	if v, ok := claims["preferred_username"].(string); ok && v != "" {
		return v
	}
	if v, ok := claims["sub"].(string); ok {
		return v
	}
	return ""
}

func keycloakRole(claims jwt.MapClaims) string {
	realmAccess, ok := claims["realm_access"].(map[string]any)
	if ok {
		if rolesAny, ok := realmAccess["roles"].([]any); ok {
			for _, r := range rolesAny {
				if role, ok := r.(string); ok {
					if role == "ADMIN" || role == "CASHIER" {
						return role
					}
				}
			}
		}
	}

	if v, ok := claims["role"].(string); ok {
		return v
	}
	return ""
}
