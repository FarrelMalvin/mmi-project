package model

import "time"

type RealisasiBonSementara struct {
	Id           uint `gorm:"primarykey" json:"id"`
	RequestPPDID uint `gorm:"uniqueindex" json:"req"`
	UserID       uint `gorm:"index" json:"user_id"`

	NomorBonSementara string    `gorm:"type:varchar(50)" json:"nomor_bon_sementara"`
	TotalRealisasi    int64     `json:"total_realisasi"`
	Selisih           int64     `json:"selisih"`
	Status            string    `gorm:"type:varchar(20)" json:"status"`
	PeriodeBerangkat  time.Time `json:"periode_berangkat"`
	PeriodeKembali    time.Time `json:"periode_kembali"`
	UrlBuktiTransfer  *string   `gorm:"type:varchar(255)" json:"url_bukti_transfer"`
	CreatedAt         time.Time `json:"created_at"`

	RBSrincian         []RBSrincian      `gorm:"foreignKey:RBSID" json:"rbs_rincian"`
	RiwayatPersetujuan []RiwayatApproval `gorm:"polymorphic:DocRef;polymorphicValue:RealisasiBonSementara" json:"riwayat_persetujuan"`
	Dokumen            []Dokumen         `gorm:"polymorphic:DocRef;polymorphicValue:RealisasiBonSementara" json:"dokumen_terdaftar"`
}

type RBSListView struct {
	Id             uint   `gorm:"id"`
	Nama           string `gorm:"nama"`
	NomorDokumen   string `gorm:"nomor_dokumen"`
	TotalRealisasi int64  `gorm:"total_realisasi"`
	TotalEstimasi  int64  `gorm:"total_estimasi"`
	Selisih        int64  `gorm:"selisih"`
	Status         string `gorm:"status"`
}

type RBSDataforCsv struct {
	NomorReferensiBS string    `gorm:"nomor_referensi_bs"`
	Nama             string    `gorm:"nama"`
	Periode          time.Time `gorm:"periode"`
	Kategori         string    `gorm:"kategori"`
	TanggalTransaksi time.Time `gorm:"tanggal_transaksi"`
	Uraian           string    `gorm:"uraian"`
	Kuantitas        int       `gorm:"kuantitas"`
	HargaUnit        int64     `gorm:"harga_unit"`
	TotalHarga       int64     `gorm:"total_harga"`
}

type KategoriData struct {
	NamaKategori string
	Items        []RBSDataforCsv
	Total        int64
}
