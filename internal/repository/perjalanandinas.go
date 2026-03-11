package repository

import (
	"context"
	"fmt"
	"golang-mmi/internal/model"

	"gorm.io/gorm"
)

type PerjalananDinasRepository interface {
	CreatePengajuanPerjalanaDinas(ctx context.Context, perjalananDinas *model.RequestPPD) error
	GetListRiwayatPerjalananDinas(ctx context.Context, page, limit int) ([]RiwayatPPDResponse, int64, error)
	GetListRiwayatPerjalananDinasByUserID(ctx context.Context, userID uint, page, limit int) ([]RiwayatPPDResponse, int64, error)
	GetDetailPerjalananDinas(ctx context.Context, ppdid uint) (model.RequestPPD, error)
	ApprovePerjalananDinas(ctx context.Context, p ApprovePerjalananDinasparams) error
	DeclinePerjalananDinas(ctx context.Context, p DeclinePerjalananDinasParams) error
	GetLastNomorDokumenGeneral(ctx context.Context, pattern string) (string, error)
	GetLastNomorDokumenSpecific(ctx context.Context, tipe string, pattern string) (string, error)
	GetListPPDForRealisasi(ctx context.Context, userID uint) ([]DropdownPPDResponse, error)
	GetListRiwayatPerjalananDinasByAtasan(ctx context.Context, userID uint, page int, limit int) ([]RiwayatPPDResponse, int64, error)
	GetListPendingPerjalananDinas(ctx context.Context, jabatan string, userID uint) ([]model.RequestPPD, error)
	GetStatusPerjalananDinas(ctx context.Context, ppdid uint) (string, error)
}

type ApprovePerjalananDinasparams struct {
	RequestPPDID uint
	NextStatus   string
	NewDokumen   []model.Dokumen
	Riwayat      *model.RiwayatApproval
}

type RiwayatPPDResponse struct {
	ID               uint   `json:"id"`
	Nama             string `json:"nama,omitempty"`
	NomorDokumen     string `json:"nomor_dokumen"`
	Tujuan           string `json:"tujuan"`
	Keperluan        string `json:"keperluan"`
	TotalEstimasi    int    `json:"total_estimasi"`
	Status           string `json:"status"`
	PeriodeBerangkat string `json:"periode_berangkat"`
}

type DropdownPPDResponse struct {
	ID            uint   `json:"id"`
	NomorDokumen  string `json:"nomor_dokumen"`
	Tujuan        string `json:"tujuan"`
	TotalEstimasi int64  `json:"total_estimasi"`
}

type DeclinePerjalananDinasParams struct {
	RequestPPDID uint
	NextStatus   string
	Riwayat      *model.RiwayatApproval
}

type PerjalananDinas struct {
	db *gorm.DB
}

func NewPerjalananDinasRepository(db *gorm.DB) PerjalananDinasRepository {
	return &PerjalananDinas{
		db: db,
	}
}

func (r *PerjalananDinas) CreatePengajuanPerjalanaDinas(ctx context.Context, perjalananDinas *model.RequestPPD) error {
	return r.db.WithContext(ctx).Create(perjalananDinas).Error
}

func (r *PerjalananDinas) GetListRiwayatPerjalananDinas(ctx context.Context, page int, limit int) ([]RiwayatPPDResponse, int64, error) {
	var listData []RiwayatPPDResponse
	var totalData int64

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	query := r.db.WithContext(ctx).
		Table("request_ppd").
		Select(`
			request_ppd.id, 
			users.nama,
			dokumens.nomor_dokumen, 
			request_ppd.tujuan, 
			request_ppd.keperluan, 
			request_ppd.total_estimasi, 
			request_ppd.status, 
			request_ppd.periode_berangkat
		`).
		Joins("LEFT JOIN dokumens ON dokumens.doc_ref_id = request_ppd.id AND dokumens.doc_ref_type = 'RequestPPD' AND dokumens.tipe_dokumen = 'bonsementara'").
		Joins("LEFT JOIN users ON users.id = request_ppd.user_id")

	err := query.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.
		Order("request_ppd.periode_berangkat DESC").
		Offset(offset).
		Limit(limit).
		Find(&listData).Error

	return listData, totalData, err
}

func (r *PerjalananDinas) GetStatusPerjalananDinas(ctx context.Context, ppdid uint) (string, error) {
	var status string
	err := r.db.WithContext(ctx).
		Model(&model.RequestPPD{}).
		Select("status").
		Where("id = ?", ppdid).
		Scan(&status).Error

	return status, err
}

func (r *PerjalananDinas) GetDetailPerjalananDinas(ctx context.Context, ppdid uint) (model.RequestPPD, error) {
	var detailData model.RequestPPD

	err := r.db.WithContext(ctx).
		Preload("RincianTambahan").
		Preload("RincianHotel").
		Preload("RincianTransportasi").
		Preload("RealisasiBonSementara").
		Preload("RiwayatPersetujuan").
		Preload("Dokumen").
		First(&detailData, ppdid).Error

	return detailData, err
}

func (r *PerjalananDinas) ApprovePerjalananDinas(ctx context.Context, p ApprovePerjalananDinasparams) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.RequestPPD{}).
			Where("id = ?", p.RequestPPDID).
			Update("status", p.NextStatus).Error; err != nil {
			return err
		}

		if len(p.NewDokumen) > 0 {
			if err := tx.Create(&p.NewDokumen).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(p.Riwayat).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *PerjalananDinas) GetListRiwayatPerjalananDinasByUserID(ctx context.Context, userID uint, page int, limit int) ([]RiwayatPPDResponse, int64, error) {
	var listData []RiwayatPPDResponse
	var totalData int64

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	query := r.db.WithContext(ctx).
		Table("request_ppds").
		Select(`
			request_ppds.id, 
			dokumens.nomor_dokumen, 
			request_ppds.tujuan, 
			request_ppds.keperluan, 
			request_ppds.total_estimasi, 
			request_ppds.status, 
			request_ppds.periode_berangkat
		`).
		Joins("LEFT JOIN dokumens ON dokumens.doc_ref_id = request_ppds.id AND dokumens.doc_ref_type = 'RequestPPD' AND dokumens.tipe_dokumen = 'bonsementara'").
		Where("request_ppds.user_id = ?", userID)
	err := query.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.
		Order("request_ppds.periode_berangkat DESC").
		Offset(offset).
		Limit(limit).
		Find(&listData).Error

	return listData, totalData, err
}

func (r *PerjalananDinas) GetListRiwayatPerjalananDinasByAtasan(ctx context.Context, userID uint, page int, limit int) ([]RiwayatPPDResponse, int64, error) {
	var listData []RiwayatPPDResponse
	var totalData int64

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	query := r.db.WithContext(ctx).
		Table("request_ppds").
		Select(`
			request_ppds.id, 
			dokumens.nomor_dokumen, 
			request_ppds.tujuan, 
			request_ppds.keperluan, 
			request_ppds.total_estimasi, 
			request_ppds.status, 
			request_ppds.periode_berangkat
		`).
		Joins("LEFT JOIN dokumens ON dokumens.doc_ref_id = request_ppds.id AND dokumens.doc_ref_type = 'RequestPPD' AND dokumens.tipe_dokumen = 'bonsementara'").
		Joins("INNER JOIN users ON user.id = user_id").
		Where("users.atasan_id = ? OR users.id = ?", userID, userID)
	err := query.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.
		Order("request_ppds.periode_berangkat DESC").
		Offset(offset).
		Limit(limit).
		Find(&listData).Error

	return listData, totalData, err
}

func (r *PerjalananDinas) GetListPendingPerjalananDinas(ctx context.Context, jabatan string, userID uint) ([]model.RequestPPD, error) {
	var listData []model.RequestPPD

	query := r.db.Debug().WithContext(ctx).Table("request_ppds").Select("request_ppds.*")

	switch jabatan {
	case "atasan":
		query = query.Joins("INNER JOIN users ON users.id = request_ppds.user_id").
			Where("request_ppds.status = ? AND users.atasan_id = ?", "Menunggu Atasan", userID)
	case "HRGA":
		query = query.Where("status = ?", "Menunggu HRGA")
	case "Direktur":
		query = query.Where("status = ?", "Menunggu Direktur")
	case "Finance":
		query = query.Where("status = ?", "Menunggu Finance")
	default:
		return []model.RequestPPD{}, nil
	}

	fmt.Println("=== DEBUG GORM ===")
fmt.Printf("Mengeksekusi query untuk Atasan dengan userID: %v\n", userID)

	err := query.Order("periode_berangkat desc").Find(&listData).Error

	return listData, err
}

func (r *PerjalananDinas) GetLastNomorDokumenGeneral(ctx context.Context, pattern string) (string, error) {
	var lastDoc string

	err := r.db.WithContext(ctx).Model(&model.Dokumen{}).
		Where("nomor_dokumen LIKE ?", pattern).
		Order("nomor_dokumen desc").
		First(&lastDoc).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return lastDoc, nil
}

func (r *PerjalananDinas) GetLastNomorDokumenSpecific(ctx context.Context, tipe string, pattern string) (string, error) {
	var LastDocspecific string

	err := r.db.WithContext(ctx).Model(&model.Dokumen{}).
		Where("tipe_dokumen = ? AND nomor_dokumen LIKE ?", tipe, pattern).
		Order("nomor_dokumen desc").
		First(&LastDocspecific).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return LastDocspecific, nil
}

func (r *PerjalananDinas) GetListPPDForRealisasi(ctx context.Context, userID uint) ([]DropdownPPDResponse, error) {
	var list []DropdownPPDResponse

	err := r.db.WithContext(ctx).
		Table("request_ppds").
		Select(`request_ppds.id, 
				dokumens.nomor_tipe_dokumen AS nomor_dokumen, 
				request_ppd.tujuan,
				request_ppd.total_estimasi`).
		Joins("INNER JOIN dokumens ON dokumens.doc_ref_id = request_ppds.id AND dokumens.doc_ref_type = 'RequestPPD'").
		Where("request_ppds.user_id = ? AND request_ppds.status = ?", userID, "Disetujui").
		Where("rbs.id IS NULL").
		Find(&list).Error

	return list, err
}

func (r *PerjalananDinas) DeclinePerjalananDinas(ctx context.Context, p DeclinePerjalananDinasParams) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.RequestPPD{}).
			Where("id = ?", p.RequestPPDID).
			Update("status", p.NextStatus).Error; err != nil {
			return err
		}

		if p.Riwayat != nil {
			if err := tx.Create(p.Riwayat).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
