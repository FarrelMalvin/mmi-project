package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"strconv"
	"strings"
	"time"

	"golang-mmi/internal/constant"
	"golang-mmi/internal/dto"
	"golang-mmi/internal/model"
	"golang-mmi/internal/repository"
	"golang-mmi/internal/utils"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

var (
	ErrRBSTidakDitemukan    = errors.New("realisasi bon sementara tidak ditemukan")
	ErrRBSStatusTidakValid  = errors.New("status tidak valid untuk aksi ini")
	ErrRBSAksesditolak      = errors.New("jabatan tidak memiliki akses ke fitur ini")
	ErrRBSJabatanTidakValid = errors.New("jabatan tidak valid untuk aksi ini")
)

var ProsesStatusRBS = map[string][]string{
	constant.StatusDraft:           {constant.StatusMenungguAtasan},
	constant.StatusMenungguAtasan:  {constant.StatusMenungguHRGA, constant.StatusDitolakAtasan},
	constant.StatusMenungguHRGA:    {constant.StatusMenungguFinance, constant.StatusDitolakFinance},
	constant.StatusMenungguFinance: {constant.StatusSelesai, constant.StatusDitolakFinance},
}

var statusRBSYangDiharapkan = map[string]string{
	constant.JabatanHRGA:     constant.StatusMenungguHRGA,
	constant.JabatanDirektur: constant.StatusMenungguDirektur,
	constant.JabatanFinance:  constant.StatusMenungguFinance,
}

type ServiceRBS interface {
	CreateRealisasiBon(ctx context.Context, req dto.CreateRBSRequest) error
	GetListRBS(ctx context.Context, req dto.RBSListRequest) ([]dto.ListRBSResponse, int64, int64, error)
	ApproveRBS(ctx context.Context, req *dto.ApproveRBSRequest) error
	DeclineRBS(ctx context.Context, req dto.DeclineRBSRequest) error
	GetDropdownPPD(ctx context.Context, userid uint) ([]dto.DropdownPPDResponse, error)
	GetListPendingRBS(ctx context.Context, userid uint, jabatan string) ([]dto.ListRBSResponse, error)
	GetListRBSDetail(ctx context.Context, ppdid uint) ([]model.RBSrincian, error)
	FillRBSPDF(ctx context.Context, ppdID uint, userID uint, templatePath string, w io.Writer) error
	ExportRBSExcel(ctx context.Context, req dto.RBSListRequest, w io.Writer) error
}

type RealisasiBonImpl struct {
	repo         repository.RealisasiBonRepository
	repoppd      repository.PerjalananDinasRepository
	servicedoc   DocumentService
	notifManager *utils.NotificationManager
}

func NewRealisasiRBSService(repo repository.RealisasiBonRepository, repoppd repository.PerjalananDinasRepository, servicedoc DocumentService, notifManager *utils.NotificationManager) *RealisasiBonImpl {
	return &RealisasiBonImpl{
		repo:         repo,
		repoppd:      repoppd,
		servicedoc:   servicedoc,
		notifManager: notifManager,
	}
}

func (s *RealisasiBonImpl) CreateRealisasiBon(ctx context.Context, req dto.CreateRBSRequest) error {

	var status string
	var autoApprovals []model.RiwayatApproval

	switch req.Jabatan {
	case constant.JabatanPegawai:
		status = constant.StatusMenungguAtasan
	case constant.JabatanAtasan:
		status = constant.StatusMenungguHRGA
		autoApprovals = append(autoApprovals, model.RiwayatApproval{
			UserID:   req.UserID,
			Jabatan:  constant.JabatanAtasan,
			Tindakan: constant.TindakanDisetujui,
			Catatan:  "Auto-approved (diajukan oleh atasan)",
		})

	case constant.JabatanHRGA:
		status = constant.StatusMenungguFinance
		autoApprovals = append(autoApprovals,
			model.RiwayatApproval{
				UserID:   req.UserID,
				Jabatan:  constant.JabatanHRGA,
				Tindakan: constant.TindakanDisetujui,
				Catatan:  "Auto-approved (diajukan oleh HRGA)",
			},
		)

	default:
		return ErrRBSJabatanTidakValid
	}

	totalEstimasi, err := s.repoppd.GetTotalEstimasi(ctx, req.RequestPPDID)
	if err != nil {
		return fmt.Errorf("gagal mengambil data estimasi PPD: %w", err)
	}

	var calculatedTotalRealisasi int64
	var rincianItems []model.RBSrincian

	for _, item := range req.Items {
		tanggal, err := time.Parse("2006-01-02", item.Tanggal)
		if err != nil {
			return fmt.Errorf("format tanggal tidak valid '%s': %w", item.Tanggal, err)
		}

		calculatedTotalRealisasi += int64(item.Total)

		rincianItems = append(rincianItems, model.RBSrincian{
			Uraian:           item.Uraian,
			TanggalTransaksi: tanggal,
			Kuantitas:        item.Qty,
			HargaUnit:        item.HargaUnit,
			TotalHarga:       item.Total,
			Kategori:         item.Kategori,
			UrlStruk:         item.UrlStruk,
		})
	}

	selisih := totalEstimasi - calculatedTotalRealisasi

	newRBS := model.RealisasiBonSementara{
		RequestPPDID:       req.RequestPPDID,
		UserID:             req.UserID,
		TotalRealisasi:     calculatedTotalRealisasi,
		Selisih:            selisih,
		PeriodeBerangkat:   req.PeriodeBerangkat,
		PeriodeKembali:     req.PeriodeKembali,
		Status:             status,
		RBSrincian:         rincianItems,
		RiwayatPersetujuan: autoApprovals,
	}

	if err := s.repo.CreateRealisasiBon(ctx, &newRBS); err != nil {
		return fmt.Errorf("gagal membuat realisasi bon sementara: %w", err)
	}

	return nil
}

func (s *RealisasiBonImpl) GetListRBS(ctx context.Context, req dto.RBSListRequest) ([]dto.ListRBSResponse, int64, int64, error) {
	var dataModels []model.RBSListView
	var totalData, totalSum int64
	var err error

	filter := repository.FilterRBS{
		Bulan: req.Bulan,
		Tahun: req.Tahun,
	}
	switch req.Jabatan {
	case constant.JabatanHRGA, constant.JabatanDirektur, constant.JabatanFinance:
		dataModels, totalData, totalSum, err = s.repo.GetListRiwayatRealisasiBon(ctx, req.Page, req.Limit, filter)

	case constant.JabatanPegawai:
		dataModels, totalData, totalSum, err = s.repo.GetListRiwayatRealisasiBonById(ctx, req.Page, req.Limit, req.UserID)

	case constant.JabatanAtasan:
		dataModels, totalData, totalSum, err = s.repo.GetListRiwayatRealisasiBon(ctx, req.Page, req.Limit, filter)
	default:
		return nil, 0, 0, ErrJabatanTidakValid
	}

	if err != nil {
		return nil, 0, 0, fmt.Errorf("gagal mengambil daftar RBS: %w", err)
	}

	response := make([]dto.ListRBSResponse, 0, len(dataModels))
	for _, item := range dataModels {
		response = append(response, dto.ListRBSResponse{
			ID: item.Id,
			//Nama:           item.Nama,
			NomorDokumen:   item.NomorDokumen,
			Status:         item.Status,
			TotalRealisasi: item.TotalRealisasi,
			TotalEstimasi:  item.TotalEstimasi,
			Selisih:        item.Selisih,
		})
	}

	return response, totalData, totalSum, nil
}

func (s *RealisasiBonImpl) ApproveRBS(ctx context.Context, req *dto.ApproveRBSRequest) error {
	currentStatus, err := s.repo.GetStatusRBS(ctx, req.RealisasiBonID)
	if err != nil {
		return ErrRBSTidakDitemukan
	}

	if err := validateJabatanDanStatusRBS(req.Jabatan, currentStatus); err != nil {
		return err
	}

	OpsiStatus, exist := ProsesStatusRBS[currentStatus]
	if !exist || len(OpsiStatus) == 0 {
		return ErrRBSStatusTidakValid
	}

	nextStatus := OpsiStatus[0]

	var newdokumen []model.Dokumen

	if req.Jabatan == constant.JabatanFinance {
		nomorUmum, errGen := s.servicedoc.GenerateNomorDokumenGeneral(ctx, constant.KodeDeptFinance)
		if errGen != nil {
			return fmt.Errorf("gagal generate nomor dokumen: %w", errGen)
		}

		newdokumen = append(newdokumen, model.Dokumen{
			DocRefID:     req.RealisasiBonID,
			DocRefType:   constant.DocRefTypeRBS,
			UserID:       req.UserID,
			NomorDokumen: nomorUmum,
			TipeDokumen:  constant.TipeDokumenRBS,
		})
	}
	riwayat := &model.RiwayatApproval{
		DocRefID:   req.RealisasiBonID,
		DocRefType: constant.DocRefTypeRBS,
		UserID:     req.UserID,
		Jabatan:    req.Jabatan,
		Tindakan:   constant.TindakanDisetujui,
		Catatan:    req.Catatan,
	}

	params := repository.ApproveRBSParam{
		RealisasiBonID: req.RealisasiBonID,
		NextStatus:     nextStatus,
		NewDokumen:     newdokumen,
		Riwayat:        riwayat,
	}

	if err := s.repo.ApproveRBS(ctx, params); err != nil {
		return fmt.Errorf("gagal menyimpan persetujuan perjalanan dinas: %w", err)
	}

	go func() {
		switch req.Jabatan {
		case constant.JabatanAtasan:
			s.notifManager.SendToRole(constant.JabatanHRGA, constant.NotifAtasanSetujuRBS)
		case constant.JabatanHRGA:
			s.notifManager.SendToRole(constant.JabatanFinance, constant.NotifHRGASetujuRBS)
		case constant.JabatanFinance:
			s.notifManager.SendToRole(constant.JabatanPegawai, constant.NotifFinanceSetujuRBS)
		}
	}()

	return nil
}

func (s *RealisasiBonImpl) GetDropdownPPD(ctx context.Context, userID uint) ([]dto.DropdownPPDResponse, error) {

	data, err := s.repoppd.GetListPPDForRealisasi(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.DropdownPPDResponse, 0)
	for _, item := range data {
		response = append(response, dto.DropdownPPDResponse{
			ID:               item.ID,
			NomorDokumen:     item.NomorDokumen,
			Tujuan:           item.Tujuan,
			PeriodeBerangkat: item.PeriodeBerangkat.Format("02/01/2006"),
			PeriodeKembali:   item.PeriodeKembali.Format("02/01/2006"),
		})
	}

	return response, nil
}

func (s *RealisasiBonImpl) GetListPendingRBS(ctx context.Context, userID uint, jabatan string) ([]dto.ListRBSResponse, error) {
	var dataModels []model.RBSListView
	var err error

	if jabatan == constant.JabatanPegawai || jabatan == "" {
		return nil, ErrRBSAksesditolak
	}

	dataModels, err = s.repo.GetListPendingRBS(ctx, jabatan, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar pending: %w", err)
	}

	response := make([]dto.ListRBSResponse, 0)
	for _, item := range dataModels {
		response = append(response, dto.ListRBSResponse{
			ID:             item.Id,
			NomorDokumen:   item.NomorDokumen,
			TotalRealisasi: item.TotalRealisasi,
			TotalEstimasi:  item.TotalEstimasi,
			Selisih:        item.Selisih,
			Status:         item.Status,
		})
	}

	return response, nil
}

func (s *RealisasiBonImpl) DeclineRBS(ctx context.Context, req dto.DeclineRBSRequest) error {

	currentStatus, err := s.repo.GetStatusRBS(ctx, req.RealisasiBonID)
	if err != nil {
		return ErrRBSTidakDitemukan
	}

	if err := validateJabatanDanStatusRBS(req.Jabatan, currentStatus); err != nil {
		return err
	}

	OpsiStatus, exist := ProsesStatusRBS[currentStatus]

	if !exist || len(OpsiStatus) < 2 {
		return ErrRBSStatusTidakValid
	}

	nextStatus := OpsiStatus[1]

	riwayat := &model.RiwayatApproval{
		DocRefID:   req.RealisasiBonID,
		DocRefType: constant.DocRefTypeRBS,
		UserID:     req.UserID,
		Jabatan:    req.Jabatan,
		Tindakan:   constant.TindakanDitolak,
		Catatan:    req.Catatan,
	}

	params := repository.DeclineRBSParam{
		RealisasiBonID: req.RealisasiBonID,
		NextStatus:     nextStatus,
		Riwayat:        riwayat,
	}

	go func() {
		s.notifManager.SendToRole(constant.JabatanPegawai, constant.NotifDitolakRBS)
	}()

	return s.repo.DeclineRBS(ctx, params)
}

func (s *RealisasiBonImpl) GetListRBSDetail(ctx context.Context, ppdid uint) ([]model.RBSrincian, error) {
	detailRBS, err := s.repo.GetDetailRBS(ctx, ppdid)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil detail: %w", err)
	}

	return detailRBS.RBSrincian, nil
}

func (s *RealisasiBonImpl) GetDataRBSForPDF(ctx context.Context, ppdID uint, userID uint) (*dto.RBSDataToPDF, error) {
	data, err := s.repo.GetDetailRBS(ctx, ppdID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data RBS untuk PDF: %w", err)
	}

	var nomorDokumen string
	var tanggalTerbit string

	for _, doc := range data.Dokumen {
		if doc.TipeDokumen == constant.TipeDokumenRBS {
			nomorDokumen = doc.NomorDokumen
			tanggalTerbit = utils.FormatTanggal(doc.CreatedAt)

			break
		}
	}

	var pathFinance string
	var namaFinance string
	var wilayahDisetujui string
	var pathAtasan string
	var namaAtasan string
	var pathHR string
	var namaHR string

	riwayats := findTandaTanganByJabatanRBS(data.RiwayatPersetujuan)
	if r, ok := riwayats[constant.JabatanFinance]; ok {
		pathFinance = r.User.PathTandaTangan
		wilayahDisetujui = r.User.Wilayah
		namaFinance = r.User.Nama
	}
	if r, ok := riwayats[constant.JabatanHRGA]; ok {
		pathHR = r.User.PathTandaTangan
		namaHR = r.User.Nama
	}
	if r, ok := riwayats[constant.JabatanAtasan]; ok {
		pathAtasan = r.User.PathTandaTangan
		namaAtasan = r.User.Nama
	}

	totalEstimasi, err := s.repoppd.GetTotalEstimasi(ctx, data.RequestPPDID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil total estimasi untuk PDF: %w", err)
	}

	pdfData := &dto.RBSDataToPDF{
		NomorDokumen:           nomorDokumen,
		TanggalTerbit:          tanggalTerbit,
		NomorBS:                data.NomorBonSementara,
		TanggalDibuat:          utils.FormatTanggal(data.CreatedAt),
		WilayahDisetujui:       wilayahDisetujui,
		Selisih:                utils.FormatRupiah(data.Selisih),
		TotalRealisasi:         utils.FormatRupiah(data.TotalRealisasi),
		TotalBon:               utils.FormatRupiah(totalEstimasi),
		Periode:                utils.FormatTanggal(data.PeriodeBerangkat) + " - " + utils.FormatTanggal(data.PeriodeKembali),
		PathTandaTanganHRGA:    pathHR,
		NamaHRGA:               namaHR,
		PathTandaTanganFinance: pathFinance,
		NamaFinance:            namaFinance,
		PathTandaTanganAtasan:  pathAtasan,
		NamaAtasan:             namaAtasan,
	}

	for _, rincian := range data.RBSrincian {
		pdfData.Rincian = append(pdfData.Rincian, dto.RBSRincianPDF{
			Uraian:    rincian.Uraian,
			Tanggal:   utils.FormatTanggal(rincian.TanggalTransaksi),
			Kuantitas: strconv.Itoa(rincian.Kuantitas),
			HargaUnit: utils.FormatRupiah(rincian.HargaUnit),
			Total:     utils.FormatRupiah(rincian.TotalHarga),
		})
	}

	return pdfData, nil
}

func (s *RealisasiBonImpl) FillRBSPDF(ctx context.Context, ppdID uint, userID uint, templatePath string, w io.Writer) error {

	pdfData, err := s.GetDataRBSForPDF(ctx, ppdID, userID)
	if err != nil {
		return fmt.Errorf("gagal mengambil data untuk PDF: %w", err)
	}
	logoPath := "internal/assets/logo/logo.png"
	logoBase64 := utils.GetBase64Image(logoPath)

	tmplData := map[string]interface{}{
		"Data":       pdfData,
		"Logo":       logoBase64,
		"TtdHRGA":    utils.GetBase64Image(pdfData.PathTandaTanganHRGA),
		"TtdFinance": utils.GetBase64Image(pdfData.PathTandaTanganFinance),
		"TtdAtasan":  utils.GetBase64Image(pdfData.PathTandaTanganAtasan),
		"LogoBase64": utils.GetBase64Image(logoPath),
	}

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("gagal memuat template HTML: %w", err)
	}

	var htmlBuffer bytes.Buffer
	if err := tmpl.Execute(&htmlBuffer, tmplData); err != nil {
		return fmt.Errorf("gagal mengeksekusi template: %w", err)
	}

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return fmt.Errorf("gagal inisialisasi wkhtmltopdf: %w", err)
	}

	pdfg.Dpi.Set(300)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)

	page := wkhtmltopdf.NewPageReader(&htmlBuffer)
	page.EnableLocalFileAccess.Set(true)
	pdfg.AddPage(page)

	if err := pdfg.Create(); err != nil {
		return fmt.Errorf("gagal men-generate PDF via wkhtmltopdf: %w", err)
	}

	_, err = w.Write(pdfg.Bytes())
	if err != nil {
		return fmt.Errorf("gagal menulis output PDF ke writer: %w", err)
	}

	return nil
}

func (s *RealisasiBonImpl) ExportRBSExcel(ctx context.Context, req dto.RBSListRequest, w io.Writer) error {
	if req.Jabatan != constant.JabatanHRGA {
		return ErrRBSAksesditolak
	}

	filter := repository.FilterRBS{
		Bulan: req.Bulan,
		Tahun: req.Tahun,
	}

	dataDatar, err := s.repo.GetDataRBSforCsv(ctx, filter)
	if err != nil {
		return fmt.Errorf("gagal mengambil data rbs: %w", err)
	}

	mapKategori := make(map[string]*model.KategoriData)
	var urutanKategori []string

	for _, item := range dataDatar {
		if _, exists := mapKategori[item.Kategori]; !exists {
			mapKategori[item.Kategori] = &model.KategoriData{
				NamaKategori: item.Kategori,
				Items:        []model.RBSDataforCsv{},
				Total:        0,
			}
			urutanKategori = append(urutanKategori, item.Kategori)
		}
		mapKategori[item.Kategori].Items = append(mapKategori[item.Kategori].Items, item)
		mapKategori[item.Kategori].Total += item.TotalHarga
	}

	f := excelize.NewFile()

	styleNum, err := f.NewStyle(`{"number_format": 3}`)
	if err != nil {
		return fmt.Errorf("gagal membuat style excel: %w", err)
	}
	styleHeader, _ := f.NewStyle(`{
		"fill": {"type": "pattern", "pattern": 1, "color": ["#D9D9D9"]},
		"font": {"bold": true},
		"alignment": {"horizontal": "center"}
	}`)
	styleKategori, _ := f.NewStyle(`{
		"fill": {"type": "pattern", "pattern": 1, "color": ["#FFF2CC"]},
		"font": {"bold": true}
	}`)

	sheet1 := "Data Pengeluaran"
	f.SetSheetName("Sheet1", sheet1)

	headers := []interface{}{"NO", "No Referensi BS", "Nama Karyawan", "Periode", "Tanggal", "Uraian", "Quantity", "Harga/Unit", "Total"}
	f.SetSheetRow(sheet1, "A1", &headers)
	f.SetCellStyle(sheet1, "A1", "I1", styleHeader)

	currentRow := 2
	itemNo := 1
	var grandTotal int64 = 0

	for _, katName := range urutanKategori {
		kat := mapKategori[katName]

		f.SetCellValue(sheet1, fmt.Sprintf("A%d", currentRow), "KATEGORI: "+kat.NamaKategori)
		f.SetCellStyle(sheet1, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("I%d", currentRow), styleKategori)
		currentRow++

		for _, item := range kat.Items {
			f.SetCellValue(sheet1, fmt.Sprintf("A%d", currentRow), itemNo)
			f.SetCellValue(sheet1, fmt.Sprintf("B%d", currentRow), item.NomorReferensiBS)
			f.SetCellValue(sheet1, fmt.Sprintf("C%d", currentRow), item.Nama)
			f.SetCellValue(sheet1, fmt.Sprintf("D%d", currentRow), utils.FormatTanggal(item.Periode))
			f.SetCellValue(sheet1, fmt.Sprintf("E%d", currentRow), utils.FormatTanggal(item.TanggalTransaksi))
			f.SetCellValue(sheet1, fmt.Sprintf("F%d", currentRow), item.Uraian)
			f.SetCellValue(sheet1, fmt.Sprintf("G%d", currentRow), item.Kuantitas)

			f.SetCellValue(sheet1, fmt.Sprintf("H%d", currentRow), item.HargaUnit)
			f.SetCellStyle(sheet1, fmt.Sprintf("H%d", currentRow), fmt.Sprintf("H%d", currentRow), styleNum)

			f.SetCellValue(sheet1, fmt.Sprintf("I%d", currentRow), item.TotalHarga)
			f.SetCellStyle(sheet1, fmt.Sprintf("I%d", currentRow), fmt.Sprintf("I%d", currentRow), styleNum)

			currentRow++
			itemNo++
		}

		f.SetCellValue(sheet1, fmt.Sprintf("H%d", currentRow), "Total "+kat.NamaKategori+":")
		f.SetCellValue(sheet1, fmt.Sprintf("I%d", currentRow), kat.Total)
		f.SetCellStyle(sheet1, fmt.Sprintf("I%d", currentRow), fmt.Sprintf("I%d", currentRow), styleNum)

		grandTotal += kat.Total
		currentRow += 2
	}

	f.SetCellValue(sheet1, fmt.Sprintf("H%d", currentRow), "GRAND TOTAL:")
	f.SetCellValue(sheet1, fmt.Sprintf("I%d", currentRow), grandTotal)
	f.SetCellStyle(sheet1, fmt.Sprintf("I%d", currentRow), fmt.Sprintf("I%d", currentRow), styleNum)

	sheet2 := "Dashboard"
	f.NewSheet(sheet2)

	f.SetCellValue(sheet2, "A1", "Kategori")
	f.SetCellValue(sheet2, "B1", "Total Pengeluaran")
	f.SetCellStyle(sheet2, "A1", "B1", styleHeader)

	for i, katName := range urutanKategori {
		rowIdx := i + 2
		kat := mapKategori[katName]
		f.SetCellValue(sheet2, fmt.Sprintf("A%d", rowIdx), kat.NamaKategori)

		f.SetCellValue(sheet2, fmt.Sprintf("B%d", rowIdx), kat.Total)
		f.SetCellStyle(sheet2, fmt.Sprintf("B%d", rowIdx), fmt.Sprintf("B%d", rowIdx), styleNum)
	}

	akhirBarisSummary := len(urutanKategori) + 1
	chartFormat := fmt.Sprintf(`{
		"type": "col",
		"series": [
			{
				"name": "%s!$B$1",
				"categories": "%s!$A$2:$A$%d",
				"values": "%s!$B$2:$B$%d"
			}
		],
		"title": {"name": "Grafik Pengeluaran per Kategori"},
		"x_axis": {
			"font": {
				"bold": true,
				"color": "#000000",
				"size": 11
			}
		},
		"y_axis": {
			"font": {
				"bold": true,
				"color": "#000000",
				"size": 11
			}
		}
	}`, sheet2, sheet2, akhirBarisSummary, sheet2, akhirBarisSummary)

	if err := f.AddChart(sheet2, "D2", chartFormat); err != nil {
		return fmt.Errorf("gagal membuat chart excel: %w", err)
	}

	pieChartFormat := fmt.Sprintf(`{
		"type": "pie",
		"series": [
			{
				"name": "%s!$B$1",
				"categories": "%s!$A$2:$A$%d",
				"values": "%s!$B$2:$B$%d"
			}
		],
		"title": {"name": "Persentase Pengeluaran per Kategori"}
	}`, sheet2, sheet2, akhirBarisSummary, sheet2, akhirBarisSummary)

	if err := f.AddChart(sheet2, "M2", pieChartFormat); err != nil {
		return fmt.Errorf("gagal membuat pie chart excel: %w", err)
	}

	if err := f.Write(w); err != nil {
		return fmt.Errorf("gagal menulis output excel ke writer: %w", err)
	}

	return nil
}

func validateJabatanDanStatusRBS(jabatan, currentStatus string) error {
	expected, ok := statusRBSYangDiharapkan[jabatan]
	if !ok {
		return ErrRBSJabatanTidakValid
	}
	if currentStatus != expected {
		return ErrRBSStatusTidakValid
	}
	return nil
}

func findTandaTanganByJabatanRBS(riwayats []model.RiwayatApproval) map[string]model.RiwayatApproval {
	hasil := make(map[string]model.RiwayatApproval)
	for _, r := range riwayats {
		if strings.Contains(r.Tindakan, "Disetujui") {
			hasil[r.Jabatan] = r
		}
	}
	return hasil
}
