package handler

import(
	"net/http"

	"github.com/labstack/echo/v5"

	"golang-mmi/internal/dto"
	"golang-mmi/internal/middleware"
	"golang-mmi/internal/service"
)

type UploadHandler struct{
	service service.UploadService

}

func NewUploadHandler(service service.UploadService) *UploadHandler{
	return &UploadHandler{
		service: service,
	}
}
	
func (h *UploadHandler) UploadStruk(c *echo.Context) error {
    ctx := c.Request().Context()
    claims, ok := middleware.GetClaimsFromContext(ctx)
    if !ok {
        return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
            Code:    http.StatusUnauthorized,
            Status:  "Unauthorized",
            Message: "Sesi tidak valid",
        })
    }

    file, err := c.FormFile("struk")
    if err != nil {
        return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Code:    http.StatusBadRequest,
            Status:  "Bad Request",
            Message: "File struk tidak ditemukan",
        })
    }

    if !isValidImageType(file.Header.Get("Content-Type")) {
        return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Code:    http.StatusBadRequest,
            Status:  "Bad Request",
            Message: "Tipe file tidak valid, hanya JPG/PNG yang diizinkan",
        })
    }

    if file.Size > 2*1024*1024 {
        return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Code:    http.StatusBadRequest,
            Status:  "Bad Request",
            Message: "Ukuran file maksimal 2MB",
        })
    }

    url, err := h.service.UploadStruk(ctx, file, claims.UserID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
            Code:    http.StatusInternalServerError,
            Status:  "Internal Server Error",
            Message: "Gagal mengupload struk",
        })
    }

    return c.JSON(http.StatusOK, dto.SuccessResponse{
        Code:    http.StatusOK,
        Status:  "Success",
        Message: "Struk berhasil diupload",
        Data: map[string]string{
            "url": url,
        },
    })
}

func isValidImageType(contentType string) bool {
    validTypes := map[string]bool{
        "image/jpeg": true,
        "image/jpg":  true,
        "image/png":  true,
    }
    return validTypes[contentType]
}