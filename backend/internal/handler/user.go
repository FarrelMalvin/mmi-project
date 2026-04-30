package handler

import (
	"net/http"
	"errors"

	"github.com/labstack/echo/v5"

	"golang-mmi/internal/middleware"
	"golang-mmi/internal/service"
	"golang-mmi/internal/dto"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) UpdateSignature(c *echo.Context) error {
	ctx := c.Request().Context()
	claims, _:= middleware.GetClaimsFromContext(ctx)

	file, err := c.FormFile("signature_file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "File tanda tangan tidak ditemukan",
		})
	}

	path, err := h.service.UpdateSignaturePath(ctx, claims.UserID, file) 
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": err.Error(),
        })
    }

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Tanda tangan berhasil diunggah",
		"signature_url": path,
	})

}

func(h *UserHandler) GetDataProfile (c *echo.Context) error{
	ctx := c.Request().Context()
	claims, ok := middleware.GetClaimsFromContext(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "Unauthorized",
			Message: "Sesi tidak valid atau tidak memiliki akses",
		})
	}


	data, err := h.service.GetDataProfile(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrAksesditolak) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Code:    http.StatusForbidden,
				Status:  "Forbidden",
				Message: "Anda tidak memiliki akses untuk melihat daftar ini",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Gagal mengambil data riwayat perjalanan dinas",
		})
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Code:    http.StatusOK,
		Status:  "Success",
		Message: "Sukses mengambil Data Profile",
		Data:    data,
	})
	
}
