package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"golang-mmi/internal/config"
	"golang-mmi/internal/dto"

)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// ClaimsContextKey is the context key for storing JWT claims.
	ClaimsContextKey ContextKey = "claims"
)

// AuthMiddleware creates an HTTP middleware that validates JWT tokens.
// It extracts the token from the Authorization header and validates it.
func AuthMiddleware(jwtService *config.JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Extract the Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			// Verify Bearer token format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			tokenString := parts[1]

			// Validate the token
			claims, err := jwtService.ValidateAccessToken(c.Request().Context(), tokenString)
			if err != nil {
				switch {
				case errors.Is(err, config.ErrExpiredToken):
					return echo.NewHTTPError(http.StatusUnauthorized, "token expired")
				case errors.Is(err, config.ErrTokenRevoked):
					return echo.NewHTTPError(http.StatusUnauthorized, "token revoked")
				default:
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
				}
			}

			// Menyisipkan claims ke dalam context request standar agar bisa dibaca oleh handler
			ctx := context.WithValue(c.Request().Context(), ClaimsContextKey, claims)
			c.SetRequest(c.Request().WithContext(ctx))

			// Lanjutkan ke handler berikutnya
			return next(c)
		}
	}
}
// GetClaimsFromContext extracts the JWT claims from the request context.
func GetClaimsFromContext(ctx context.Context) (*config.CustomClaims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*config.CustomClaims)
	return claims, ok
}
func RequireRoles(jwtService *config.JWTService, allowedRoles ...string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c *echo.Context) error {
            authHeader := c.Request().Header.Get("Authorization")
            if authHeader == "" {
                return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
                    Code:    http.StatusUnauthorized,
                    Status:  "Unauthorized",
                    Message: "Token tidak ditemukan",
                })
            }

            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
                    Code:    http.StatusUnauthorized,
                    Status:  "Unauthorized",
                    Message: "Format token tidak valid, gunakan format Bearer <token>",
                })
            }
            tokenString := parts[1]

            claims, err := jwtService.ValidateAccessToken(c.Request().Context(), tokenString)
            if err != nil {
                return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
                    Code:    http.StatusUnauthorized,
                    Status:  "Unauthorized",
                    Message: "Token tidak valid atau kedaluwarsa",
                })
            }

            hasAccess := false
            for _, role := range allowedRoles {
                if claims.Jabatan == role {
                    hasAccess = true
                    break
                }
            }

            if !hasAccess {
                return c.JSON(http.StatusForbidden, dto.ErrorResponse{
                    Code:    http.StatusForbidden,
                    Status:  "Forbidden",
                    Message: "Anda tidak memiliki hak akses untuk fitur ini",
                })
            }

            c.Set("user_claims", claims)
            ctx := context.WithValue(c.Request().Context(), ClaimsContextKey, claims)
            c.SetRequest(c.Request().WithContext(ctx))

            return next(c)
        }
    }
}