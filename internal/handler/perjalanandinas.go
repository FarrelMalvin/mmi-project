package handler

import (
	"net/http"

	"errors"
	"fmt"
	"strconv"


	"github.com/labstack/echo/v5"
	"github.com/go-playground/validator/v10"

	"golang-mmi/internal/constant"
	"golang-mmi/internal/dto"
	"golang-mmi/internal/middleware"
	"golang-mmi/internal/service"
)

type PerjalananDinasHandler struct {
	service service.ServicePPD
}

func NewPerjalananDinasHandler(service service.ServicePPD) *PerjalananDinasHandler {
	return &PerjalananDinasHandler{
		service: service,
	}
}

var validate = validator.New()

func (h *PerjalananDinasHandler) ApprovePerjalananDinas(c *echo.Context) error {
	var req dto.ApprovePPDRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "Format data request tidak valid",
		})
	}

	idPpdStr := c.Param("id")
	idPpd, err := strconv.Atoi(idPpdStr)
	if err != nil || idPpd <= 0 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "ID Perjalanan Dinas tidak valid",
		})
	}
	req.RequestPPDID = uint(idPpd)

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

	err = h.service.ApprovePerjalananDinas(ctx, req)
	if err != nil {
		if errors.Is(err, service.ErrPPDTidakDitemukan) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:    http.StatusNotFound,
				Status:  "Not Found",
				Message: "Perjalanan dinas tidak ditemukan",
			})
		}
		if errors.Is(err, service.ErrStatusTidakValid) {
			return c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{
				Code:    http.StatusUnprocessableEntity,
				Status:  "Unprocessable Entity",
				Message: "Status tidak valid untuk aksi ini",
			})
		}
		if errors.Is(err, service.ErrJabatanTidakValid) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Code:    http.StatusForbidden,
				Status:  "Forbidden",
				Message: "Jabatan tidak memiliki izin untuk aksi ini",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Gagal menyetujui perjalanan dinas",
		})
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Code:    http.StatusOK,
		Status:  "Success",
		Message: "Perjalanan dinas berhasil disetujui",
	})
}

func (h *PerjalananDinasHandler) DeclinePerjalananDinas(c *echo.Context) error {
	var req dto.DeclinePPDRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "Format data request tidak valid",
		})
	}

	idPpdStr := c.Param("id")
	idPpd, err := strconv.Atoi(idPpdStr)
	if err != nil || idPpd <= 0 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "ID Perjalanan Dinas tidak valid",
		})
	}
	req.RequestPPDID = uint(idPpd)

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

	err = h.service.DeclinePerjalananDinas(ctx, req)
	if err != nil {
		if errors.Is(err, service.ErrPPDTidakDitemukan) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:    http.StatusNotFound,
				Status:  "Not Found",
				Message: "Perjalanan dinas tidak ditemukan",
			})
		}
		if errors.Is(err, service.ErrStatusTidakValid) {
			return c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{
				Code:    http.StatusUnprocessableEntity,
				Status:  "Unprocessable Entity",
				Message: "Status tidak valid untuk aksi ini",
			})
		}
		if errors.Is(err, service.ErrJabatanTidakValid) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Code:    http.StatusForbidden,
				Status:  "Forbidden",
				Message: "Jabatan tidak memiliki izin untuk aksi ini",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Gagal menolak perjalanan dinas",
		})
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Code:    http.StatusOK,
		Status:  "Success",
		Message: "Perjalanan dinas berhasil diolak",
	})
}
func (h *PerjalananDinasHandler) GetRiwayatPerjalananDinas(c *echo.Context) error {
	var req dto.ListPPDRequest
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

	data, total, err := h.service.GetListPerjalananDinas(ctx, req)
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

	totalPage := int(total) / req.Limit
	if int(total)%req.Limit > 0 {
		totalPage++
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Code:    http.StatusOK,
		Status:  "Success",
		Message: "Sukses mengambil riwayat perjalanan dinas",
		Data:    data,
		Meta: map[string]interface{}{
			"page":       req.Page,
			"limit":      req.Limit,
			"total_data": total,
			"total_page": totalPage,
		},
	})
}

func (h *PerjalananDinasHandler) GetListPendingPerjalananDinas(c *echo.Context) error {
	var req dto.ListPPDRequest

	ctx := c.Request().Context()
	claims, ok := middleware.GetClaimsFromContext(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Status:  "Unauthorized",
			Message: "Sesi tidak valid atau tidak memiliki akses",
		})
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	req.UserID = claims.UserID
	req.Jabatan = claims.Jabatan

	data, total, err := h.service.GetListPendingPerjalananDinas(ctx, req)
	if err != nil {
		if errors.Is(err, service.ErrAksesditolak) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Code:    http.StatusForbidden,
				Status:  "Forbidden",
				Message: "Anda tidak memiliki akses ke fitur ini",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Status:  "Internal Server Error",
			Message: "Gagal mengambil daftar dokumen pending",
		})
	}

	totalPage := int(total) / req.Limit
	if int(total)%req.Limit > 0 {
		totalPage++
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Code:    http.StatusOK,
		Status:  "Success",
		Message: "Sukses mengambil daftar pending perjalanan dinas",
		Data:    data,
		Meta: map[string]interface{}{
			"page":       req.Page,
			"limit":      req.Limit,
			"total_data": total,
			"total_page": totalPage,
		},
	})
}

func (h *PerjalananDinasHandler) CreatePengajuanPerjalanaDinas(c *echo.Context) error {
    var req dto.CreatePPDRequest

    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Code:    http.StatusBadRequest,
            Status:  "Bad Request",
            Message: "Format data request tidak valid",
        })
    }

	if err := validate.Struct(req); err != nil {
		errMessage := "Data tidak lengkap atau tidak valid"
		if validationErrors, ok := err.(validator.ValidationErrors); ok && len(validationErrors) > 0 {
			errMessage = "Kesalahan pada field: " + validationErrors[0].Field()
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

    err := h.service.CreatePengajuanPerjalanaDinas(ctx, req)
    if err != nil {
        if errors.Is(err, service.ErrJabatanTidakValid) {
            return c.JSON(http.StatusForbidden, dto.ErrorResponse{
                Code:    http.StatusForbidden,
                Status:  "Forbidden",
                Message: "Jabatan tidak memiliki izin untuk membuat pengajuan",
            })
        }
        if errors.Is(err, service.ErrKategoriTidakValid) {
            return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
                Code:    http.StatusBadRequest,
                Status:  "Bad Request",
                Message: "Kategori rincian tidak valid",
            })
        }
        return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
            Code:    http.StatusInternalServerError,
            Status:  "Internal Server Error",
            Message: "Gagal membuat pengajuan perjalanan dinas",
        })
    }

    return c.JSON(http.StatusCreated, dto.SuccessResponse{
        Code:    http.StatusCreated,
        Status:  "Created",
        Message: "Pengajuan Perjalanan Dinas berhasil dibuat dan sedang menunggu persetujuan",
    })
}

func (h *PerjalananDinasHandler) GetPerjalananDetail(c *echo.Context) error {
    idParam := c.Param("id")
    ppdid, err := strconv.ParseUint(idParam, 10, 32)
    if err != nil || ppdid == 0 {
        return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Code:    http.StatusBadRequest,
            Status:  "Bad Request",
            Message: "Format ID dokumen tidak valid",
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

    data, err := h.service.GetPerjalananDetail(ctx, uint(ppdid), claims.UserID, claims.Jabatan)
    if err != nil {
        if errors.Is(err, service.ErrPPDTidakDitemukan) {
            return c.JSON(http.StatusNotFound, dto.ErrorResponse{
                Code:    http.StatusNotFound,
                Status:  "Not Found",
                Message: "Perjalanan dinas tidak ditemukan",
            })
        }
        if errors.Is(err, service.ErrAksesditolak) {
            return c.JSON(http.StatusForbidden, dto.ErrorResponse{
                Code:    http.StatusForbidden,
                Status:  "Forbidden",
                Message: "Akses ditolak, dokumen bukan milik Anda",
            })
        }
        return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
            Code:    http.StatusInternalServerError,
            Status:  "Internal Server Error",
            Message: "Gagal mengambil data detail perjalanan dinas",
        })
    }

    return c.JSON(http.StatusOK, dto.SuccessResponse{
        Code:    http.StatusOK,
        Status:  "Success",
        Message: "Sukses mengambil detail perjalanan dinas",
        Data:    data,
    })
}

func (h *PerjalananDinasHandler) GetItemsByPPDID(c *echo.Context) error {
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
    ppdID, err := strconv.Atoi(idParam)
    if err != nil || ppdID <= 0 {
        return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Code:    http.StatusBadRequest,
            Status:  "Bad Request",
            Message: "ID Perjalanan Dinas tidak valid",
        })
    }

    data, err := h.service.GetItemsByPPDID(ctx, uint(ppdID), claims.UserID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
            Code:    http.StatusInternalServerError,
            Status:  "Internal Server Error",
            Message: "Gagal mengambil data rincian item",
        })
    }

    return c.JSON(http.StatusOK, dto.SuccessResponse{
        Code:    http.StatusOK,
        Status:  "Success",
        Message: "Berhasil mengambil rincian item perjalanan dinas",
        Data:    data,
    })
}
func (h *PerjalananDinasHandler) GeneratePPDPDF(c *echo.Context) error {
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
    ppdID, err := strconv.Atoi(idParam)
    if err != nil || ppdID <= 0 {
        return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Code:    http.StatusBadRequest,
            Status:  "Bad Request",
            Message: "ID Perjalanan Dinas tidak valid",
        })
    }

    templatePath := constant.PPDTemplatePath

    filename := fmt.Sprintf("Pengajuan_Perjalanan_Dinas_%d.pdf", ppdID)
    c.Response().Header().Set("Content-Type", "application/pdf")
    c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

    err = h.service.FillPPDPDF(ctx, uint(ppdID), claims.UserID, templatePath, claims.Jabatan, c.Response())
    if err != nil {
        c.Response().Header().Del("Content-Type")
        c.Response().Header().Del("Content-Disposition")

		if errors.Is(err, service.ErrAksesditolak) {
            return c.JSON(http.StatusForbidden, dto.ErrorResponse{
                Code:    http.StatusForbidden,
                Status:  "Forbidden",
                Message: "Akses ditolak, dokumen bukan milik Anda",
            })
        }

        return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
            Code:    http.StatusInternalServerError,
            Status:  "Internal Server Error",
            Message: "Gagal menghasilkan PDF Pengajuan Perjalanan Dinas",
        })
    }

    return nil
}

func (h *PerjalananDinasHandler) GenerateBSPDF(c *echo.Context) error {
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
    ppdID, err := strconv.Atoi(idParam)
    if err != nil || ppdID <= 0 {
        return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Code:    http.StatusBadRequest,
            Status:  "Bad Request",
            Message: "ID Perjalanan Dinas tidak valid",
        })
    }

    templatePath := constant.BSTemplatePath

    filename := fmt.Sprintf("Bon_Sementara_%d.pdf", ppdID)
    c.Response().Header().Set("Content-Type", "application/pdf")
    c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

    err = h.service.FillBSPDF(ctx, uint(ppdID), claims.UserID, templatePath, c.Response())
    if err != nil {
        c.Response().Header().Del("Content-Type")
        c.Response().Header().Del("Content-Disposition")

        return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
            Code:    http.StatusInternalServerError,
            Status:  "Internal Server Error",
            Message: "Gagal menghasilkan PDF Bon Sementara",
        })
    }

    return nil
}