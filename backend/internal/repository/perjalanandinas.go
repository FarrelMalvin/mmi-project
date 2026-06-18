package repository

import (
	"context"

	"gorm.io/gorm"

	"golang-mmi/internal/model"
)

type PerjalananDinasRepository interface {
	CreatePengajuanPerjalananDinas(ctx context.Context, perjalananDinas *model.RequestPPD, dokumen *model.Dokumen) error
	GetListRiwayatPerjalananDinas(ctx context.Context, page, limit int) ([]model.PPDListView, int64, error)
	GetListRiwayatPerjalananDinasByUserID(ctx context.Context, userID uint, page, limit int) ([]model.PPDListView, int64, error)
	GetDetailPerjalananDinas(ctx context.Context, ppdid uint) (model.RequestPPD, error)
	ApprovePerjalananDinas(ctx context.Context, p ApprovePerjalananDinasparams) error
	DeclinePerjalananDinas(ctx context.Context, p DeclinePerjalananDinasParams) error
	GetListPPDForRealisasi(ctx context.Context, userID uint) ([]model.DropdownPPDView, error)
	GetListRiwayatPerjalananDinasByAtasan(ctx context.Context, userID uint, page int, limit int) ([]model.PPDListView, int64, error)
	GetListPendingPerjalananDinas(ctx context.Context, jabatan string, userID uint, page int, limit int) ([]model.PPDListView, int64, error)
	GetStatusPerjalananDinas(ctx context.Context, ppdid uint) (string, error)
	GetItemsByPPDID(ctx context.Context, ppdID uint, userid uint) ([]model.PPDItemView, error)
	GetTotalEstimasi(ctx context.Context, ppdid uint) (int64, error)
	GetUserIDByPPDID(ctx context.Context, ppdID uint) (uint, error)
	GetNomorBS(ctx context.Context, ppdID uint) (string, error)
	UpdatePengajuanPerjalananDinas(ctx context.Context, data model.RequestPPD) error
}

type ApprovePerjalananDinasparams struct {
	RequestPPDID  uint
	NextStatus    string
	CurrentStatus string
	NewDokumen    []model.Dokumen
	Riwayat       *model.RiwayatApproval
}

type DeclinePerjalananDinasParams struct {
	RequestPPDID  uint
	NextStatus    string
	CurrentStatus string
	Riwayat       *model.RiwayatApproval
}

type Pagination struct {
	Page   int
	Limit  int
	Offset int
}

type PerjalananDinas struct {
	db *gorm.DB
}

func NewPerjalananDinasRepository(db *gorm.DB) PerjalananDinasRepository {
	return &PerjalananDinas{
		db: db,
	}
}

func (r *PerjalananDinas) CreatePengajuanPerjalananDinas(ctx context.Context, perjalananDinas *model.RequestPPD, dokumen *model.Dokumen) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(perjalananDinas).Error; err != nil {
			return err
		}

		if dokumen == nil {
			return nil
		}
		dokumen.DocRefID = perjalananDinas.Id	
		return tx.Create(dokumen).Error

	})
}

func (r *PerjalananDinas) GetListRiwayatPerjalananDinas(ctx context.Context, page int, limit int) ([]model.PPDListView, int64, error) {
	var listData []model.PPDListView
	var totalData int64

	pagination := normalizePagination(page, limit)
	offset := pagination.Offset

	if err := r.db.WithContext(ctx).Table("request_ppds").Count(&totalData).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Table("request_ppds").
		Select(`
			request_ppds.id, 
			users.nama,
			dokumens.nomor_tipe_dokumen, 
			dokumens.nomor_dokumen,
			request_ppds.tujuan, 
			request_ppds.keperluan, 
			request_ppds.total_estimasi, 
			request_ppds.status, 
			request_ppds.periode_berangkat
		`).
		Joins("LEFT JOIN dokumens ON dokumens.doc_ref_id = request_ppds.id AND dokumens.doc_ref_type = 'RequestPPD' AND dokumens.tipe_dokumen = 'Bon Sementara'").
		Joins("LEFT JOIN users ON users.id = request_ppds.user_id").
		Order("request_ppds.periode_berangkat DESC").
		Limit(limit).
		Offset(offset).
		Scan(&listData).Error

	return listData, totalData, err
}
func (r *PerjalananDinas) GetStatusPerjalananDinas(ctx context.Context, ppdid uint) (string, error) {
	var status string
	err := r.db.WithContext(ctx).
		Model(&model.RequestPPD{}).
		Select("status").
		Where("id = ?", ppdid).
		First(&status).Error

	if err != nil {
		return "", err
	}

	return status, err
}

func (r *PerjalananDinas) GetDetailPerjalananDinas(ctx context.Context, ppdid uint) (model.RequestPPD, error) {
	var detailData model.RequestPPD

	err := r.db.WithContext(ctx).
		Preload("RincianTambahan").
		Preload("User").
		Preload("RincianHotel").
		Preload("RincianTransportasi").
		Preload("RealisasiBonSementara").
		Preload("RiwayatPersetujuan.User").
		Preload("Dokumen").
		First(&detailData, ppdid).Error

	return detailData, err
}

func (r *PerjalananDinas) ApprovePerjalananDinas(ctx context.Context, p ApprovePerjalananDinasparams) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.RequestPPD{}).
			Where("id = ? AND status = ?", p.RequestPPDID, p.CurrentStatus).
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

func (r *PerjalananDinas) GetListRiwayatPerjalananDinasByUserID(ctx context.Context, userID uint, page int, limit int) ([]model.PPDListView, int64, error) {
	var listData []model.PPDListView
	var totalData int64

	Pagination := normalizePagination(page, limit)
	offset := Pagination.Offset
	query := r.db.WithContext(ctx).
		Table("request_ppds").
		Select(`
			request_ppds.id, 
			dokumens.nomor_tipe_dokumen, 
			dokumens.nomor_dokumen,
			request_ppds.tujuan, 
			request_ppds.keperluan, 
			request_ppds.total_estimasi, 
			request_ppds.status, 
			request_ppds.periode_berangkat
		`).
		Joins("LEFT JOIN dokumens ON dokumens.doc_ref_id = request_ppds.id AND dokumens.doc_ref_type = 'RequestPPD' AND dokumens.tipe_dokumen = 'Bon Sementara'").
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

func (r *PerjalananDinas) GetListRiwayatPerjalananDinasByAtasan(ctx context.Context, userID uint, page int, limit int) ([]model.PPDListView, int64, error) {
	var listData []model.PPDListView
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
			dokumens.nomor_tipe_dokumen, 
			dokumens.nomor_dokumen,
			request_ppds.tujuan, 
			request_ppds.keperluan, 
			request_ppds.total_estimasi, 
			request_ppds.status, 
			request_ppds.periode_berangkat,
			users.nama
		`).
		Joins("LEFT JOIN dokumens ON dokumens.doc_ref_id = request_ppds.id AND dokumens.doc_ref_type = 'RequestPPD' AND dokumens.tipe_dokumen = 'Bon Sementara'").
		Joins("INNER JOIN users ON users.id = request_ppds.user_id").
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

func (r *PerjalananDinas) GetListPendingPerjalananDinas(ctx context.Context, jabatan string, userID uint, page int, limit int) ([]model.PPDListView, int64, error) {
	var listData []model.PPDListView
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
			request_ppds.tujuan, 
			request_ppds.keperluan, 
			request_ppds.total_estimasi, 
			request_ppds.status, 
			request_ppds.periode_berangkat,
			users.nama
		`).
		Joins("INNER JOIN users ON users.id = request_ppds.user_id")
	switch jabatan {
	case "Atasan":
		query = query.Where("request_ppds.status ILIKE ? AND users.atasan_id = ?", "Menunggu Atasan", userID)
	case "HRGA":
		query = query.Where("status = ?", "Menunggu HRGA")
	case "Direktur":
		query = query.Where("status = ?", "Menunggu Direktur")
	case "Finance":
		query = query.Where("status = ?", "Menunggu Finance")
	default:
		return []model.PPDListView{}, 0, nil
	}

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

func (r *PerjalananDinas) GetListPPDForRealisasi(ctx context.Context, userID uint) ([]model.DropdownPPDView, error) {
	var list []model.DropdownPPDView

	err := r.db.WithContext(ctx).
		Table("request_ppds").
		Select(`request_ppds.id, 
				dokumens.nomor_tipe_dokumen AS nomor_dokumen, 
				request_ppds.tujuan,
				request_ppds.periode_berangkat,
				request_ppds.periode_kembali,
				request_ppds.total_estimasi`).
		Joins("INNER JOIN dokumens ON dokumens.doc_ref_id = request_ppds.id AND dokumens.doc_ref_type = 'RequestPPD' AND dokumens.tipe_dokumen = 'Bon Sementara'").
		Joins("LEFT JOIN realisasi_bon_sementaras AS rbs ON rbs.request_ppd_id = request_ppds.id").
		Where("request_ppds.user_id = ? AND request_ppds.status = ?", userID, "Selesai").
		Where("rbs.id IS NULL").
		Find(&list).Error

	return list, err
}

func (r *PerjalananDinas) DeclinePerjalananDinas(ctx context.Context, p DeclinePerjalananDinasParams) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.RequestPPD{}).
			Where("id = ? AND status = ?", p.RequestPPDID, p.CurrentStatus).
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

func (r *PerjalananDinas) GetItemsByPPDID(ctx context.Context, ppdID uint, userid uint) ([]model.PPDItemView, error) {
	var items []model.PPDItemView

	query := `
   	SELECT 
        h.id, 
        'Akomodasi: ' || h.nama_hotel AS uraian, 
        1 AS kuantitas,          	
        h.harga_total AS harga_unit, 
		'Akomodasi' AS kategori,
        h.harga_total AS total    
    FROM ppd_hotels h
    INNER JOIN request_ppds r ON h.request_ppd_id = r.id
    WHERE h.request_ppd_id = ? AND r.user_id = ?
    
    UNION ALL
    
    SELECT 
        t.id, 
        'Transportasi: ' || t.jenis_transportasi AS uraian, 
        1 AS kuantitas, 
        t.harga AS harga_unit, 
		'Transportasi' AS kategori,
        t.harga AS total
    FROM ppd_transportasis t
    INNER JOIN request_ppds r ON t.request_ppd_id = r.id
    WHERE t.request_ppd_id = ? AND r.user_id = ?
    
    UNION ALL
    
    SELECT 
        rt.id, 
        rt.keterangan AS uraian, 
        rt.kuantitas AS kuantitas,        
        rt.harga AS harga_unit,   
		rt.kategori AS kategori, 
        (rt.harga * rt.kuantitas) AS total
    FROM ppd_rincian_tambahans rt
    INNER JOIN request_ppds r ON rt.request_ppd_id = r.id
    WHERE rt.request_ppd_id = ? AND r.user_id = ?
    
	`
	err := r.db.WithContext(ctx).Raw(query, ppdID, userid, ppdID, userid, ppdID, userid).Scan(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *PerjalananDinas) GetTotalEstimasi(ctx context.Context, ppdid uint) (int64, error) {
	var totalEstimasi int64

	err := r.db.WithContext(ctx).
		Table("request_ppds").
		Select("total_estimasi").
		Where("id = ?", ppdid).
		Scan(&totalEstimasi).Error
	if err != nil {
		return 0, err
	}

	return totalEstimasi, nil
}

func (r *PerjalananDinas) GetUserIDByPPDID(ctx context.Context, ppdID uint) (uint, error) {
	var userID uint
	err := r.db.WithContext(ctx).
		Model(&model.RequestPPD{}).
		Select("user_id").
		Where("id = ?", ppdID).
		Row().Scan(&userID)

	return userID, err
}

func (r *PerjalananDinas) GetNomorBS(ctx context.Context, ppdID uint) (string, error) {
	var nomorBS string
	err := r.db.WithContext(ctx).
		Model(&model.Dokumen{}).
		Select("nomor_tipe_dokumen").
		Where("doc_ref_id = ? AND doc_ref_type = ? AND tipe_dokumen = ?", ppdID, "RequestPPD", "Bon Sementara").
		Row().Scan(&nomorBS)

	return nomorBS, err
}

func (r *PerjalananDinas) UpdatePengajuanPerjalananDinas(ctx context.Context, data model.RequestPPD) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.RequestPPD{}).Where("id = ?", data.Id).Updates(map[string]interface{}{
			"total_estimasi": data.TotalEstimasi,
		}).Error; err != nil {
			return err
		}

		if err := r.UpdateRincianTambahan(tx, data.Id, &data.RincianTambahan); err != nil {
			return err
		}

		if err := r.UpdateRincianTransportasi(tx, data.Id, data.RincianTransportasi); err != nil {
			return err
		}

		if err := r.UpdateRincianHotel(tx, data.Id, data.RincianHotel); err != nil {
			return err
		}

		return nil
	})
}

func (r *PerjalananDinas) UpdateRincianTambahan(tx *gorm.DB, ppdid uint, items *[]model.PPDRincianTambahan) error {
	for _, item := range *items {
		if item.Id == 0 {
			return nil
		}
		if err := tx.Model(&model.PPDRincianTambahan{}).Where("id = ? AND request_ppd_id = ?", item.Id, ppdid).Updates(map[string]interface{}{
			"harga":     item.Harga,
			"kuantitas": item.Kuantitas,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *PerjalananDinas) UpdateRincianTransportasi(tx *gorm.DB, ppdid uint, items []model.PPDTransportasi) error {
	for _, item := range items {
		if item.Id == 0 {
			return nil
		}
		if err := tx.Model(&model.PPDTransportasi{}).Where("id = ? AND request_ppd_id = ?", item.Id, ppdid).Updates(map[string]interface{}{
			"harga": item.Harga,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *PerjalananDinas) UpdateRincianHotel(tx *gorm.DB, ppdid uint, rincianHotel *model.PPDHotel) error {
	if rincianHotel == nil || rincianHotel.Id == 0 {
		return nil
	}
	return tx.Model(&model.PPDHotel{}).Where("id = ? AND request_ppd_id = ?", rincianHotel.Id, ppdid).Updates(map[string]interface{}{
		"harga_per_malam": rincianHotel.HargaPerMalam,
		"harga_total":     rincianHotel.HargaTotal,
	}).Error
}

func normalizePagination(page, limit int) Pagination {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	return Pagination{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}
