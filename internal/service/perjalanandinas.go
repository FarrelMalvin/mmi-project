package service

import (
	"context"
	"errors"
	"fmt"
	"golang-mmi/internal/config"
	"golang-mmi/internal/model"
	"golang-mmi/internal/repository"
	"strconv"
	"strings"
	"time"
)

var ProsesStatus = map[string][]string{
	"Draft": {"Menunggu Atasan", "Menunggu HRGA"},

	"Menunggu Atasan":   {"Menunggu HRGA", "Ditolak Atasan"},
	"Menunggu HRGA":     {"Menunggu Direktur", "Ditolak HRGA"},
	"Menunggu Direktur": {"Menunggu Finance", "Ditolak Direktur"},
	"Menunggu Finance":  {"Selesai", "Ditolak Finance"},
}

type PerjalananDinasImpl struct {
	repo repository.PerjalananDinasRepository
}

func NewPerjalananDinasService(repo repository.PerjalananDinasRepository) *PerjalananDinasImpl {
	return &PerjalananDinasImpl{
		repo: repo,
	}
}

type ServicePPD interface {
	CreatePengajuanPerjalanaDinas(ctx context.Context, req *model.RequestPPD) error
	GetListPerjalananDinas(ctx context.Context, page, limit int) ([]repository.RiwayatPPDResponse, int64, error)
	DeclinePerjalananDinas(ctx context.Context, ppdid uint, catatan string) error
	ApprovePerjalananDinas(ctx context.Context, ppdid uint, catatan string) error
	GenerateNomorDokumenGeneral(ctx context.Context, prefix string) (string, error)
	GetDropdownPPD(ctx context.Context) ([]repository.DropdownPPDResponse, error)
	GenerateNomorDokumenSpecific(ctx context.Context, tipe string, prefix string) (string, error)
	GetListPerjalananDetail(ctx context.Context, ppdid uint) ([]model.PPDRincianTambahan, error)
	GetListPendingPerjalananDinas(ctx context.Context) ([]model.RequestPPD, error)
}

func (s *PerjalananDinasImpl) CreatePengajuanPerjalanaDinas(ctx context.Context, req *model.RequestPPD) error {
	UserID, okID := ctx.Value("user_id").(uint)
	if !okID {
		return errors.New("invalid user ID in context")
	}
	jabatan, okjabatan := ctx.Value("jabatan").(string)
	if !okjabatan {
		return errors.New("invalid jabatan in context")
	}

	var rincianBersih []model.PPDRincianTambahan
	for _, item := range req.RincianTambahan {
		if item.Kategori == "Transportasi" || item.Kategori == "Konsumsi" || item.Kategori == "BBM" || item.Kategori == "Entertaiment" || item.Kategori == "Tol" || item.Kategori == "Parkir" {
			rincianBersih = append(rincianBersih, item)
		} else {
			return fmt.Errorf("kategori rincian tidak valid: %s", item.Kategori)
		}
	}

	req.RincianTambahan = rincianBersih

	var totalHitunganBackend int64 = 0

	if req.RincianTransportasi != nil && req.RincianTransportasi.JenisTransportasi != "" {
		req.RincianTransportasi.Kategori = "Transportasi"
		totalHitunganBackend += req.RincianTransportasi.Harga
	}

	if req.RincianHotel != nil && req.RincianHotel.NamaHotel != "" {
		req.RincianHotel.Kategori = "Akomodasi"
		totalHitunganBackend += req.RincianHotel.Harga
	}

	for _, item := range req.RincianTambahan {
		totalHitunganBackend += (item.Harga * int64(item.Kuantitas))
	}

	req.TotalEstimasi = totalHitunganBackend
	req.UserID = UserID

	switch jabatan {
	case "Pegawai":
		req.Status = "Menunggu Atasan"
	case "Atasan":
		req.Status = "Menunggu HRGA"
	case "HRGA":
		req.Status = "Menunggu Direktur"
	default:
		return errors.New("invalid jabatan untuk membuat pengajuan")
	}

	return s.repo.CreatePengajuanPerjalanaDinas(ctx, req)
}

func (s *PerjalananDinasImpl) GetListPerjalananDinas(ctx context.Context, page, limit int) ([]repository.RiwayatPPDResponse, int64, error) {

	claims, ok := config.GetClaimsFromContext(ctx)
	if !ok {
		return nil, 0, errors.New("unauthorized: gagal mengambil data profil dari token")
	}

	userID := claims.UserID
	jabatan := claims.Jabatan

	switch jabatan {
	case "HRGA", "Direktur", "Finance":
		return s.repo.GetListRiwayatPerjalananDinas(ctx, page, limit)
	case "Pegawai":
		return s.repo.GetListRiwayatPerjalananDinasByUserID(ctx, userID, page, limit)
	case "Atasan":
		return s.repo.GetListRiwayatPerjalananDinasByAtasan(ctx, userID, page, limit)
	}
	return nil, 0, errors.New("akses ditolak: jabatan tidak memiliki izin untuk melihat daftar perjalanan dinas")
}

func (s *PerjalananDinasImpl) DeclinePerjalananDinas(ctx context.Context, ppdid uint, catatan string) error {

	claims, ok := config.GetClaimsFromContext(ctx)
	if !ok {
		return errors.New("unauthorized: gagal mengambil data profil dari token")
	}

	userID := claims.UserID
	jabatan := claims.Jabatan

	currentStatus, err := s.repo.GetStatusPerjalananDinas(ctx, ppdid)
	if err != nil {
		return errors.New("perjalanan dinas tidak ditemukan")
	}

	if (jabatan == "Atasan" && currentStatus != "Menunggu Atasan") ||
		(jabatan == "HRGA" && currentStatus != "Menunggu HRGA") ||
		(jabatan == "Direktur" && currentStatus != "Menunggu Direktur") ||
		(jabatan == "Finance" && currentStatus != "Menunggu Finance") {
		return errors.New("tidak dapat menolak perjalanan dinas dengan status saat ini")
	}

	OpsiStatus, exist := ProsesStatus[currentStatus]

	if !exist || len(OpsiStatus) < 2 {
		return errors.New("status perjalanan dinas tidak valid untuk ditolak")
	}

	nextStatus := OpsiStatus[1]

	riwayat := &model.RiwayatApproval{
		RequestPPDID: ppdid,
		UserID:       userID,
		Jabatan:      jabatan,
		Tindakan:     "Ditolak",
		Catatan:      catatan,
	}

	params := repository.DeclinePerjalananDinasParams{
		RequestPPDID: ppdid,
		NextStatus:   nextStatus,
		Riwayat:      riwayat,
	}

	return s.repo.DeclinePerjalananDinas(ctx, params)
}

func (s *PerjalananDinasImpl) GenerateNomorDokumenGeneral(ctx context.Context, prefix string) (string, error) {
	now := time.Now()
	bulan := fmt.Sprintf("%02d", int(now.Month()))
	tahun := fmt.Sprintf("%d", now.Year())
	dept := "GA"

	pattern := fmt.Sprintf("%s/%s/%s/%s/%%", prefix, dept, bulan, tahun)

	lastno, err := s.repo.GetLastNomorDokumenGeneral(ctx, pattern)
	if err != nil {
		return "", err
	}

	newSeq := 1
	if lastno != "" {
		parts := strings.Split(lastno, "/")
		if len(parts) >= 5 {
			lastSeqStr := parts[len(parts)-1]
			lastSeqInt, _ := strconv.Atoi(lastSeqStr)
			newSeq = lastSeqInt + 1
		}
	}
	return fmt.Sprintf("%s/%s/%s/%s/%03d", prefix, dept, bulan, tahun, newSeq), nil
}

func (s *PerjalananDinasImpl) GenerateNomorDokumenSpecific(ctx context.Context, tipe string, prefix string) (string, error) {
	now := time.Now()
	bulan := fmt.Sprintf("%02d", int(now.Month()))
	tahun := fmt.Sprintf("%d", now.Year())
	dept := "GA"

	pattern := fmt.Sprintf("%s/%s/%s/%s/%%", prefix, dept, bulan, tahun)
	lastno, err := s.repo.GetLastNomorDokumenSpecific(ctx, tipe, pattern)
	if err != nil {
		return "", err
	}
	newseq := 1
	if lastno != "" {
		parts := strings.Split(lastno, "/")
		if len(parts) >= 5 {
			lastSeqStr := parts[len(parts)-1]
			lastSeqInt, _ := strconv.Atoi(lastSeqStr)
			newseq = lastSeqInt + 1
		}
	}

	return fmt.Sprintf("%s/%s/%s/%s/%03d", prefix, dept, bulan, tahun, newseq), nil
}

func (s *PerjalananDinasImpl) GetDropdownPPD(ctx context.Context) ([]repository.DropdownPPDResponse, error) {

	claims, ok := config.GetClaimsFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized: gagal mengambil data profil dari token")
	}

	userID := claims.UserID

	return s.repo.GetListPPDForRealisasi(ctx, userID)
}

func (s *PerjalananDinasImpl) GetListPendingPerjalananDinas(ctx context.Context) ([]model.RequestPPD, error) {

	claims, ok := config.GetClaimsFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized: gagal mengambil data profil dari token")
	}

	userID := claims.UserID
	jabatan := claims.Jabatan

	if jabatan == "Pegawai" {
		return nil, errors.New("forbidden: pegawai tidak memiliki akses ke daftar pending approval")
	}
	listData, err := s.repo.GetListPendingPerjalananDinas(ctx, jabatan, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar pending: %w", err)
	}

	return listData, nil
}

func (s *PerjalananDinasImpl) GetListPerjalananDetail(ctx context.Context, ppdid uint) ([]model.PPDRincianTambahan, error) {
	detailPPD, err := s.repo.GetDetailPerjalananDinas(ctx, ppdid)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil detail: %w", err)
	}

	return detailPPD.RincianTambahan, nil
}

func (s *PerjalananDinasImpl) ApprovePerjalananDinas(ctx context.Context, ppdid uint, catatan string) error {

	claims, ok := config.GetClaimsFromContext(ctx)
	if !ok {
		return errors.New("unauthorized: gagal mengambil data profil dari token")
	}

	userID := claims.UserID
	jabatan := claims.Jabatan

	currentStatus, err := s.repo.GetStatusPerjalananDinas(ctx, ppdid)
	if err != nil {
		return errors.New("perjalanan dinas tidak ditemukan")
	}

	if (jabatan == "Atasan" && currentStatus != "Menunggu Atasan") ||
		(jabatan == "HRGA" && currentStatus != "Menunggu HRGA") ||
		(jabatan == "Direktur" && currentStatus != "Menunggu Direktur") ||
		(jabatan == "Finance" && currentStatus != "Menunggu Finance") {
		return errors.New("tidak dapat menyetujui perjalanan dinas dengan status saat ini")
	}

	OpsiStatus, exist := ProsesStatus[currentStatus]
	if !exist || len(OpsiStatus) == 0 {
		return errors.New("status perjalanan dinas tidak valid")
	}

	nextStatus := OpsiStatus[0]

	var newdokumen []model.Dokumen

	if jabatan == "Atasan" {
		nomorUmum, errGen := s.GenerateNomorDokumenGeneral(ctx, "MMI")
		if errGen != nil {
			return fmt.Errorf("gagal generate nomor dokumen: %w", errGen)
		}

		newdokumen = append(newdokumen, model.Dokumen{
			DocRefID:     ppdid,
			DocRefType:   "RequestPPD",
			UserID:       userID,
			NomorDokumen: nomorUmum,
			TipeDokumen:  "Pengajuan Perjalanan Dinas",
		})
	}

	if jabatan == "Finance" {
		nomorUmum, errGen := s.GenerateNomorDokumenGeneral(ctx, "DK")
		if errGen != nil {
			return fmt.Errorf("gagal generate nomor dokumen umum: %w", errGen)
		}

		nomorSpesifik, errSpec := s.GenerateNomorDokumenSpecific(ctx, "Bon Sementara", "BS")
		if errSpec != nil {
			return fmt.Errorf("gagal generate nomor dokumen spesifik: %w", errSpec)
		}

		newdokumen = append(newdokumen, model.Dokumen{
			DocRefID:         ppdid,
			DocRefType:       "RequestPPD",
			UserID:           userID,
			NomorDokumen:     nomorUmum,
			NomorTipeDokumen: nomorSpesifik,
			TipeDokumen:      "Bon Sementara",
		})
	}

	riwayat := &model.RiwayatApproval{
		RequestPPDID: ppdid,
		UserID:       userID,
		Jabatan:      jabatan,
		Tindakan:     "Disetujui",
		Catatan:      catatan,
	}

	params := repository.ApprovePerjalananDinasparams{
		RequestPPDID: ppdid,
		NextStatus:   nextStatus,
		NewDokumen:   newdokumen,
		Riwayat:      riwayat,
	}

	return s.repo.ApprovePerjalananDinas(ctx, params)
}
