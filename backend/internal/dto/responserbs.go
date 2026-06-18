package dto

import (
	"time"
)

type ListRBSResponse struct {
	ID                uint   `json:"id"`
	NomorBonSementara string `json:"nomor_bon_sementara"`
	NomorDokumen      string `json:"nomor_dokumen"`
	Nama              string `json:"nama,omitempty"`
	TotalRealisasi    int64  `json:"total_realisasi"`
	TotalEstimasi     int64  `json:"total_estimasi"`
	Selisih           int64  `json:"selisih"`
	Periode           string `json:"periode"`
	Status            string `json:"status"`
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
	NomorDokumen             string
	TanggalTerbit            string
	Periode                  string
	NomorBS                  string
	TanggalDibuat            string
	PathTandaTanganPengaju   string
	NamaPengaju              string
	PathTandaTanganDiketahui string
	NamaDiketahui            string
	PathTandaTanganFinance   string
	NamaFinance              string
	TotalRealisasi           string
	TotalBon                 string
	TotalEstimasi            string
	Selisih                  string
	WilayahDisetujui         string
	Rincian                  []RBSRincianPDF
	RiwayatPersetujuan       []RiwayatApprovaltoPDF
}

type RBSRincianPDF struct {
	Uraian    string
	Tanggal   string
	Kuantitas string
	HargaUnit string
	Total     string
}

type RBSDetailResponse struct {
	NomorDokumen      string                       `json:"nomor_dokumen"`
	NoRefBonSementara string                       `json:"no_ref_bon_sementara"`
	Pemohon           string                       `json:"pemohon"`
	Status            string                       `json:"status"`
	TanggalBerangkat  time.Time                    `json:"tanggal_berangkat"`
	TanggalKedatangan time.Time                    `json:"tanggal_kedatangan"`
	TotalRealisasi    int64                        `json:"total_realisasi"`
	Selisih           int64                        `json:"selisih"`
	UrlBuktiTransfer  *string                      `json:"url_bukti_transfer,omitempty"`
	RiwayatApproval   []RiwayatPersetujuanResponse `json:"riwayat_persetujuan"`
	RincianRealisasi  []RincianRealisasi           `json:"rincian_realisasi"`
}

type RincianRealisasi struct {
	Id               uint      `json:"id"`
	TanggalTranskasi time.Time `json:"tanggal_transaksi"`
	Kategori         string    `json:"kategori"`
	Uraian           string    `json:"uraian"`
	Kuantitas        int       `json:"kuantitas"`
	HargaUnit        int64     `json:"harga_unit"`
	Total            int64     `json:"total"`
	UrlStruk         string    `json:"url_struk"`
}
