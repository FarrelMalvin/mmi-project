package service

import (
	"context"
	"errors"
	"fmt"
	"golang-mmi/internal/config"
	"golang-mmi/internal/model"
	"golang-mmi/internal/repository"
)

var ProsesStatusRBS = map[string][]string{
	"Draft": {"Menunggu HRGA"},

	"Menunggu HRGA":     {"Menunggu Direktur", "Ditolak HRGA"},
	"Menunggu Direktur": {"Menunggu Finance", "Ditolak Direktur"},
	"Menunggu Finance":  {"Selesai", "Ditolak Finance"},
}

type ServiceRBS interface {
	CreateRealisasiBon(ctx context.Context, req *model.RealisasiBonSementara) error
	GetListRBS(ctx context.Context, page, limit, bulan, tahun int) ([]repository.RiwayatRBSResponse, int64, int64, error)
	ApproveRBS(ctx context.Context, rbsid uint, catatan string) error
	DeclineRBS(ctx context.Context, ppdid uint, catatan string) error
	GetDropdownPPD(ctx context.Context) ([]repository.DropdownPPDResponse, error)
	GetListPendingRBS(ctx context.Context) ([]model.RealisasiBonSementara, error)
	GetListRBSDetail(ctx context.Context, ppdid uint) ([]model.RBSrincian, error)
}

type RealisasiBonImpl struct {
	repo       repository.RealisasiBonRepository
	serviceppd ServicePPD
}

func NewRealisasiRBSService(repo repository.RealisasiBonRepository, serviceppd ServicePPD) *RealisasiBonImpl {
	return &RealisasiBonImpl{
		repo:       repo,
		serviceppd: serviceppd,
	}
}

func (s *RealisasiBonImpl) CreateRealisasiBon(ctx context.Context, req *model.RealisasiBonSementara) error {

	claims, ok := config.GetClaimsFromContext(ctx)
	if !ok {
		return errors.New("unauthorized: gagal mengambil data profil dari token")
	}

	userID := claims.UserID
	jabatan := claims.Jabatan

	userID, okID := ctx.Value("user_id").(uint)
	if !okID {
		return errors.New("invalid user ID in context")
	}
	jabatan, okjabatan := ctx.Value("jabatan").(string)
	if !okjabatan {
		return errors.New("invalid jabatan in context")
	}

	req.UserID = userID

	switch jabatan {
	case "Pegawai":
		req.Status = "Menunggu Atasan"
	case "Atasan":
		req.Status = "Menunggu HRGA"
	case "HRGA":
		req.Status = "Menunggu Direktur"
	default:
		return errors.New("invalid jabatan")
	}

	return s.repo.CreateRealisasiBon(ctx, req)

}

func (s *RealisasiBonImpl) GetListRBS(ctx context.Context, page, limit, bulan, tahun int) ([]repository.RiwayatRBSResponse, int64, int64, error) {

	jabatan, okjabatan := ctx.Value("jabatan").(string)
	if !okjabatan {
		return nil, 0, 0, errors.New("invalid jabatan in context")
	}

	filter := repository.FilterRBS{
		Bulan: bulan,
		Tahun: tahun,
	}
	switch jabatan {
	case "HRGA", "Direktur", "Finance":
		return s.repo.GetListRiwayatRealisasiBon(ctx, page, limit, filter)

	case "Pegawai":

		userID, okID := ctx.Value("user_id").(uint)
		if !okID {
			return nil, 0, 0, errors.New("invalid user ID in context")
		}

		return s.repo.GetListRiwayatRealisasiBonById(ctx, page, limit, userID)

	case "Atasan":
		return s.repo.GetListRiwayatRealisasiBon(ctx, page, limit, filter)

	default:
		return nil, 0, 0, errors.New("jabatan tidak terdaftar")
	}
}

func (s *RealisasiBonImpl) ApproveRBS(ctx context.Context, rbsid uint, catatan string) error {

	claims, ok := config.GetClaimsFromContext(ctx)
	if !ok {
		return errors.New("unauthorized: gagal mengambil data profil dari token")
	}

	userID := claims.UserID
	jabatan := claims.Jabatan

	currentStatus, err := s.repo.GetStatusRBS(ctx, rbsid)
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

	if jabatan == "Finance" {
		nomorUmum, errGen := s.serviceppd.GenerateNomorDokumenGeneral(ctx, "DK")
		if errGen != nil {
			return fmt.Errorf("gagal generate nomor dokumen: %w", errGen)
		}

		newdokumen = append(newdokumen, model.Dokumen{
			DocRefID:     rbsid,
			DocRefType:   "RealisasiBonSementara",
			UserID:       userID,
			NomorDokumen: nomorUmum,
			TipeDokumen:  "Umum",
		})
	}

	riwayat := &model.RiwayatApproval{
		RequestPPDID: rbsid,
		UserID:       userID,
		Jabatan:      jabatan,
		Tindakan:     "Disetujui",
		Catatan:      catatan,
	}

	params := repository.ApproveRBSResponse{
		RealisasiBonID: rbsid,
		NextStatus:     nextStatus,
		NewDokumen:     newdokumen,
		Riwayat:        riwayat,
	}

	return s.repo.ApproveRBS(ctx, params)
}

func (s *RealisasiBonImpl) GetDropdownPPD(ctx context.Context) ([]repository.DropdownPPDResponse, error) {

	dropdownData, err := s.serviceppd.GetDropdownPPD(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data dropdown PPD: %w", err)
	}

	return dropdownData, nil
}

func (s *RealisasiBonImpl) GetListPendingRBS(ctx context.Context) ([]model.RealisasiBonSementara, error) {

	claims, ok := config.GetClaimsFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized: gagal mengambil data profil dari token")
	}

	userID := claims.UserID
	jabatan := claims.Jabatan

	if jabatan == "Pegawai" {
		return nil, errors.New("forbidden: pegawai tidak memiliki akses ke daftar pending approval")
	}
	listData, err := s.repo.GetListPendingRBS(ctx, jabatan, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar pending: %w", err)
	}

	return listData, nil
}

func (s *RealisasiBonImpl) DeclineRBS(ctx context.Context, ppdid uint, catatan string) error {

	claims, ok := config.GetClaimsFromContext(ctx)
	if !ok {
		return errors.New("unauthorized: gagal mengambil data profil dari token")
	}

	userID := claims.UserID
	jabatan := claims.Jabatan

	currentStatus, err := s.repo.GetStatusRBS(ctx, ppdid)
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

	params := repository.DeclineRBSResponse{
		RealisasiBonID: ppdid,
		NextStatus:     nextStatus,
		Riwayat:        riwayat,
	}

	return s.repo.DeclineRBS(ctx, params)
}

func (s *RealisasiBonImpl) GetListRBSDetail(ctx context.Context, ppdid uint) ([]model.RBSrincian, error) {
	detailRBS, err := s.repo.GetDetailRBS(ctx, ppdid)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil detail: %w", err)
	}

	return detailRBS.RBSrincian, nil
}
