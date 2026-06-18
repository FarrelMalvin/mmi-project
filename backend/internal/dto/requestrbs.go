package dto

import (
	"time"
)

type CreateRBSRequest struct {
	RequestPPDID      uint      `json:"request_ppd_id"`
	TotalRealisasi    int64     `json:"total_realisasi"`
	Selisih           int64     `json:"selisih"`
	PeriodeBerangkat  time.Time `json:"periode_berangkat" validate:"required"`
	NomorBonSementara string    `json:"nomor_bon_sementara" validate:"required"`
	PeriodeKembali    time.Time `json:"periode_kembali" validate:"required"`
	BuktiTransferField *string   `json:"bukti_transfer_field"`
	UrlBuktiTransfer  *string   `json:"url_bukti_transfer"`

	Items []RBSItemRequest `json:"items"`

	UserID  uint   `json:"-"`
	Jabatan string `json:"-"`
}

type RBSItemRequest struct {
	Uraian     string    `json:"uraian"`
	Tanggal    time.Time `json:"tanggal"`
	Kuantitas  int       `json:"kuantitas"`
	HargaUnit  int64     `json:"harga_unit"`
	Kategori   string    `json:"kategori"`
	Total      int64     `json:"total"`
	StrukField string    `json:"struk_field"`
	UrlStruk   string    `json:"url_struk"`
}

type ApproveRBSRequest struct {
	RealisasiBonID uint   `param:"id" json:"-"`
	Catatan        string `json:"catatan"`
	UserID         uint   `json:"-"`
	Jabatan        string `json:"-"`
}

type DeclineRBSRequest struct {
	RealisasiBonID uint   `param:"id" json:"-"`
	Catatan        string `json:"catatan"`
	UserID         uint   `json:"-"`
	Jabatan        string `json:"-"`
}

type RBSListRequest struct {
	UserID  uint   `json:"-"`
	Jabatan string `json:"-"`
	Page    int    `query:"page"`
	Limit   int    `query:"limit"`
	Month   int    `query:"month"`
	Year    int    `query:"year"`
}

type RBSUpdateRequest struct {
	RincianTambahan []RincianRealisasi `json:"rincian_tambahan"`
}
