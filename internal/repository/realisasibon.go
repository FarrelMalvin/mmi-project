package repository

import (
	"context"
	"fmt"
	"golang-mmi/internal/model"
	"golang-mmi/internal/constant"

	"gorm.io/gorm"
)

type RealisasiBonRepository interface {
	CreateRealisasiBon(ctx context.Context, realisasiBon *model.RealisasiBonSementara) error
	GetListRiwayatRealisasiBon(ctx context.Context, page int, limit int, f FilterRBS) ([]model.RBSListView, int64, int64, error)
	GetListRiwayatRealisasiBonById(ctx context.Context, page int, limit int, userID uint) ([]model.RBSListView, int64, int64, error)
	ApproveRBS(ctx context.Context, p ApproveRBSParam) error
	DeclineRBS(ctx context.Context, p DeclineRBSParam) error
	GetListPendingRBS(ctx context.Context, jabatan string, userID uint) ([]model.RBSListView, error)
	GetStatusRBS(ctx context.Context, rbsid uint) (string, error)
	GetDetailRBS(ctx context.Context, rbsid uint) (model.RealisasiBonSementara, error)
	GetDataRBSforCsv(ctx context.Context, f FilterRBS) ([]model.RBSDataforCsv, error)
	
	
}

func NewRealisasiBonRepository(db *gorm.DB) RealisasiBonRepository {
    return &RealisasiBon{
        db: db,
    }
}


type RealisasiBon struct {
	db *gorm.DB
}


type FilterRBS struct {
	Tahun int
	Bulan int
}

type ApproveRBSParam struct {
	RealisasiBonID uint
	NextStatus     string
	NewDokumen     []model.Dokumen
	Riwayat        *model.RiwayatApproval
}

type DeclineRBSParam struct {
	RealisasiBonID uint
	NextStatus     string
	Riwayat        *model.RiwayatApproval
}

type PPDforRealisasiItem struct{
	Kuantitas int
	HargaUnit int64
	Total int64
	
}

func (r *RealisasiBon) CreateRealisasiBon(ctx context.Context, realisasiBon *model.RealisasiBonSementara) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(realisasiBon).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *RealisasiBon) GetListRiwayatRealisasiBon(ctx context.Context, page int, limit int, f FilterRBS) ([]model.RBSListView, int64, int64, error) {
	var listData []model.RBSListView
	var totaldata int64
	var totalsum int64

	offset := (page - 1) * limit

	baseQuery := r.db.WithContext(ctx).
		Table("realisasi_bon_sementaras").
		Joins("INNER JOIN request_ppds AS rp ON realisasi_bon_sementaras.request_ppd_id = rp.id").
		Joins("LEFT JOIN dokumens ON dokumens.doc_ref_id = realisasi_bon_sementaras.id AND dokumens.doc_ref_type = 'RealisasiBonSementara'").
		Joins("LEFT JOIN users ON users.id = rp.user_id")

	if f.Tahun > 0 {
		baseQuery = baseQuery.Where("EXTRACT(YEAR FROM realisasi_bon_sementaras.created_at) = ?", f.Tahun)
	}
	if f.Bulan > 0 {
		baseQuery = baseQuery.Where("EXTRACT(MONTH FROM realisasi_bon_sementaras.created_at) = ?", f.Bulan)
	}

	baseQuery.Count(&totaldata)

	baseQuery.Session(&gorm.Session{}).
		Select("COALESCE(SUM(realisasi_bon_sementaras.total_realisasi), 0)").
		Scan(&totalsum)

	err := baseQuery.Session(&gorm.Session{}).
		Select(`
			realisasi_bon_sementaras.id,
			dokumens.nomor_dokumen,
			users.nama,
			realisasi_bon_sementaras.total_realisasi,
			rp.total_estimasi,
			(rp.total_estimasi - realisasi_bon_sementaras.total_realisasi) AS selisih,
			realisasi_bon_sementaras.status
		`).
		Order("realisasi_bon_sementaras.id DESC").
		Offset(offset).
		Limit(limit).
		Scan(&listData).Error

	return listData, totaldata, totalsum, err
}
func (r *RealisasiBon) GetListRiwayatRealisasiBonById(ctx context.Context, page int, limit int, userID uint) ([]model.RBSListView, int64, int64,error) {
	var listData []model.RBSListView
	var totaldata int64
	var totalpage int64

	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).
		Table("realisasi_bon_sementaras").
		Select(`
			realisasi_bon_sementaras.id,
			dokumens.nomor_dokumen,
			realisasi_bon_sementaras.total_realisasi,
			request_ppds.total_estimasi,
			(request_ppds.total_estimasi - realisasi_bon_sementaras.total_realisasi) AS selisih,
			realisasi_bon_sementaras.status
		`).
		Joins("INNER JOIN request_ppds ON realisasi_bon_sementaras.request_ppd_id = request_ppds.id").
		Joins("LEFT JOIN dokumens ON dokumens.realisasi_bon_id = realisasi_bon_sementaras.id").
		Where("request_ppds.user_id = ?", userID).
		Offset(offset).
		Limit(limit)

	if err := query.Scan(&totaldata).Error; err != nil {
		return nil, 0, 0, err
	}

	err := query.
		Order("realisasi_bon_sementaras.id DESC").
		Offset(offset).
		Limit(limit).
		Find(&listData).Error

	return listData, totaldata, totalpage,err
}

func (r *RealisasiBon) ApproveRBS(ctx context.Context, p ApproveRBSParam) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.RealisasiBonSementara{}).
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

func (r *RealisasiBon) GetListPendingRBS(ctx context.Context, jabatan string, userID uint) ([]model.RBSListView, error) {
	var listData []model.RBSListView

	query := r.db.WithContext(ctx).
		Table("realisasi_bon_sementaras").
		Select(`
			realisasi_bon_sementaras.id,
			realisasi_bon_sementaras.nomor_bon_sementara AS nomor_dokumen,
			realisasi_bon_sementaras.total_realisasi,
			request_ppds.total_estimasi,
			(request_ppds.total_estimasi - realisasi_bon_sementaras.total_realisasi) AS selisih,
			realisasi_bon_sementaras.status
		`).
		Joins("INNER JOIN request_ppds ON realisasi_bon_sementaras.request_ppd_id = request_ppds.id").
		Joins("LEFT JOIN dokumens ON dokumens.doc_ref_id = realisasi_bon_sementaras.id")
	
	switch jabatan {
	case constant.JabatanAtasan:
		query = query.Joins("INNER JOIN user ON user.id = requestppd.user_id").
			Where("realisasi_bon_sementaras.status = ? AND user.atasan_id = ?", constant.StatusMenungguAtasan, userID)
	case constant.JabatanHRGA:
		query = query.Where("realisasi_bon_sementaras.status = ?", constant.StatusMenungguHRGA)
	case constant.JabatanFinance:
		query = query.Where("realisasi_bon_sementaras.status = ?", constant.StatusMenungguFinance)
	default:
		return []model.RBSListView{}, nil
	}

	err := query.Order("realisasi_bon_sementaras.periode_berangkat desc").Find(&listData).Error

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

func (r *RealisasiBon) DeclineRBS(ctx context.Context, p DeclineRBSParam) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.RealisasiBonSementara{}).
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
		Preload("RBSrincian").
		Preload("Dokumen").
		Preload("RiwayatPersetujuan.User").
		First(&detailData, rbsid).Error

	return detailData, err
}
func (r *RealisasiBon) GetDataRBSforCsv(ctx context.Context, f FilterRBS) ([]model.RBSDataforCsv, error) {
    var data []model.RBSDataforCsv

    query := r.db.WithContext(ctx).
        Table("realisasi_bon_sementaras").
        Select(`
            realisasi_bon_sementaras.total_realisasi,
            realisasi_bon_sementaras.periode_berangkat AS periode,
            rb_srincians.uraian,
            rb_srincians.kategori,
            rb_srincians.kuantitas,
            rb_srincians.tanggal_transaksi,
            rb_srincians.harga_unit,
            rb_srincians.total_harga,
            users.nama,
            realisasi_bon_sementaras.nomor_bon_sementara AS nomor_referensi_bs
        `).
        Joins("LEFT JOIN rb_srincians ON rb_srincians.rbs_id = realisasi_bon_sementaras.id").
        Joins("LEFT JOIN users ON users.id = realisasi_bon_sementaras.user_id").
        Where("realisasi_bon_sementaras.status = ?", constant.StatusSelesai)

    if f.Bulan > 0 {
        query = query.Where(
            "EXTRACT(MONTH FROM realisasi_bon_sementaras.periode_berangkat) = ?",
            f.Bulan,
        )
    }
    if f.Tahun > 0 {
        query = query.Where(
            "EXTRACT(YEAR FROM realisasi_bon_sementaras.periode_berangkat) = ?",
            f.Tahun,
        )
    }

    err := query.
        Order("rb_srincians.kategori, rb_srincians.tanggal_transaksi").
        Scan(&data).Error
    if err != nil {
        return nil, fmt.Errorf("gagal mengambil data RBS untuk excel: %w", err)
    }

    return data, nil
}