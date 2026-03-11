package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"golang-mmi/internal/model"
	"golang-mmi/internal/service"
)

type RealisasiBonsHandler struct {
	service service.ServiceRBS
}

func NewRealisasiBonHandler(service service.ServiceRBS) *RealisasiBonsHandler {
	return &RealisasiBonsHandler{
		service: service,
	}
}

func (h *RealisasiBonsHandler) ApproveRBS(c *echo.Context) error {
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

	err = h.service.ApproveRBS(ctx, uint(ppdid), reqBody.Catatan)
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

func (h *RealisasiBonsHandler) DeclineRBS(c *echo.Context) error {
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

	err = h.service.DeclineRBS(ctx, uint(ppdid), reqBody.Catatan)
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

func (h *RealisasiBonsHandler) GetListRBS(c *echo.Context) error {
	var req model.GetListRBSRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Format query parameter tidak valid",
			"error":   err.Error(),
		})
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	ctx := c.Request().Context()
	data, totalData, totalPage, err := h.service.GetListRBS(ctx, req.Page, req.Limit, req.Tahun, req.Bulan)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Gagal mengambil daftar Realisasi Bon Sementara",
			"error":   err.Error(),
		})
	}

	if len(data) == 0 {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Data tidak ditemukan",
			"data":    []interface{}{},
			"meta": map[string]interface{}{
				"page":       req.Page,
				"limit":      req.Limit,
				"total_data": 0,
				"total_page": 0,
			},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Sukses mengambil daftar Realisasi Bon Sementara",
		"data":    data,
		"meta": map[string]interface{}{
			"page":       req.Page,
			"limit":      req.Limit,
			"total_data": totalData,
			"total_page": totalPage,
		},
	})
}

func (h *RealisasiBonsHandler) GetListPendingRBS(c *echo.Context) error {
	ctx := c.Request().Context()
	data, err := h.service.GetListPendingRBS(ctx)
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

func (h *RealisasiBonsHandler) CreateRealisasiBon(c *echo.Context) error {
	var req model.CreateRBSRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Format data request tidak valid",
			"error":   err.Error(),
		})
	}

	ctx := c.Request().Context()

	var rincianModel []model.RBSrincian
	for _, r := range req.RBSrincian {
		rincianModel = append(rincianModel, model.RBSrincian{
			HargaUnit: r.Harga,
			Kuantitas: r.Jumlah,
		})
	}

	reqModel := &model.RealisasiBonSementara{
		RequestPPDID: req.RequestPPDID,
		RBSrincian:   rincianModel,
	}

	err := h.service.CreateRealisasiBon(ctx, reqModel)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Gagal membuat realisasi",
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Realisasi Bon berhasil diajukan",
	})
}

func (h *RealisasiBonsHandler) GetListRBSDetail(c *echo.Context) error {
	idParam := c.Param("id")

	ppdid, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Format ID dokumen tidak valid. ID harus berupa angka.",
			"error":   err.Error(),
		})
	}

	ctx := c.Request().Context()

	data, err := h.service.GetListRBSDetail(ctx, uint(ppdid))
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

func (h *RealisasiBonsHandler) GetDropdownPPD(c *echo.Context) error {
	ctx := c.Request().Context()

	data, err := h.service.GetDropdownPPD(ctx)
	if err != nil {

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Gagal mengambil daftar PPD untuk dropdown",
			"error":   err.Error(),
		})
	}

	if len(data) == 0 {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Tidak ada data PPD yang tersedia untuk direalisasikan",
			"data":    []interface{}{},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Sukses mengambil data dropdown PPD",
		"data":    data,
	})
}
