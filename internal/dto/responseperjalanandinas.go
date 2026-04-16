package dto

import (
	"time"
)

type ListPPDResponse struct {
	ID               uint      `json:"id"`
	Nama             string    `json:"nama,omitempty"`
	NomorTipeDokumen string    `json:"nomor_tipe_dokumen,omitempty"`
	Tujuan           string    `json:"tujuan"`
	Keperluan        string    `json:"keperluan"`
	TotalEstimasi    int       `json:"total_estimasi"`
	Status           string    `json:"status"`
	PeriodeBerangkat time.Time `json:"periode_berangkat"`
	IsDownloadable   bool      `json:"is_downloadable"`
}

type ListPPDRequest struct {
	Page    int    `query:"page"`
	Limit   int    `query:"limit"`
	UserID  uint   `json:"-"`
	Jabatan string `json:"-"`
}

type DropdownPPDResponse struct {
	ID               uint   `json:"id"`
	NomorDokumen     string `json:"nomor_dokumen"`
	PeriodeKembali   string `json:"periode_kembali"`
	PeriodeBerangkat string `json:"periode_berangkat"`
	Tujuan           string `json:"tujuan"`
}

type DetailPPDResponse struct {
	ID               uint      `json:"id"`
	Tujuan           string    `json:"tujuan"`
	NomorTipeDokumen string    `json:"nomor_tipe_dokumen,omitempty"`
	PeriodeBerangkat time.Time `json:"periode_berangkat"`
	PeriodeKembali   time.Time `json:"periode_kembali"`
	Keperluan        string    `json:"keperluan"`
	Status           string    `json:"status"`
	TotalEstimasi    int64     `json:"total_estimasi"`
	UrlDokumen       string    `json:"url_dokumen"`

	RincianTambahan     []RincianTambahanResponse    `json:"rincian_tambahan"`
	RincianTransportasi []RincianTransportResponse   `json:"rincian_transportasi"`
	RiwayatApproval     []RiwayatPersetujuanResponse `json:"riwayat_persetujuan"`
	RincianHotel        *RincianHotelResponse        `json:"rincian_hotel"`
}

type RincianTambahanResponse struct {
	Harga      int64  `json:"harga"`
	Kuantitas  int    `json:"kuantitas"`
	Keterangan string `json:"keterangan"`
	Kategori   string `json:"kategori"`
}

type RincianTransportResponse struct {
	TipePerjalanan    string    `json:"tipe_perjalanan"`
	KotaAsal          string    `json:"kota_asal"`
	KotaTujuan        string    `json:"kota_tujuan"`
	JenisTransportasi string    `json:"jenis_transportasi"`
	NomorKendaraan    *string   `json:"nomor_kendaraan"`
	Harga             int64     `json:"harga"`
	Kategori          string    `json:"kategori"`
	JamBerangkat      time.Time `json:"jam_berangkat"`
}

type RincianHotelResponse struct {
	NamaHotel     string `json:"nama_hotel"`
	Kategori      string `json:"kategori"`
	HargaPerMalam int64  `json:"harga_per_malam"`
	TotalHarga    int64  `json:"total_harga"`
}

type DokumenTerdaftarResponse struct {
	NamaPengaju      string `json:"nama_pengaju"`
	NomorDokumen     string `gorm:"type:varchar(50)" json:"nomor_dokumen"`
	TipeDokumen      string `gorm:"type:varchar(50)" json:"tipe_dokumen"`
	NomorTipeDokumen string `gorm:"type:varchar(50)" json:"nomor_tipe_dokumen"`
}

type RiwayatPersetujuanResponse struct {
	Nama     string `json:"nama"`
	Jabatan  string `json:"jabatan"`
	Tindakan string `json:"tindakan"`
	Catatan  string `json:"catatan"`
}

type PPDItemResponse struct {
	ID        uint   `json:"id"`
	Uraian    string `json:"uraian"`
	Qty       int    `json:"qty"`
	HargaUnit int64  `json:"harga_unit"`
	Kategori  string `json:"kategori"`
	Total     int64  `json:"total"`
}

type PPDItemDetailResponse struct {
	NomorReferensi string            `json:"nomor_referensi"`
	Periode        string            `json:"periode"`
	Items          []PPDItemResponse `json:"items"`
}

type PPDDataToPDF struct {
	NomorDokumen                   string
	TanggalTerbit                  string
	Nama                           string
	Jabatan                        string
	Nik                            string
	Tujuan                         string
	Wilayah                        string
	Keperluan                      string
	Periode                        string
	NamaHotel                      string
	PeriodeHotel                   string
	TujuanHotel                    string
	CheckIn                        string
	CheckOut                       string
	HargaHotel                     string
	AsalKeberangkatan              string
	TujuanKeberangkatan            string
	JamBerangkatKeberangkatan      string
	JenisTransportasiKeberangkatan string
	NomorKendaraanKeberangkatan    string
	AsalKedatangan                 string
	TujuanKedatangan               string
	JenisTransportasiKedatangan    string
	NomorKendaraanKedatangan       string
	JamBerangkatKedatangan         string
	PathTandaTanganPengaju         string
	PathTandaTanganAtasan          string
	PathTandaTanganHRGA            string
	TanggalDisetujuiAtasan         string
	TanggalDisetujuiHRGA           string
	TanggalDiajukan                 string
	Rincian                        []PPDDataToPDFDetail
	RiwayatPersetujuan             []RiwayatApprovaltoPDF
}

type PPDDataToPDFDetail struct {
	Kategori string
	Harga    string
	Total    string
}

type RiwayatApprovaltoPDF struct {
	PathTandaTangan string
	Jabatan         string
	Tindakan        string
	Nama            string
}

type BSDataToPDF struct {
	NomorDokumen            string
	NomorTipeDokumen        string
	TanggalTerbit           string
	TanggalPengajuan        string
	TanggalPenyelesaian     string
	Nama                    string
	Keperluan               string
	Jumlah                  string
	PathTandaTanganHRGA     string
	NamaHRGA                string
	PathTandaTanganDirektur string
	NamaDirektur            string
	PathTandaTanganFinance  string
	NamaFinance             string
	RiwayatPersetujuan      []RiwayatApprovaltoPDF
}
