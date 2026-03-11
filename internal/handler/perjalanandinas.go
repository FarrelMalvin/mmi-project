package handler

import (
	"net/http"

	"strconv"
	"strings"
	"context"

	"github.com/labstack/echo/v5"

	"golang-mmi/internal/model"
	"golang-mmi/internal/service"
	"golang-mmi/internal/config"
)

type PerjalananDinasHandler struct {
	service service.ServicePPD
}

func NewPerjalananDinasHandler(service service.ServicePPD) *PerjalananDinasHandler {
	return &PerjalananDinasHandler{
		service: service,
	}
}

func (h *PerjalananDinasHandler) ApprovePerjalananDinas(c *echo.Context) error {
	idParam := c.Param("id")
	ppdid, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Format ID dokumen tidak valid",
		})
	}

	var reqBody model.ApprovePPDRequest
	if err := c.Bind(&reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Gagal membaca data request body",
		})
	}

	ctx := c.Request().Context()


	err = h.service.ApprovePerjalananDinas(ctx, uint(ppdid), reqBody.Catatan)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "perjalanan dinas tidak ditemukan" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "tidak dapat menyetujui perjalanan dinas dengan status saat ini" || err.Error() == "status perjalanan dinas tidak valid" {
			statusCode = http.StatusUnprocessableEntity
		}

		return c.JSON(statusCode, map[string]interface{}{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Dokumen Perjalanan Dinas berhasil disetujui",
	})
}

func (h *PerjalananDinasHandler) DeclinePerjalananDinas(c *echo.Context) error {
	idParam := c.Param("id")
	ppdid, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Format ID dokumen tidak valid",
		})
	}

	var reqBody model.DeclinePPDRequest
	if err := c.Bind(&reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Gagal membaca data request body",
		})
	}

	ctx := c.Request().Context()

	err = h.service.ApprovePerjalananDinas(ctx, uint(ppdid), reqBody.Catatan)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "perjalanan dinas tidak ditemukan" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "tidak dapat menolak perjalanan dinas dengan status saat ini" || err.Error() == "status perjalanan dinas tidak valid" {
			statusCode = http.StatusUnprocessableEntity
		}

		return c.JSON(statusCode, map[string]interface{}{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Dokumen Perjalanan Dinas berhasil disetujui",
	})
}

func (h *PerjalananDinasHandler) GetRiwayatPerjalananDinas(c *echo.Context) error {

	ctx := c.Request().Context()

	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	data, total, err := h.service.GetListPerjalananDinas(ctx, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Gagal mengambil data riwayat",
			"error":   err.Error(),
		})
	}

	totalPage := int(total) / limit
	if int(total)%limit > 0 {
		totalPage++
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Sukses mengambil riwayat",
		"data":    data,
		"meta": map[string]interface{}{
			"page":       page,
			"limit":      limit,
			"total_data": total,
			"total_page": totalPage,
		},
	})
}

func (h *PerjalananDinasHandler) GetListPendingPerjalananDinas(c *echo.Context) error {
	ctx := c.Request().Context()
	data, err := h.service.GetListPendingPerjalananDinas(ctx)
	if err != nil {
		statusCode := http.StatusInternalServerError

		if err.Error() == "invalid user ID in context" || err.Error() == "invalid jabatan in context" {
			statusCode = http.StatusUnauthorized
		}

		return c.JSON(statusCode, map[string]interface{}{
			"message": "Gagal mengambil daftar dokumen pending",
			"error":   err.Error(),
		})
	}

	if len(data) == 0 {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Tidak ada dokumen perjalanan dinas yang menunggu persetujuan Anda",
			"data":    []interface{}{},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Sukses mengambil daftar perjalanan dinas",
		"data":    data,
	})
}

func (h *PerjalananDinasHandler) CreatePengajuanPerjalanaDinas(c *echo.Context) error {
	var reqBody model.RequestPPD

	if err := c.Bind(&reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Format data request tidak valid",
			"error":   err.Error(),
		})
	}

	ctx := c.Request().Context()

	claims, ok := config.GetClaimsFromContext(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"message": "Gagal mengambil data profil dari token",
		})
	}

	reqBody.UserID = claims.UserID

	ctx = context.WithValue(ctx, "user_id", claims.UserID)
	ctx = context.WithValue(ctx, "jabatan", claims.Jabatan)

	err := h.service.CreatePengajuanPerjalanaDinas(ctx, &reqBody)
	if err != nil {
		statusCode := http.StatusInternalServerError

		if err.Error() == "invalid jabatan untuk membuat pengajuan" ||
			err.Error() == "invalid user ID in context" ||
			err.Error() == "invalid jabatan in context" {
			statusCode = http.StatusUnauthorized

		} else if strings.HasPrefix(err.Error(), "kategori rincian tidak valid") {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, map[string]interface{}{
			"message": "Gagal membuat pengajuan perjalanan dinas",
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Pengajuan Perjalanan Dinas berhasil dibuat dan sedang menunggu persetujuan",
	})
}

func (h *PerjalananDinasHandler) GetListPerjalananDetail(c *echo.Context) error {
	idParam := c.Param("id")

	ppdid, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Format ID dokumen tidak valid. ID harus berupa angka.",
			"error":   err.Error(),
		})
	}

	ctx := c.Request().Context()


	data, err := h.service.GetListPerjalananDetail(ctx, uint(ppdid))
	if err != nil {
		statusCode := http.StatusInternalServerError

		if err.Error() == "detail perjalanan dinas tidak ditemukan" {
			statusCode = http.StatusNotFound
		}

		return c.JSON(statusCode, map[string]interface{}{
			"message": "Gagal mengambil data detail perjalanan dinas",
			"error":   err.Error(),
		})
	}
	if len(data) == 0 {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Tidak ada rincian tambahan untuk perjalanan dinas ini",
			"data":    []interface{}{},
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Sukses mengambil detail rincian tambahan",
		"data":    data,
	})
}
