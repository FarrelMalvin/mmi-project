package dto

type ListRBSResponse struct {
	ID             uint   `json:"id"`
	NomorDokumen   string `json:"nomor_dokumen,omitempty"`
	Nama           string `json:"nama,omitempty"`
	TotalRealisasi int64  `json:"total_realisasi"`
	TotalEstimasi  int64  `json:"total_estimasi"`
	Selisih        int64  `json:"selisih"`
	Status         string `json:"status"`
}

type RBSDataToExcel struct {
	NomorDokumen string
	Nama         string
	Kategori     string
	Tanggal      string
	Periode      string
	Uraian       string
	Kuantitas    int
	HargaUnit    int64
	Total        int64
}

type RBSDataToPDF struct {
	NomorDokumen           string
	TanggalTerbit          string
	Periode                string
	NomorBS                string
	TanggalDibuat          string
	PathTandaTanganHRGA    string
	NamaHRGA               string
	PathTandaTanganAtasan  string
	NamaAtasan             string
	PathTandaTanganFinance string
	NamaFinance            string
	TotalRealisasi         string
	TotalBon               string
	TotalEstimasi          string
	Selisih                string
	WilayahDisetujui       string
	Rincian                []RBSRincianPDF
	RiwayatPersetujuan     []RiwayatApprovaltoPDF
}

type RBSRincianPDF struct {
	Uraian    string
	Tanggal   string
	Kuantitas string
	HargaUnit string
	Total     string
}
