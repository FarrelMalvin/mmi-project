package config

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// ClaimsContextKey is the context key for storing JWT claims.
	ClaimsContextKey ContextKey = "claims"
)

// AuthMiddleware creates an HTTP middleware that validates JWT tokens.
// It extracts the token from the Authorization header and validates it.
func AuthMiddleware(jwtService *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			// Verify Bearer token format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Validate the token
			claims, err := jwtService.ValidateAccessToken(r.Context(), tokenString)
			if err != nil {
				switch {
				case errors.Is(err, ErrExpiredToken):
					http.Error(w, "token expired", http.StatusUnauthorized)
				case errors.Is(err, ErrTokenRevoked):
					http.Error(w, "token revoked", http.StatusUnauthorized)
				default:
					http.Error(w, "invalid token", http.StatusUnauthorized)
				}
				return
			}

			// Add claims to the request context
			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaimsFromContext extracts the JWT claims from the request context.
func GetClaimsFromContext(ctx context.Context) (*CustomClaims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*CustomClaims)
	return claims, ok
}

func RequireRoles(jwtService *JWTService, allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Token tidak ditemukan"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Format token salah"})
			}
			tokenString := parts[1]

			claims, err := jwtService.ValidateAccessToken(c.Request().Context(), tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Token tidak valid atau kedaluwarsa"})
			}

			hasAccess := false
			for _, role := range allowedRoles {
				if claims.Jabatan == role {
					hasAccess = true
					break
				}
			}

			if !hasAccess {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Anda tidak memiliki hak akses untuk fitur ini"})
			}

			c.Set("user_claims", claims)

			ctx := context.WithValue(c.Request().Context(), ClaimsContextKey, claims)
	
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
