package middleware

import (
	"fmt"
	"log"
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
		log.Printf("[AUTH] validating keycloak token | path=%s len=%d", path, len(tokenString))

		token, err := jwt.Parse(tokenString, jwks.Keyfunc,
			jwt.WithValidMethods([]string{"RS256"}),
			jwt.WithExpirationRequired(),
			jwt.WithLeeway(5*time.Second),
		)

		if err != nil || !token.Valid {
			log.Printf("[AUTH] invalid or expired keycloak token | path=%s error=%v", path, err)
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("INVALID_TOKEN", "Invalid or expired token"),
			)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("[AUTH] invalid keycloak token claims | path=%s", path)
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("INVALID_TOKEN", "Invalid token claims"),
			)
			return
		}

		role := keycloakRole(claims)
		username := keycloakUsername(claims)
		log.Printf("[AUTH] keycloak token valid | path=%s username=%s role=%s", path, username, role)
		if role == "" {
			log.Printf("[AUTH] keycloak token has no ADMIN/CASHIER role | path=%s username=%s", path, username)
		}

		c.Set("userID", claims["sub"])
		c.Set("username", username)
		c.Set("role", role)
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
