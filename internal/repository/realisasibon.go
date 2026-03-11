package repository

import (
	"context"
	"golang-mmi/internal/model"

	"gorm.io/gorm"
)

type RealisasiBonRepository interface {
	CreateRealisasiBon(ctx context.Context, realisasiBon *model.RealisasiBonSementara) error
	GetListRiwayatRealisasiBon(ctx context.Context, page int, limit int, f FilterRBS) ([]RiwayatRBSResponse, int64, int64, error)
	GetListRiwayatRealisasiBonById(ctx context.Context, page int, limit int, userID uint) ([]RiwayatRBSResponse, int64, int64, error)
	ApproveRBS(ctx context.Context, p ApproveRBSResponse) error
	DeclineRBS(ctx context.Context, p DeclineRBSResponse) error
	GetListPendingRBS(ctx context.Context, jabatan string, userID uint) ([]model.RealisasiBonSementara, error)
	GetStatusRBS(ctx context.Context, rbsid uint) (string, error)
	GetDetailRBS(ctx context.Context, rbsid uint) (model.RealisasiBonSementara, error)
}

func NewRealisasiBonRepository(db *gorm.DB) RealisasiBonRepository {
    return &RealisasiBon{
        db: db,
    }
}


type RealisasiBon struct {
	db *gorm.DB
}

type RiwayatRBSResponse struct {
	ID             uint   `json:"id"`
	NomorDokumen   string `json:"nomor_dokumen"`
	Nama           string `json:"nama,omitempty"`
	TotalRealisasi int64  `json:"total_realisasi"`
	TotalEstimasi  int64  `json:"total_estimasi"`
	Selisih        int64  `json:"selisih"`
	Status         string `json:"status"`
}

type FilterRBS struct {
	Tahun int
	Bulan int
}

type ApproveRBSResponse struct {
	RealisasiBonID uint
	NextStatus     string
	NewDokumen     []model.Dokumen
	Riwayat        *model.RiwayatApproval
}

type DeclineRBSResponse struct {
	RealisasiBonID uint
	NextStatus     string
	Riwayat        *model.RiwayatApproval
}

func (r *RealisasiBon) CreateRealisasiBon(ctx context.Context, realisasiBon *model.RealisasiBonSementara) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(realisasiBon).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *RealisasiBon) GetListRiwayatRealisasiBon(ctx context.Context, page int, limit int, f FilterRBS) ([]RiwayatRBSResponse, int64, int64, error) {
	var listData []RiwayatRBSResponse
	var totaldata int64
	var totalsum int64

	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).
		Table("realisasi_bon_sementara").
		Select(`
			realisasi_bon_sementara.id,
			dokumen.nomor_dokumen,
			user.nama,
			realisasi_bon_sementara.total_realisasi,
			request_ppd.total_estimasi,
			(request_ppd.total_estimasi - realisasi_bon_sementara.total_realisasi) AS selisih,
			realisasi_bon_sementara.status
		`).
		Joins("INNER JOIN request_ppd ON realisasi_bon_sementara.request_ppd_id = request_ppd.id").
		Joins("LEFT JOIN dokumen ON dokumen.realisasi_bon_id = realisasi_bon_sementara.id").
		Joins("LEFT JOIN user ON user.id = request_ppd.user_id").
		Offset(offset).
		Limit(limit)

	if f.Tahun > 0 {
		query = query.Where("YEAR(realisasi_bon_sementara.created_at) = ?", f.Tahun)
	}
	if f.Bulan > 0 {
		query = query.Where("MONTH(realisasi_bon_sementara.created_at) = ?", f.Bulan)
	}

	query.Count(&totaldata)

	query.Select(`Sum(realisasi_bon_sementara.total_realisasi)`).Scan(&totalsum)

	err := query.
		Order("realisasi_bon_sementara.id DESC").
		Offset(offset).
		Limit(limit).
		Find(&listData).Error

	return listData, totaldata, totalsum, err
}

func (r *RealisasiBon) GetListRiwayatRealisasiBonById(ctx context.Context, page int, limit int, userID uint) ([]RiwayatRBSResponse, int64, int64,error) {
	var listData []RiwayatRBSResponse
	var totaldata int64
	var totalpage int64

	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).
		Table("realisasi_bon_sementara").
		Select(`
			realisasi_bon_sementara.id,
			dokumen.nomor_dokumen,
			realisasi_bon_sementara.total_realisasi,
			request_ppd.total_estimasi,
			(request_ppd.total_estimasi - realisasi_bon_sementara.total_realisasi) AS selisih,
			realisasi_bon_sementara.status
		`).
		Joins("INNER JOIN request_ppd ON realisasi_bon_sementara.request_ppd_id = request_ppd.id").
		Joins("LEFT JOIN dokumen ON dokumen.realisasi_bon_id = realisasi_bon_sementara.id").
		Where("request_ppd.user_id = ?", userID).
		Offset(offset).
		Limit(limit)

	if err := query.Scan(&totaldata).Error; err != nil {
		return nil, 0, 0, err
	}

	err := query.
		Order("realisasi_bon_sementara.id DESC").
		Offset(offset).
		Limit(limit).
		Find(&listData).Error

	return listData, totaldata, totalpage,err
}

func (r *RealisasiBon) ApproveRBS(ctx context.Context, p ApproveRBSResponse) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.RequestPPD{}).
			Where("id = ?", p.RealisasiBonID).
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

func (r *RealisasiBon) GetListPendingRBS(ctx context.Context, jabatan string, userID uint) ([]model.RealisasiBonSementara, error) {
	var listData []model.RealisasiBonSementara

	query := r.db.WithContext(ctx).Table("realisasibonsementara")

	switch jabatan {
	case "atasan":
		query = query.Joins("INNER JOIN user ON user.id = requestppd.user_id").
			Where("requestppd.status = ? AND user.atasan_id = ?", "Menunggu Atasan", userID)
	case "Direktur":
		query = query.Where("status = ?", "Menunggu Direktur")
	case "Finance":
		query = query.Where("status = ?", "Menunggu Finance")
	default:
		return []model.RealisasiBonSementara{}, nil
	}

	err := query.Order("periode_berangkat desc").Find(&listData).Error

	return listData, err
}

func (r *RealisasiBon) GetStatusRBS(ctx context.Context, rbsid uint) (string, error) {
	var status string
	err := r.db.WithContext(ctx).
		Model(&model.RealisasiBonSementara{}).
		Select("status").
		Where("id = ?", rbsid).
		Scan(&status).Error

	return status, err
}

func (r *RealisasiBon) DeclineRBS(ctx context.Context, p DeclineRBSResponse) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.RequestPPD{}).
			Where("id = ?", p.RealisasiBonID).
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

func (r *RealisasiBon) GetDetailRBS(ctx context.Context, rbsid uint) (model.RealisasiBonSementara, error) {
	var detailData model.RealisasiBonSementara

	err := r.db.WithContext(ctx).
		Preload("RBSRincian").
		Preload("Dokumen").
		First(&detailData, rbsid).Error

	return detailData, err
}
