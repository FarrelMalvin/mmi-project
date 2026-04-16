package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"golang-mmi/internal/middleware"
	"golang-mmi/internal/service"
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
