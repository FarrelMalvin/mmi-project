package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/go-playground/validator/v10"

	"golang-mmi/internal/constant"
	"golang-mmi/internal/dto"
	"golang-mmi/internal/middleware"
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

var validaterbs = validator.New()

func (h *RealisasiBonsHandler) ApproveRBS(c *echo.Context) error {
	var req dto.ApproveRBSRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "Format data request tidak valid",
		})
	}

	idParam := c.Param("id")
	rbsID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || rbsID == 0 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "ID Realisasi Bon Sementara tidak valid",
		})
	}
	req.RealisasiBonID = uint(rbsID)

	ctx := c.Request().Context()
	claims, ok := middleware.GetClaimsFromContext(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "Unauthorized",
			Message: "Sesi tidak valid atau tidak memiliki akses",
		})
	}
	req.UserID = claims.UserID
	req.Jabatan = claims.Jabatan

	if err := h.service.ApproveRBS(ctx, &req); err != nil {
		if errors.Is(err, service.ErrRBSTidakDitemukan) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:    http.StatusNotFound,
				Status:  "Not Found",
				Message: "Realisasi bon sementara tidak ditemukan",
			})
		}
		if errors.Is(err, service.ErrRBSStatusTidakValid) {
			return c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{
				Code:    http.StatusUnprocessableEntity,
				Status:  "Unprocessable Entity",
				Message: "Status tidak valid untuk aksi ini",
			})
		}
		if errors.Is(err, service.ErrRBSJabatanTidakValid) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Code:    http.StatusForbidden,
				Status:  "Forbidden",
				Message: "Jabatan tidak memiliki izin untuk aksi ini",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Gagal menyetujui realisasi bon sementara",
		})
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Code:    http.StatusOK,
		Status:  "Success",
		Message: "Realisasi bon sementara berhasil disetujui",
	})
}

func (h *RealisasiBonsHandler) DeclineRBS(c *echo.Context) error {
	var req dto.DeclineRBSRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "Format data request tidak valid",
		})
	}

	idParam := c.Param("id")
	rbsID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || rbsID == 0 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "ID Realisasi Bon Sementara tidak valid",
		})
	}
	req.RealisasiBonID = uint(rbsID)

	ctx := c.Request().Context()
	claims, ok := middleware.GetClaimsFromContext(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "Unauthorized",
			Message: "Sesi tidak valid atau tidak memiliki akses",
		})
	}
	req.UserID = claims.UserID
	req.Jabatan = claims.Jabatan

	if err := h.service.DeclineRBS(ctx, req); err != nil {
		if errors.Is(err, service.ErrRBSTidakDitemukan) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:    http.StatusNotFound,
				Status:  "Not Found",
				Message: "Realisasi bon sementara tidak ditemukan",
			})
		}
		if errors.Is(err, service.ErrRBSStatusTidakValid) {
			return c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{
				Code:    http.StatusUnprocessableEntity,
				Status:  "Unprocessable Entity",
				Message: "Status tidak valid untuk aksi ini",
			})
		}
		if errors.Is(err, service.ErrRBSJabatanTidakValid) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Code:    http.StatusForbidden,
				Status:  "Forbidden",
				Message: "Jabatan tidak memiliki izin untuk aksi ini",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Gagal menolak realisasi bon sementara",
		})
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Code:    http.StatusOK,
		Status:  "Success",
		Message: "Realisasi bon sementara berhasil ditolak",
	})
}

func (h *RealisasiBonsHandler) GetListRBS(c *echo.Context) error {
	var req dto.RBSListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "Format query parameter tidak valid",
		})
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	
	ctx := c.Request().Context()
	claims, ok := middleware.GetClaimsFromContext(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "Unauthorized",
			Message: "Sesi tidak valid atau tidak memiliki akses",
		})
	}

	req.UserID = claims.UserID
	req.Jabatan = claims.Jabatan

	data, totalData, totalSum, err := h.service.GetListRBS(ctx, req)
	if err != nil {
		if errors.Is(err, service.ErrRBSJabatanTidakValid) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Code:    http.StatusForbidden,
				Status:  "Forbidden",
				Message: "Anda tidak memiliki akses untuk melihat daftar ini",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Gagal mengambil daftar realisasi bon sementara",
		})
	}

	totalPage := int(totalData) / req.Limit
	if int(totalData)%req.Limit > 0 {
		totalPage++
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Code:    http.StatusOK,
		Status:  "Success",
		Message: "Sukses mengambil daftar realisasi bon sementara",
		Data:    data,
		Meta: map[string]interface{}{
			"page":       req.Page,
			"limit":      req.Limit,
			"total_data": totalData,
			"total_sum":  totalSum,
			"total_page": totalPage,
		},
	})
}

func (h *RealisasiBonsHandler) GetListPendingRBS(c *echo.Context) error {
	ctx := c.Request().Context()
	claims, ok := middleware.GetClaimsFromContext(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "Unauthorized",
			Message: "Sesi tidak valid atau tidak memiliki akses",
		})
	}

	data, err := h.service.GetListPendingRBS(ctx, claims.UserID, claims.Jabatan)
	if err != nil {
		if errors.Is(err, service.ErrRBSAksesditolak) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Code:    http.StatusForbidden,
				Status:  "Forbidden",
				Message: "Anda tidak memiliki akses ke fitur ini",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Gagal mengambil daftar pending realisasi bon sementara",
		})
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Code:    http.StatusOK,
		Status:  "Success",
		Message: "Sukses mengambil daftar pending realisasi bon sementara",
		Data:    data,
	})
}

func (h *RealisasiBonsHandler) CreateRealisasiBon(c *echo.Context) error {
	var req dto.CreateRBSRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "Format data request tidak valid",
		})
	}

	if err := validaterbs.Struct(req); err != nil {
		errMessage := "Data tidak lengkap atau tidak valid"
		if validationErrors, ok := err.(validator.ValidationErrors); ok && len(validationErrors) > 0 {
		}

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: errMessage,
		})
	}

	ctx := c.Request().Context()
	claims, ok := middleware.GetClaimsFromContext(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "Unauthorized",
			Message: "Sesi tidak valid atau tidak memiliki akses",
		})
	}
	req.UserID = claims.UserID
	req.Jabatan = claims.Jabatan

	if err := h.service.CreateRealisasiBon(ctx, req); err != nil {
		if errors.Is(err, service.ErrRBSJabatanTidakValid) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Code:    http.StatusForbidden,
				Status:  "Forbidden",
				Message: "Jabatan tidak memiliki izin untuk membuat realisasi",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Gagal membuat realisasi bon sementara",
		})
	}

	return c.JSON(http.StatusCreated, dto.SuccessResponse{
		Code:    http.StatusCreated,
		Status:  "Created",
		Message: "Realisasi bon sementara berhasil diajukan",
	})
}

func (h *RealisasiBonsHandler) GetListRBSDetail(c *echo.Context) error {
	idParam := c.Param("id")
	rbsID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || rbsID == 0 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "Format ID dokumen tidak valid",
		})
	}

	ctx := c.Request().Context()
	data, err := h.service.GetRBSDetail(ctx, uint(rbsID))
	if err != nil {
		if errors.Is(err, service.ErrRBSTidakDitemukan) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:    http.StatusNotFound,
				Status:  "Not Found",
				Message: "Detail realisasi bon sementara tidak ditemukan",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Gagal mengambil detail realisasi bon sementara",
		})
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Code:    http.StatusOK,
		Status:  "Success",
		Message: "Sukses mengambil detail realisasi bon sementara",
		Data:    data,
	})
}

func (h *RealisasiBonsHandler) GetDropdownPPD(c *echo.Context) error {
	ctx := c.Request().Context()
	claims, ok := middleware.GetClaimsFromContext(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "Unauthorized",
			Message: "Sesi tidak valid atau tidak memiliki akses",
		})
	}

	data, err := h.service.GetDropdownPPD(ctx, claims.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Gagal mengambil daftar PPD untuk dropdown",
		})
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Code:    http.StatusOK,
		Status:  "Success",
		Message: "Sukses mengambil data dropdown PPD",
		Data:    data,
	})
}

func (h *RealisasiBonsHandler) GenerateRBSPDF(c *echo.Context) error {
	ctx := c.Request().Context()
	claims, ok := middleware.GetClaimsFromContext(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "Unauthorized",
			Message: "Sesi tidak valid atau tidak memiliki akses",
		})
	}

	idParam := c.Param("id")
	rbsID, err := strconv.Atoi(idParam)
	if err != nil || rbsID <= 0 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "ID Realisasi Bon Sementara tidak valid",
		})
	}

	filename := fmt.Sprintf("Realisasi_Bon_Sementara_%d.pdf", rbsID)
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	err = h.service.FillRBSPDF(ctx, uint(rbsID), claims.UserID, constant.RBSTemplatePath, c.Response())
	if err != nil {
		c.Response().Header().Del("Content-Type")
		c.Response().Header().Del("Content-Disposition")

		if errors.Is(err, service.ErrRBSAksesditolak) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Code:    http.StatusForbidden,
				Status:  "Forbidden",
				Message: "Akses ditolak",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Gagal menghasilkan PDF Realisasi Bon Sementara",
		})
	}

	return nil
}

func (h *RealisasiBonsHandler) DownloadExcel(c *echo.Context) error {
	ctx := c.Request().Context()
	claims, ok := middleware.GetClaimsFromContext(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "Unauthorized",
			Message: "Sesi tidak valid atau tidak memiliki akses",
		})
	}

	var req dto.RBSListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "Format filter tidak valid",
		})
	}
	req.Jabatan = claims.Jabatan
	req.UserID = claims.UserID

	filename := buildExcelFilename(req.Month, req.Year)
	c.Response().Header().Set(
		"Content-Type",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	)
	c.Response().Header().Set(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"`, filename),
	)

	if err := h.service.ExportRBSExcel(ctx, req, c.Response()); err != nil {
		c.Response().Header().Del("Content-Type")
		c.Response().Header().Del("Content-Disposition")

		if errors.Is(err, service.ErrRBSAksesditolak) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Code:    http.StatusForbidden,
				Status:  "Forbidden",
				Message: "Hanya HRGA yang dapat mengakses fitur export excel",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Gagal menghasilkan file Excel",
		})
	}

	return nil
}

func buildExcelFilename(bulan, tahun int) string {
	if bulan > 0 && tahun > 0 {
		return fmt.Sprintf("Rekap_Realisasi_Bon_Sementara_%02d_%d.xlsx", bulan, tahun)
	}
	if tahun > 0 {
		return fmt.Sprintf("Rekap_Realisasi_Bon_Sementara_%d.xlsx", tahun)
	}
	return "Rekap_Realisasi_Bon_Sementara_Semua.xlsx"
}