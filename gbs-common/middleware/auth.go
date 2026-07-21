package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"gbs-common/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthConfig holds configuration for the composite auth middleware
type AuthConfig struct {
	JWKSURL     string // Keycloak JWKS URL (e.g., https://auth.example.com/realms/myrealm/protocol/openid-connect/certs)
	JWTSecret   string // JWT secret for legacy/local auth
	AllowDual   bool   // If true, accept both JWT and Keycloak tokens
}

// NewCompositeAuthMiddleware creates a middleware that accepts both JWT (HS256) and Keycloak (RS256) tokens.
// It automatically detects token type based on algorithm in the token header.
// For Keycloak tokens (RS256), it validates against the JWKS endpoint.
// For legacy JWT tokens (HS256), it validates using the provided secret.
func NewCompositeAuthMiddleware(jwksURL, jwtSecret string) (gin.HandlerFunc, error) {
	var keycloakHandler gin.HandlerFunc
	var err error

	if jwksURL != "" {
		keycloakHandler, err = NewKeycloakMiddleware(jwksURL)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Keycloak middleware: %w", err)
		}
	}

	legacyHandler := NewAuthMiddleware(jwtSecret)

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

		// Parse token header to detect algorithm
		parser := jwt.NewParser()
		token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("INVALID_TOKEN", "Invalid token format"),
			)
			return
		}

		alg, _ := token.Header["alg"].(string)

		// Route based on algorithm
		switch alg {
		case "RS256":
			// Keycloak token (RS256)
			if keycloakHandler != nil {
				keycloakHandler(c)
				return
			}
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("INVALID_TOKEN", "Keycloak authentication is not configured"),
			)
			return

		case "HS256", "HS384", "HS512":
			// Legacy JWT token (HS256)
			if jwtSecret != "" {
				legacyHandler(c)
				return
			}
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("INVALID_TOKEN", "JWT authentication is not configured"),
			)
			return

		default:
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error("INVALID_TOKEN", fmt.Sprintf("Unsupported token algorithm: %s", alg)),
			)
			return
		}
	}, nil
}

// NewCompositeAuthMiddlewareWithConfig creates auth middleware with explicit configuration
func NewCompositeAuthMiddlewareWithConfig(cfg AuthConfig) (gin.HandlerFunc, error) {
	return NewCompositeAuthMiddleware(cfg.JWKSURL, cfg.JWTSecret)
}

func NewAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	secret := []byte(jwtSecret)
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
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		},
			jwt.WithValidMethods([]string{"HS256"}),
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
