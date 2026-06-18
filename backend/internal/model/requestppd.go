package model

import (
	"time"
)

type RequestPPD struct {
	Id     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"index;not null" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user"`

	Tujuan           string    `gorm:"type:varchar(100);not null" json:"tujuan"`
	PeriodeBerangkat time.Time `gorm:"index;not null" json:"periode_berangkat"`
	PeriodeKembali   time.Time `gorm:"not null" json:"periode_kembali"`
	Keperluan        string    `gorm:"type:text;not null" json:"keperluan"`
	Status           string    `gorm:"type:varchar(50);not null;index" json:"status"`
	TotalEstimasi    int64     `gorm:"not null;default:0" json:"total_estimasi"`
	UrlDokumen       string    `gorm:"type:varchar(255)" json:"url_dokumen"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	RincianTambahan     []PPDRincianTambahan `gorm:"foreignKey:RequestPPDID" json:"rincian_tambahan"`
	RincianHotel        *PPDHotel            `gorm:"foreignKey:RequestPPDID" json:"rincian_hotel,omitempty"`
	RincianTransportasi []PPDTransportasi    `gorm:"foreignKey:RequestPPDID" json:"rincian_transportasi"`

	RealisasiBonSementara *RealisasiBonSementara `gorm:"foreignKey:RequestPPDID" json:"realisasi_bon_sementara,omitempty"`
	RiwayatPersetujuan    []RiwayatApproval      `gorm:"polymorphic:DocRef;polymorphicValue:RequestPPD" json:"riwayat_persetujuan"`
	Dokumen               []Dokumen              `gorm:"polymorphic:DocRef;polymorphicValue:RequestPPD" json:"dokumen_terdaftar"`
}

type PPDListView struct {
	ID               uint      `gorm:"column:id"`
	Nama             string    `gorm:"column:nama"`
	NomorTipeDokumen string    `gorm:"column:nomor_tipe_dokumen"`
	NomorDokumen     string    `gorm:"column:nomor_dokumen"`
	Tujuan           string    `gorm:"column:tujuan"`
	Keperluan        string    `gorm:"column:keperluan"`
	TotalEstimasi    int64     `gorm:"column:total_estimasi"`
	Status           string    `gorm:"column:status"`
	PeriodeBerangkat time.Time `gorm:"column:periode_berangkat"`
}

type PPDItemView struct {
	ID        uint   `gorm:"column:id"`
	Uraian    string `gorm:"column:uraian"`
	Kuantitas int    `gorm:"column:kuantitas"`
	Kategori  string `gorm:"column:kategori"`
	HargaUnit int64  `gorm:"column:harga_unit"`
	Total     int64  `gorm:"column:total"`
}

type DropdownPPDView struct {
	ID               uint      `gorm:"column:id"`
	NomorDokumen     string    `gorm:"column:nomor_dokumen"`
	Tujuan           string    `gorm:"column:tujuan"`
	PeriodeBerangkat time.Time `gorm:"column:periode_berangkat"`
	PeriodeKembali   time.Time `gorm:"column:periode_kembali"`
}
