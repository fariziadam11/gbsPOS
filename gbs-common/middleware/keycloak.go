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

// KeycloakConfig holds configuration for Keycloak authentication
type KeycloakConfig struct {
	JWKSURL  string // Keycloak JWKS URL
	ClientID string // Optional: expected client_id in token
}

// NewKeycloakMiddleware creates a middleware that validates Keycloak/OIDC tokens using JWKS.
// The token must be signed with RS256 and contain either ADMIN or CASHIER role in realm_access.
func NewKeycloakMiddleware(jwksURL string) (gin.HandlerFunc, error) {
	return NewKeycloakMiddlewareWithConfig(KeycloakConfig{
		JWKSURL: jwksURL,
	})
}

// NewKeycloakMiddlewareWithConfig creates Keycloak middleware with explicit configuration
func NewKeycloakMiddlewareWithConfig(cfg KeycloakConfig) (gin.HandlerFunc, error) {
	if cfg.JWKSURL == "" {
		return nil, fmt.Errorf("JWKS URL is required")
	}

	// Create JWKS from the Keycloak JWKS endpoint
	jwks, err := keyfunc.NewDefault([]string{cfg.JWKSURL})
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
				response.Error("INVALID_TOKEN", "Invalid or expired Keycloak token"),
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

		// Validate client_id if configured
		if cfg.ClientID != "" {
			if clientID, _ := claims["azp"].(string); clientID != cfg.ClientID {
				c.AbortWithStatusJSON(
					http.StatusUnauthorized,
					response.Error("INVALID_TOKEN", "Invalid client_id"),
				)
				return
			}
		}

		role := keycloakRole(claims)
		username := keycloakUsername(claims)
		if role == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("INVALID_TOKEN", "No ADMIN/CASHIER role found in Keycloak token"),
			)
			return
		}

		// Set standard claims in context
		c.Set("userID", claims["sub"])
		c.Set("username", username)
		c.Set("role", role)

		// Optionally set additional Keycloak claims
		if email, ok := claims["email"].(string); ok {
			c.Set("email", email)
		}
		if name, ok := claims["name"].(string); ok {
			c.Set("name", name)
		}

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
