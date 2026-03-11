package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5" // Gunakan Echo Context

	"golang-mmi/internal/config"
	"golang-mmi/internal/service"
)

type AuthHandler struct {
	jwtService  *config.JWTService
	userService service.UserService
}

func NewAuthHandler(jwtService *config.JWTService, userService service.UserService) *AuthHandler {
	return &AuthHandler{
		jwtService:  jwtService,
		userService: userService,
	}
}

type LoginRequest struct {
	Nama     string `json:"nama"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

func (h *AuthHandler) Login(c *echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	user, err := h.userService.ValidateCredentials(c.Request().Context(), req.Password, req.Nama)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	tokenPair, err := h.jwtService.GenerateTokenPair(c.Request().Context(), user.Id, user.Jabatan)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate tokens"})
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/auth/refresh",
		MaxAge:   int(7 * 24 * time.Hour / time.Second),
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"access_token": tokenPair.AccessToken,
		"expires_at":   tokenPair.ExpiresAt,
	})
}

func (h *AuthHandler) Refresh(c *echo.Context) error {
	var refreshToken string

	cookie, err := c.Cookie("refresh_token")
	if err == nil {
		refreshToken = cookie.Value
	} else {
		var req RefreshRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "refresh token required"})
	}

	// Extract Claims
	token, _, _ := jwt.NewParser().ParseUnverified(refreshToken, &config.CustomClaims{})
	claims, ok := token.Claims.(*config.CustomClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
	}

	user, err := h.userService.GetUserByID(c.Request().Context(), claims.UserID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not found"})
	}

	tokenPair, err := h.jwtService.RefreshTokens(c.Request().Context(), refreshToken, user.Jabatan)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/auth/refresh",
		MaxAge:   int(7 * 24 * time.Hour / time.Second),
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"access_token": tokenPair.AccessToken,
		"expires_at":   tokenPair.ExpiresAt,
	})
}

func (h *AuthHandler) Logout(c *echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 {
			_ = h.jwtService.RevokeAccessToken(c.Request().Context(), parts[1])
		}
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/auth/refresh",
		MaxAge:   -1,
	})

	return c.NoContent(http.StatusNoContent)
}
