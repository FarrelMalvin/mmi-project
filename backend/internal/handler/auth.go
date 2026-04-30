package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"

	"golang-mmi/internal/config"
	"golang-mmi/internal/dto"
	"golang-mmi/internal/service"
)

type AuthHandler struct {
	jwtService  *config.JWTService
	userService service.UserService
	authService service.AuthService
}

func NewAuthHandler(jwtService *config.JWTService, userService service.UserService, authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		jwtService:  jwtService,
		userService: userService,
		authService: authService,
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
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "Format request tidak valid",
		})
	}

	user, err := h.authService.ValidateCredentials(c.Request().Context(), req.Password, req.Nama)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredetial) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Status:  "Unauthorized",
				Message: "Nama pengguna atau password salah",
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Terjadi kesalahan pada sistem saat memvalidasi kredensial",
		})
	}

	tokenPair, err := h.jwtService.GenerateTokenPair(c.Request().Context(), user.Id, user.Jabatan, user.Nama)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Gagal membuat token otentikasi",
		})
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken,
		HttpOnly: true,
		Secure:   false, //true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/", //"/api/v1/auth/refresh",
		MaxAge:   int(7 * 24 * time.Hour / time.Second),
	})

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Code:    http.StatusOK,
		Status:  "OK",
		Message: "Login Berhasil",
		Data: map[string]interface{}{
			"access_token": tokenPair.AccessToken,
			"expires_at":   tokenPair.ExpiresAt,
		},
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
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:    http.StatusBadRequest,
				Status:  "Bad Request",
				Message: "Format request tidak valid",
			})
		}
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "Refresh token tidak ditemukan",
		})
	}

	token, _, _ := jwt.NewParser().ParseUnverified(refreshToken, &config.CustomClaims{})
	claims, ok := token.Claims.(*config.CustomClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "Unauthorized",
			Message: "Klaim token tidak valid",
		})
	}

	user, err := h.userService.GetUserByID(c.Request().Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Status:  "Unauthorized",
				Message: "Sesi tidak valid, akun pengguna tidak ditemukan",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Terjadi kesalahan saat memeriksa data pengguna",
		})
	}

	tokenPair, err := h.jwtService.RefreshTokens(c.Request().Context(), refreshToken, user.Jabatan, user.Nama)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "Unauthorized",
			Message: "Refresh token tidak valid atau sudah kedaluwarsa",
		})
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken,
		HttpOnly: true,
		Secure:   false, //nanti diganti true
		SameSite: http.SameSiteStrictMode,
		Path:     "/", //"/api/v1/auth/refresh",
		MaxAge:   int(7 * 24 * time.Hour / time.Second),
	})

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Code:    http.StatusOK,
		Status:  "OK",
		Message: "Login Berhasil",
		Data: map[string]interface{}{
			"access_token": tokenPair.AccessToken,
			"expires_at":   tokenPair.ExpiresAt,
		},
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
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/", //"/api/v1/auth/refresh",
		MaxAge:   -1,
	})

	return c.NoContent(http.StatusNoContent)
}
