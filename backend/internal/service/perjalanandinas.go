package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang-mmi/internal/constant"
	"golang-mmi/internal/dto"
	"golang-mmi/internal/model"
	"golang-mmi/internal/repository"
	"golang-mmi/internal/utils"
)

var (
	// 403 Forbidden
	ErrAksesditolak      = errors.New("akses ditolak: dokumen bukan milik anda")
	ErrJabatanTidakValid = errors.New("jabatan tidak memiliki izin untuk aksi ini")

	// 404 Not Found
	ErrPPDTidakDitemukan = errors.New("perjalanan dinas tidak ditemukan")

	// 422 Unprocessable
	ErrStatusTidakValid   = errors.New("status perjalanan dinas tidak valid untuk aksi ini")
	ErrKategoriTidakValid = errors.New("kategori rincian tidak valid")
)

var ProsesStatus = map[string][]string{
	constant.StatusDraft: {
		constant.StatusMenungguAtasan,
		constant.StatusMenungguHRGA,
	},
	constant.StatusMenungguAtasan: {
		constant.StatusMenungguHRGA,
		constant.StatusDitolakAtasan,
	},
	constant.StatusMenungguHRGA: {
		constant.StatusMenungguDirektur,
		constant.StatusDitolakHRGA,
	},
	constant.StatusMenungguDirektur: {
		constant.StatusMenungguFinance,
		constant.StatusDitolakDirektur,
	},
	constant.StatusMenungguFinance: {
		constant.StatusSelesai,
		constant.StatusDitolakFinance,
	},
}

var KategoriValid = map[string]bool{
	constant.KategoriTransportasi:  true,
	constant.KategoriKonsumsi:      true,
	constant.KategoriBBM:           true,
	constant.KategoriEntertainment: true,
	constant.KategoriTol:           true,
	constant.KategoriParkir:        true,
}

var statusYangDiharapkan = map[string]string{
	constant.JabatanAtasan:   constant.StatusMenungguAtasan,
	constant.JabatanHRGA:     constant.StatusMenungguHRGA,
	constant.JabatanDirektur: constant.StatusMenungguDirektur,
	constant.JabatanFinance:  constant.StatusMenungguFinance,
}

var statusDownloadable = map[string]bool{
	constant.StatusMenungguDirektur: true,
	constant.StatusMenungguFinance:  true,
	constant.StatusSelesai:          true,
}

type PerjalananDinasImpl struct {
	repo         repository.PerjalananDinasRepository
	servicedoc   DocumentService
	notifManager *utils.NotificationManager
}

func NewPerjalananDinasService(repo repository.PerjalananDinasRepository, servicedoc DocumentService, notifManager *utils.NotificationManager) *PerjalananDinasImpl {
	return &PerjalananDinasImpl{
		repo:         repo,
		servicedoc:   servicedoc,
		notifManager: notifManager,
	}
}

type ServicePPD interface {
	CreatePengajuanPerjalanaDinas(ctx context.Context, req dto.CreatePPDRequest) error
	GetListPerjalananDinas(ctx context.Context, req dto.ListPPDRequest) ([]dto.ListPPDResponse, int64, error)
	DeclinePerjalananDinas(ctx context.Context, req dto.DeclinePPDRequest) error
	ApprovePerjalananDinas(ctx context.Context, req dto.ApprovePPDRequest) error
	GetPerjalananDetail(ctx context.Context, ppdid uint, user uint, jabatan string) (dto.DetailPPDResponse, error)
	GetListPendingPerjalananDinas(ctx context.Context, req dto.ListPPDRequest) ([]dto.ListPPDResponse, int64, error)
	GetItemsByPPDID(ctx context.Context, ppdID uint, userid uint) (dto.PPDItemDetailResponse, error)
	FillPPDPDF(ctx context.Context, ppdID uint, userID uint, templatePath string, jabatan string, w io.Writer) error
	FillBSPDF(ctx context.Context, ppdID uint, userID uint, templatePath string, w io.Writer) error
	EditPerjalananDinas(ctx context.Context, ppdID uint, req dto.UpdatePPDRequest) error
}

func (s *PerjalananDinasImpl) CreatePengajuanPerjalanaDinas(ctx context.Context, req dto.CreatePPDRequest) error {
	var rincianBersih []model.PPDRincianTambahan
	for _, item := range req.RincianTambahan {
		if !KategoriValid[item.Kategori] {
			return ErrKategoriTidakValid
		}
		rincianBersih = append(rincianBersih, item)
	}
	req.RincianTambahan = rincianBersih

	var autoApprovals []model.RiwayatApproval
	var totalHitung int64

	if req.RincianTransportasi != nil {
		for i := range *req.RincianTransportasi {
			(*req.RincianTransportasi)[i].Kategori = constant.KategoriTransportasi
			totalHitung += (*req.RincianTransportasi)[i].Harga
		}
	}

	if req.RincianHotel != nil && req.RincianHotel.NamaHotel != "" {
		req.RincianHotel.Kategori = constant.KategoriAkomodasi
		totalHitung += req.RincianHotel.Harga
	}

	for _, item := range req.RincianTambahan {
		totalHitung += item.Harga * int64(item.Kuantitas)
	}

	var status string
	var targetNotif string
	switch req.Jabatan {
	case constant.JabatanPegawai:
		status = constant.StatusMenungguAtasan
		targetNotif = constant.JabatanAtasan

	case constant.JabatanAtasan:
		status = constant.StatusMenungguHRGA
		targetNotif = constant.JabatanHRGA
		autoApprovals = append(autoApprovals, model.RiwayatApproval{
			UserID:   req.UserID,
			Jabatan:  constant.JabatanAtasan,
			Tindakan: constant.TindakanDisetujui,
			Catatan:  "Auto-approved (diajukan oleh atasan)",
		})

	case constant.JabatanHRGA:
		status = constant.StatusMenungguDirektur
		targetNotif = constant.JabatanDirektur
		autoApprovals = append(autoApprovals, model.RiwayatApproval{
			UserID:   req.UserID,
			Jabatan:  constant.JabatanHRGA,
			Tindakan: constant.TindakanDisetujui,
			Catatan:  "Auto-approved (diajukan oleh HRGA)",
		})

	default:
		return ErrJabatanTidakValid
	}

	newPPD := model.RequestPPD{
		UserID:              req.UserID,
		Tujuan:              req.Tujuan,
		PeriodeBerangkat:    req.TanggalBerangkat,
		Keperluan:           req.Keperluan,
		PeriodeKembali:      req.TanggalKembali,
		UrlDokumen:          req.UrlDokumen,
		RincianTambahan:     req.RincianTambahan,
		RincianTransportasi: req.RincianTransportasi,
		TotalEstimasi:       totalHitung,
		Status:              status,
		RiwayatPersetujuan:  autoApprovals,
	}

	if err := s.repo.CreatePengajuanPerjalanaDinas(ctx, &newPPD); err != nil {
		return fmt.Errorf("gagal membuat pengajuan perjalanan dinas: %w", err)
	}

	if req.Jabatan == constant.JabatanHRGA {
		nomorUmum, errGen := s.servicedoc.GenerateNomorDokumenGeneral(ctx, constant.KodeDeptHRD)
		if errGen != nil {
			return fmt.Errorf("gagal generate nomor dokumen: %w", errGen)
		}

		newDokumen := model.Dokumen{
			DocRefID:     newPPD.Id,
			DocRefType:   constant.DocRefTypePPD,
			UserID:       req.UserID,
			NomorDokumen: nomorUmum,
			TipeDokumen:  constant.TipeDokumenPPD,
		}

		if err := s.servicedoc.SaveDokumen(ctx, &newDokumen); err != nil {
			return fmt.Errorf("gagal menyimpan dokumen: %w", err)
		}
	}

	if targetNotif != "" {
		pesan := fmt.Sprintf(
			"Ada pengajuan Perjalanan Dinas baru menunggu persetujuan (Tujuan: %s)",
			req.Tujuan,
		)
		go s.notifManager.SendToRole(targetNotif, pesan)
	}

	return nil
}

func (s *PerjalananDinasImpl) GetListPerjalananDinas(ctx context.Context, req dto.ListPPDRequest) ([]dto.ListPPDResponse, int64, error) {
	var dataModels []model.PPDListView
	var totalData int64
	var err error

	switch req.Jabatan {
	case constant.JabatanHRGA, constant.JabatanDirektur, constant.JabatanFinance:
		dataModels, totalData, err = s.repo.GetListRiwayatPerjalananDinas(ctx, req.Page, req.Limit)
	case constant.JabatanPegawai:
		dataModels, totalData, err = s.repo.GetListRiwayatPerjalananDinasByUserID(ctx, req.UserID, req.Page, req.Limit)
	case constant.JabatanAtasan:
		dataModels, totalData, err = s.repo.GetListRiwayatPerjalananDinasByAtasan(ctx, req.UserID, req.Page, req.Limit)
	default:
		return nil, 0, ErrAksesditolak
	}

	if err != nil {
		return nil, 0, err
	}

	response := make([]dto.ListPPDResponse, 0)
	for _, item := range dataModels {

		response = append(response, dto.ListPPDResponse{
			ID:               item.ID,
			NomorTipeDokumen: item.NomorTipeDokumen,
			NomorDokumen:     item.NomorDokumen,
			Nama:             item.Nama,
			Tujuan:           item.Tujuan,
			Keperluan:        item.Keperluan,
			TotalEstimasi:    int(item.TotalEstimasi),
			Status:           item.Status,
			IsDownloadable:   statusDownloadable[item.Status],
			PeriodeBerangkat: item.PeriodeBerangkat,
		})
	}

	return response, totalData, nil
}

func (s *PerjalananDinasImpl) DeclinePerjalananDinas(ctx context.Context, req dto.DeclinePPDRequest) error {
	currentStatus, err := s.repo.GetStatusPerjalananDinas(ctx, req.RequestPPDID)
	if err != nil {
		return ErrPPDTidakDitemukan
	}

	if err := validateJabatanDanStatus(req.Jabatan, currentStatus); err != nil {
		return err
	}

	OpsiStatus, exist := ProsesStatus[currentStatus]

	if !exist || len(OpsiStatus) < 2 {
		return ErrStatusTidakValid
	}

	nextStatus := OpsiStatus[1]

	riwayat := &model.RiwayatApproval{
		DocRefID:   req.RequestPPDID,
		DocRefType: constant.DocRefTypePPD,
		UserID:     req.UserID,
		Jabatan:    req.Jabatan,
		Tindakan:   constant.TindakanDitolak,
		Catatan:    req.Catatan,
	}

	params := repository.DeclinePerjalananDinasParams{
		RequestPPDID: req.RequestPPDID,
		NextStatus:   nextStatus,
		Riwayat:      riwayat,
	}

	go func() {
		s.notifManager.SendToRole(constant.JabatanPegawai, constant.NotifDitolakPPD)
	}()

	return s.repo.DeclinePerjalananDinas(ctx, params)
}

func (s *PerjalananDinasImpl) GetListPendingPerjalananDinas(ctx context.Context, req dto.ListPPDRequest) ([]dto.ListPPDResponse, int64, error) {
	var dataModels []model.PPDListView
	var err error
	var totalData int64

	if req.Jabatan == constant.JabatanPegawai {
		return nil, 0, ErrAksesditolak
	}

	dataModels, totalData, err = s.repo.GetListPendingPerjalananDinas(ctx, req.Jabatan, req.UserID, req.Page, req.Limit)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil daftar pending: %w", err)
	}

	response := make([]dto.ListPPDResponse, 0)
	for _, item := range dataModels {
		response = append(response, dto.ListPPDResponse{
			ID:               item.ID,
			NomorTipeDokumen: item.NomorTipeDokumen,
			Nama:             item.Nama,
			Tujuan:           item.Tujuan,
			Keperluan:        item.Keperluan,
			TotalEstimasi:    int(item.TotalEstimasi),
			Status:           item.Status,
			PeriodeBerangkat: item.PeriodeBerangkat,
		})
	}

	return response, totalData, nil
}

func (s *PerjalananDinasImpl) GetPerjalananDetail(ctx context.Context, ppdid uint, userID uint, jabatan string) (dto.DetailPPDResponse, error) {
	data, err := s.repo.GetDetailPerjalananDinas(ctx, ppdid)
	if err != nil {
		return dto.DetailPPDResponse{}, fmt.Errorf("gagal mengambil detail: %w", err)
	}

	if jabatan == constant.JabatanPegawai && data.UserID != userID {
		return dto.DetailPPDResponse{}, ErrAksesditolak
	}
	resp := dto.DetailPPDResponse{
		ID:               data.Id,
		Tujuan:           data.Tujuan,
		PeriodeBerangkat: data.PeriodeBerangkat,
		PeriodeKembali:   data.PeriodeKembali,
		Keperluan:        data.Keperluan,
		Status:           data.Status,
		TotalEstimasi:    data.TotalEstimasi,
		UrlDokumen:       data.UrlDokumen,
	}

	for _, item := range data.RincianTambahan {
		resp.RincianTambahan = append(resp.RincianTambahan, dto.RincianTambahanResponse{
			Harga:      item.Harga,
			Kuantitas:  item.Kuantitas,
			Keterangan: item.Keterangan,
			Kategori:   item.Kategori,
		})
	}

	if data.RincianTransportasi != nil {
		for _, item := range *data.RincianTransportasi {
			resp.RincianTransportasi = append(resp.RincianTransportasi, dto.RincianTransportResponse{
				TipePerjalanan:    item.TipePerjalanan,
				KotaAsal:          item.KotaAsal,
				KotaTujuan:        item.KotaTujuan,
				JenisTransportasi: item.JenisTransportasi,
				NomorKendaraan:    item.NomorKendaraan,
				Harga:             item.Harga,
				Kategori:          item.Kategori,
				JamBerangkat:      item.Jamberangkat,
			})
		}

	}

	if data.RincianHotel != nil {
		resp.RincianHotel = &dto.RincianHotelResponse{
			NamaHotel:  data.RincianHotel.NamaHotel,
			TotalHarga: data.RincianHotel.Harga,
		}
	}

	for _, item := range data.RiwayatPersetujuan {
		resp.RiwayatApproval = append(resp.RiwayatApproval, dto.RiwayatPersetujuanResponse{
			Nama:           item.User.Nama,
			Tindakan:       item.Tindakan,
			Jabatan:        item.Jabatan,
			Catatan:        item.Catatan,
			WaktuDisetujui: item.CreatedAt,
		})
	}

	return resp, nil
}

func (s *PerjalananDinasImpl) ApprovePerjalananDinas(ctx context.Context, req dto.ApprovePPDRequest) error {
	currentStatus, err := s.repo.GetStatusPerjalananDinas(ctx, req.RequestPPDID)
	if err != nil {
		return ErrPPDTidakDitemukan
	}

	if err := validateJabatanDanStatus(req.Jabatan, currentStatus); err != nil {
		return err
	}

	OpsiStatus, exist := ProsesStatus[currentStatus]
	if !exist || len(OpsiStatus) == 0 {
		return ErrStatusTidakValid
	}

	nextStatus := OpsiStatus[0]

	pengajuID, err := s.repo.GetUserIDByPPDID(ctx, req.RequestPPDID)
	if err != nil {
		return ErrPPDTidakDitemukan
	}

	var newdokumen []model.Dokumen

	if req.Jabatan == constant.JabatanHRGA {
		nomorUmum, errGen := s.servicedoc.GenerateNomorDokumenGeneral(ctx, "HRD")
		if errGen != nil {
			return fmt.Errorf("gagal generate nomor dokumen: %w", errGen)
		}

		newdokumen = append(newdokumen, model.Dokumen{
			DocRefID:     req.RequestPPDID,
			DocRefType:   constant.DocRefTypePPD,
			UserID:       pengajuID,
			NomorDokumen: nomorUmum,
			TipeDokumen:  constant.TipeDokumenPPD,
		})
	}

	if req.Jabatan == constant.JabatanFinance {
		nomorUmum, errGen := s.servicedoc.GenerateNomorDokumenGeneral(ctx, constant.KodeDeptFinance)
		if errGen != nil {
			return fmt.Errorf("gagal generate nomor dokumen umum: %w", errGen)
		}

		nomorSpesifik, errSpec := s.servicedoc.GenerateNomorDokumenSpecific(ctx, constant.TipeDokumenBS, constant.KodePrefixBS)
		if errSpec != nil {
			return fmt.Errorf("gagal generate nomor dokumen spesifik: %w", errSpec)
		}

		newdokumen = append(newdokumen, model.Dokumen{
			DocRefID:         req.RequestPPDID,
			DocRefType:       constant.DocRefTypePPD,
			UserID:           pengajuID,
			NomorDokumen:     nomorUmum,
			NomorTipeDokumen: nomorSpesifik,
			TipeDokumen:      constant.TipeDokumenBS,
		})

	}

	riwayat := &model.RiwayatApproval{
		DocRefID:   req.RequestPPDID,
		DocRefType: constant.DocRefTypePPD,
		UserID:     req.UserID,
		Jabatan:    req.Jabatan,
		Tindakan:   constant.TindakanDisetujui,
		Catatan:    req.Catatan,
	}

	params := repository.ApprovePerjalananDinasparams{
		RequestPPDID: req.RequestPPDID,
		NextStatus:   nextStatus,
		NewDokumen:   newdokumen,
		Riwayat:      riwayat,
	}

	if err := s.repo.ApprovePerjalananDinas(ctx, params); err != nil {
		return fmt.Errorf("gagal menyimpan persetujuan perjalanan dinas: %w", err)
	}

	go func() {
		switch req.Jabatan {
		case constant.JabatanAtasan:
			s.notifManager.SendToRole(constant.JabatanHRGA, constant.NotifAtasanSetujuPPD)
		case constant.JabatanHRGA:
			s.notifManager.SendToRole(constant.JabatanDirektur, constant.NotifHRGASetujuPPD)
		case constant.JabatanDirektur:
			s.notifManager.SendToRole(constant.JabatanFinance, constant.NotifDirekturSetujuPPD)
		case constant.JabatanFinance:
			s.notifManager.SendToRole(constant.JabatanPegawai, constant.NotifFinanceSetujuPPD)
		}
	}()

	return nil
}

func (s *PerjalananDinasImpl) GetItemsByPPDID(ctx context.Context, ppdID uint, userid uint) (dto.PPDItemDetailResponse, error) {
	items, err := s.repo.GetItemsByPPDID(ctx, ppdID, userid)
	if err != nil {

		return dto.PPDItemDetailResponse{}, fmt.Errorf("gagal mengambil rincian item PPD: %w", err)
	}

	nomorReferensi, err := s.repo.GetNomorBS(ctx, ppdID)
	if err != nil {
		return dto.PPDItemDetailResponse{}, fmt.Errorf("gagal mengambil nomor referensi: %w", err)
	}

	itemResponses := make([]dto.PPDItemResponse, 0, len(items))
	for _, item := range items {
		itemResponses = append(itemResponses, dto.PPDItemResponse{
			ID:        item.ID,
			Uraian:    item.Uraian,
			Qty:       item.Kuantitas,
			HargaUnit: item.HargaUnit,
			Kategori:  item.Kategori,
			Total:     item.Total,
		})
	}

	response := dto.PPDItemDetailResponse{
		NomorReferensi: nomorReferensi,
		Items:          itemResponses,
	}

	return response, nil
}

func (s *PerjalananDinasImpl) GetDataPPDForPDF(ctx context.Context, ppdID uint, userid uint, jabatan string) (dto.PPDDataToPDF, error) {
	data, err := s.repo.GetDetailPerjalananDinas(ctx, ppdID)
	if err != nil {
		return dto.PPDDataToPDF{}, fmt.Errorf("gagal mengambil detail: %w", err)
	}

	if jabatan == constant.JabatanPegawai && data.UserID != userid {
		return dto.PPDDataToPDF{}, ErrAksesditolak
	}

	if jabatan == constant.JabatanAtasan {
		isOwner := data.UserID == userid
		isBawahan := data.User.AtasanID != nil && *data.User.AtasanID == userid

		if !isOwner && !isBawahan {
			return dto.PPDDataToPDF{}, ErrAksesditolak
		}
	}

	var nomorDokumen string
	var tanggalTerbit string
	docPPD := findDokumenByTipe(data.Dokumen, constant.TipeDokumenPPD)
	if docPPD != nil {
		nomorDokumen = docPPD.NomorDokumen
		tanggalTerbit = utils.FormatTanggal(docPPD.CreatedAt)
	}

	riwayats := findTandaTanganByJabatan(data.RiwayatPersetujuan)

	pathAtasan := ""
	tanggalDisetujuiAtasan := ""
	if r, ok := riwayats[constant.JabatanAtasan]; ok {
		pathAtasan = r.User.PathTandaTangan
		tanggalDisetujuiAtasan = utils.FormatTanggal(r.CreatedAt)
	}

	pathHRGA := ""
	tanggalDisetujuiHRGA := ""
	if r, ok := riwayats[constant.JabatanHRGA]; ok {
		pathHRGA = r.User.PathTandaTangan
		tanggalDisetujuiHRGA = utils.FormatTanggal(r.CreatedAt)
	}

	pathDirektur := ""
	tanggalDisetujuiDirektur := ""
	if r, ok := riwayats[constant.JabatanDirektur]; ok {
		pathDirektur = r.User.PathTandaTangan
		tanggalDisetujuiDirektur = utils.FormatTanggal(r.CreatedAt)
	}

	pdf := dto.PPDDataToPDF{
		NomorDokumen:             nomorDokumen,
		TanggalTerbit:            tanggalTerbit,
		Nama:                     data.User.Nama,
		Jabatan:                  data.User.Jabatan,
		Wilayah:                  data.User.Wilayah,
		Nik:                      data.User.Nik,
		Tujuan:                   data.Tujuan,
		Keperluan:                data.Keperluan,
		Periode:                  fmt.Sprintf("%s - %s", utils.FormatTanggal(data.PeriodeBerangkat), utils.FormatTanggal(data.PeriodeKembali)),
		PathTandaTanganPengaju:   data.User.PathTandaTangan,
		PathTandaTanganAtasan:    pathAtasan,
		PathTandaTanganHRGA:      pathHRGA,
		PathTandaTanganDirektur:  pathDirektur,
		TanggalDisetujuiAtasan:   tanggalDisetujuiAtasan,
		TanggalDisetujuiHRGA:     tanggalDisetujuiHRGA,
		TanggalDisetujuiDirektur: tanggalDisetujuiDirektur,
		TanggalDiajukan:          utils.FormatTanggal(data.CreatedAt),
	}

	if data.RincianHotel != nil && data.RincianHotel.NamaHotel != "" {
		pdf.NamaHotel = data.RincianHotel.NamaHotel
		pdf.CheckIn = utils.FormatTanggal(data.RincianHotel.CheckIn)
		pdf.CheckOut = utils.FormatTanggal(data.RincianHotel.CheckOut)
		pdf.HargaHotel = utils.FormatRupiah(data.RincianHotel.Harga)
		pdf.PeriodeHotel = pdf.Periode
		pdf.TujuanHotel = data.Tujuan
	}

	if data.RincianTransportasi != nil {
		for _, t := range *data.RincianTransportasi {
			noKendaraan := "-"
			if t.NomorKendaraan != nil {
				noKendaraan = *t.NomorKendaraan
			}

			if t.TipePerjalanan == "Keberangkatan" {
				pdf.AsalKeberangkatan = t.KotaAsal
				pdf.TujuanKeberangkatan = t.KotaTujuan
				pdf.JenisTransportasiKeberangkatan = t.JenisTransportasi
				pdf.NomorKendaraanKeberangkatan = noKendaraan
				pdf.JamBerangkatKeberangkatan = t.Jamberangkat.Format("15:04")
			} else if t.TipePerjalanan == "Kedatangan" {
				pdf.AsalKedatangan = t.KotaAsal
				pdf.TujuanKedatangan = t.KotaTujuan
				pdf.JenisTransportasiKedatangan = t.JenisTransportasi
				pdf.NomorKendaraanKedatangan = noKendaraan
				pdf.JamBerangkatKedatangan = t.Jamberangkat.Format("15:04")
			}
		}

	}

	for _, r := range data.RincianTambahan {
		total := r.Harga * int64(r.Kuantitas)
		pdf.Rincian = append(pdf.Rincian, dto.PPDDataToPDFDetail{
			Kategori: r.Kategori,
			Harga:    utils.FormatNominal(r.Harga),
			Total:    utils.FormatNominal(total),
		})
	}

	return pdf, nil
}

func (s *PerjalananDinasImpl) FillPPDPDF(ctx context.Context, ppdID uint, userID uint, templatePath string, jabatan string, w io.Writer) error {
	pdfData, err := s.GetDataPPDForPDF(ctx, ppdID, userID, jabatan)
	if err != nil {
		return fmt.Errorf("gagal mengambil data untuk PDF: %w", err)
	}

	formData := map[string]string{
		"no_dokumen":                  pdfData.NomorDokumen,
		"tanggal_terbit":              pdfData.TanggalTerbit,
		"nama":                        pdfData.Nama,
		"jabatan":                     pdfData.Jabatan,
		"wilayah":                     pdfData.Wilayah,
		"nik":                         pdfData.Nik,
		"tujuan":                      pdfData.Tujuan,
		"keperluan":                   pdfData.Keperluan,
		"periode":                     pdfData.Periode,
		"nama_hotel":                  pdfData.NamaHotel,
		"check_in":                    pdfData.CheckIn,
		"check_out":                   pdfData.CheckOut,
		"harga_hotel":                 pdfData.HargaHotel,
		"periode_hotel":               pdfData.PeriodeHotel,
		"tujuan_hotel":                pdfData.TujuanHotel,
		"asal_keberangkatan":          pdfData.AsalKeberangkatan,
		"tujuan_keberangkatan":        pdfData.TujuanKeberangkatan,
		"jam_berangkat_keberangkatan": pdfData.JamBerangkatKeberangkatan,
		"asal_kedatangan":             pdfData.AsalKedatangan,
		"tujuan_kedatangan":           pdfData.TujuanKedatangan,
		"jam_berangkat_kedatangan":    pdfData.JamBerangkatKedatangan,
	}

	for i := 1; i <= 8; i++ {
		formData[fmt.Sprintf("checkbox_%d", i)] = ""
	}
	formData["nomor_kendaraan_keberangkatan"] = ""
	formData["jenis_transportasi_lain_keberangkatan"] = ""
	formData["nomor_kendaraan_kedatangan"] = ""
	formData["jenis_transportasi_lain_kedatangan"] = ""

	transBerangkat := pdfData.JenisTransportasiKeberangkatan
	if transBerangkat == constant.TransportasiPesawat {
		formData["checkbox_1"] = "X"
	} else if transBerangkat == constant.TransportasiKeretaApi {
		formData["checkbox_2"] = "X"
	} else if transBerangkat == constant.TransportasiMobilDinas {
		formData["checkbox_4"] = "X"
		formData["nomor_kendaraan_keberangkatan"] = pdfData.NomorKendaraanKeberangkatan
	} else if transBerangkat != "" {
		formData["checkbox_3"] = "X"
		formData["jenis_transportasi_lain_keberangkatan"] = transBerangkat
	}

	transDatang := pdfData.JenisTransportasiKedatangan
	if transDatang == constant.TransportasiPesawat {
		formData["checkbox_5"] = "X"
	} else if transDatang == constant.TransportasiKeretaApi {
		formData["checkbox_6"] = "X"
	} else if transDatang == constant.TransportasiMobilDinas {
		formData["checkbox_8"] = "X"
		formData["nomor_kendaraan_kedatangan"] = pdfData.NomorKendaraanKedatangan
	} else if transDatang != "" {
		formData["checkbox_7"] = "X"
		formData["jenis_transportasi_lain_kedatangan"] = transDatang
	}

	var totalSeluruh int64
	var totalLainLain int64
	var hasLainLain bool

	for i, rincian := range pdfData.Rincian {
		formData[fmt.Sprintf("keterangan_%d", i+1)] = rincian.Kategori
		formData[fmt.Sprintf("harga_%d", i+1)] = rincian.Harga
		formData[fmt.Sprintf("total%d", i+1)] = rincian.Total

		cleanTotal := strings.ReplaceAll(rincian.Total, "Rp", "")
		cleanTotal = strings.ReplaceAll(cleanTotal, ".", "")
		cleanTotal = strings.ReplaceAll(cleanTotal, " ", "")
		cleanTotal = strings.ReplaceAll(cleanTotal, ",", "")

		totalInt, _ := strconv.ParseInt(cleanTotal, 10, 64)
		totalSeluruh += totalInt

		if i < 9 {
			formData[fmt.Sprintf("keterangan_%d", i+1)] = rincian.Kategori
			formData[fmt.Sprintf("harga_%d", i+1)] = rincian.Harga

			cleanVal := strings.ReplaceAll(rincian.Total, "Rp", "")
			formData[fmt.Sprintf("total%d", i+1)] = strings.TrimSpace(cleanVal)
		} else {
			hasLainLain = true
			totalLainLain += totalInt
		}

		if hasLainLain {
			formData["keterangan_10"] = "Lain-lain"
			formData["harga_10"] = utils.FormatNominal(totalLainLain)
			formData["total10"] = utils.FormatNominal(totalLainLain)
		}
	}
	formData["text_52irsw"] = utils.FormatNominal(totalSeluruh)

	if len(formData) == 0 {
		return fmt.Errorf("formData kosong, tidak ada data untuk diisi ke PDF")
	}

	var signatures []utils.SignatureConfig

	if pdfData.PathTandaTanganPengaju != "" {
		signatures = append(signatures, utils.SignatureConfig{
			Path:   strings.TrimLeft(pdfData.PathTandaTanganPengaju, "/"),
			Offset: "20 220",
			Scale:  0.2,
		})
		formData["tanggal_diajukan_karyawan"] = pdfData.TanggalDiajukan
	}

	if pdfData.PathTandaTanganAtasan != "" {
		signatures = append(signatures, utils.SignatureConfig{
			Path:   strings.TrimLeft(pdfData.PathTandaTanganAtasan, "/"),
			Offset: "170 220",
			Scale:  0.2,
		})
		formData["tanggal_disetujui_atasan"] = pdfData.TanggalDisetujuiAtasan
	}

	if pdfData.PathTandaTanganHRGA != "" {
		signatures = append(signatures, utils.SignatureConfig{
			Path:   strings.TrimLeft(pdfData.PathTandaTanganHRGA, "/"),
			Offset: "300 220",
			Scale:  0.2,
		})
		formData["tanggal_disetujui_hrga"] = pdfData.TanggalDisetujuiHRGA
	}

	if pdfData.PathTandaTanganDirektur != "" {
		signatures = append(signatures, utils.SignatureConfig{
			Path:   strings.TrimLeft(pdfData.PathTandaTanganDirektur, "/"),
			Offset: "460 220",
			Scale:  0.2,
		})
		formData["tanggal_disetujui_direktur"] = pdfData.TanggalDisetujuiDirektur
	}

	return utils.GenerateFormPDF(templatePath, formData, signatures, w)
}

func (s *PerjalananDinasImpl) GetDataBSForPDF(ctx context.Context, ppdID uint, userID uint) (dto.BSDataToPDF, error) {
	data, err := s.repo.GetDetailPerjalananDinas(ctx, ppdID)
	if err != nil {
		return dto.BSDataToPDF{}, fmt.Errorf("gagal mengambil detail: %w", err)
	}

	var nomorDokumen string
	var nomorTipeDokumen string
	var tanggalTerbit string

	docBS := findDokumenByTipe(data.Dokumen, constant.TipeDokumenBS)
	if docBS != nil {
		nomorDokumen = docBS.NomorDokumen
		nomorTipeDokumen = docBS.NomorTipeDokumen
		tanggalTerbit = utils.FormatTanggal(docBS.CreatedAt)
	}

	var pathHRGA string
	var namaHRGA string
	var pathDirektur string
	var namaDirektur string
	var pathFinance string
	var namaFinance string
	var tanggalPenyelesaian string

	riwayats := findTandaTanganByJabatan(data.RiwayatPersetujuan)

	pathHRGA = ""
	if r, ok := riwayats[constant.JabatanHRGA]; ok {
		pathHRGA = r.User.PathTandaTangan
		namaHRGA = r.User.Nama
	}

	pathDirektur = ""
	if r, ok := riwayats[constant.JabatanDirektur]; ok {
		pathDirektur = r.User.PathTandaTangan
		namaDirektur = r.User.Nama
	}

	pathFinance = ""
	if r, ok := riwayats[constant.JabatanFinance]; ok {
		pathFinance = r.User.PathTandaTangan
		namaFinance = r.User.Nama
		tanggalPenyelesaian = utils.FormatTanggal(r.CreatedAt)
	}

	pdf := dto.BSDataToPDF{
		NomorDokumen:            nomorDokumen,
		NomorTipeDokumen:        nomorTipeDokumen,
		TanggalTerbit:           tanggalTerbit,
		TanggalPengajuan:        utils.FormatTanggal(data.CreatedAt),
		TanggalPenyelesaian:     tanggalPenyelesaian,
		Jumlah:                  utils.FormatRupiah(data.TotalEstimasi),
		Keperluan:               data.Keperluan,
		PathTandaTanganHRGA:     pathHRGA,
		NamaHRGA:                namaHRGA,
		PathTandaTanganDirektur: pathDirektur,
		PathTandaTanganFinance:  pathFinance,
		NamaFinance:             namaFinance,
		NamaDirektur:            namaDirektur,
	}

	return pdf, nil

}

func (s *PerjalananDinasImpl) FillBSPDF(ctx context.Context, ppdID uint, userID uint, templatePath string, w io.Writer) error {
	pdfData, err := s.GetDataBSForPDF(ctx, ppdID, userID)
	if err != nil {
		return fmt.Errorf("gagal mengambil data untuk PDF: %w", err)
	}

	formData := map[string]string{
		"no_dokumen":           pdfData.NomorDokumen,
		"nama_pengaju":         pdfData.NamaHRGA,
		"no_bon_sementara":     pdfData.NomorTipeDokumen,
		"tanggal_terbit":       pdfData.TanggalTerbit,
		"tanggal_pengajuan":    pdfData.TanggalPengajuan,
		"tanggal_penyelesaian": pdfData.TanggalPenyelesaian,
		"jumlah":               pdfData.Jumlah,
		"keperluan":            pdfData.Keperluan,
		"nama_hrga":            pdfData.NamaHRGA,
		"nama_direktur":        pdfData.NamaDirektur,
		"nama_finance":         pdfData.NamaFinance,
	}

	var signatures []utils.SignatureConfig

	if pdfData.PathTandaTanganHRGA != "" {
		signatures = append(signatures, utils.SignatureConfig{
			Path:   strings.TrimLeft(pdfData.PathTandaTanganHRGA, "/"),
			Offset: "50 410",
			Scale:  0.2,
		})
	}
	if pdfData.PathTandaTanganDirektur != "" {
		signatures = append(signatures, utils.SignatureConfig{
			Path:   strings.TrimLeft(pdfData.PathTandaTanganDirektur, "/"),
			Offset: "430 410",
			Scale:  0.2,
		})
	}
	if pdfData.PathTandaTanganFinance != "" {
		signatures = append(signatures, utils.SignatureConfig{
			Path:   strings.TrimLeft(pdfData.PathTandaTanganFinance, "/"),
			Offset: "240 410",
			Scale:  0.2,
		})
	}

	return utils.GenerateFormPDF(templatePath, formData, signatures, w)
}

func validateJabatanDanStatus(jabatan, currentStatus string) error {
	expected, ok := statusYangDiharapkan[jabatan]
	if !ok {
		return fmt.Errorf("jabatan tidak dikenali: %s", jabatan)
	}
	if currentStatus != expected {
		return errors.New("tidak dapat memproses perjalanan dinas dengan status saat ini")
	}
	return nil
}

func findDokumenByTipe(dokumens []model.Dokumen, tipe string) *model.Dokumen {
	for _, doc := range dokumens {
		if doc.TipeDokumen == tipe {
			return &doc
		}
	}
	return nil
}

func findTandaTanganByJabatan(riwayats []model.RiwayatApproval) map[string]model.RiwayatApproval {
	hasil := make(map[string]model.RiwayatApproval)
	for _, r := range riwayats {
		if strings.Contains(r.Tindakan, "Disetujui") {
			hasil[r.Jabatan] = r
		}
	}
	return hasil
}

func (s *PerjalananDinasImpl) EditPerjalananDinas(ctx context.Context, ppdID uint, req dto.UpdatePPDRequest) error {
	var totalHitung int64

	for _, tambahan := range req.RincianTambahan {
		totalHitung += int64(tambahan.Harga) * int64(tambahan.Kuantitas)
	}

	if req.RincianTransportasi != nil {
		for _, transport := range *req.RincianTransportasi {
			totalHitung += int64(transport.Harga)
		}
	}
	if req.RincianHotel != nil {
		totalHitung += int64(req.RincianHotel.Harga)
	}

	updateData := model.RequestPPD{
		Id:                  ppdID,
		TotalEstimasi:       totalHitung,
		RincianTambahan:     req.RincianTambahan,
		RincianTransportasi: req.RincianTransportasi,
		RincianHotel:        req.RincianHotel,
	}

	err := s.repo.UpdatePengajuanPerjalananDinas(ctx, updateData)
	if err != nil {
		return fmt.Errorf("gagal mengedit perjalanan dinas: %w", err)
	}

	return nil
}
