package model

import (
	"time"
)

type RequestPPD struct {
	Id     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"index" json:"user_id"`

	Tujuan           string    `gorm:"type:varchar(100);not null" json:"tujuan"`
	PeriodeBerangkat time.Time `gorm:"index" json:"periode_berangkat"`
	PeriodeKembali   time.Time `json:"periode_kembali"`
	Keperluan        string    `gorm:"type:text" json:"keperluan"`
	Status           string    `gorm:"type:varchar(20);default:'pending';index" json:"status"`
	TotalEstimasi    int64     `json:"total_estimasi"`
	Kuantitas        int       `gorm:"default:1" json:"quantitas"`
	UrlDokumen       string    `gorm:"type:varchar(255)" json:"url_dokumen"`

	RincianTambahan       []PPDRincianTambahan  `gorm:"foreignKey:RequestPPDID" json:"rincian_tambahan"`
	RincianHotel          *PPDHotel             `gorm:"foreignKey:RequestPPDID" json:"rincian_hotel"`
	RincianTransportasi   *PPDTransportasi      `gorm:"foreignKey:RequestPPDID" json:"rincian_transportasi"`
	RealisasiBonSementara RealisasiBonSementara `gorm:"foreignKey:RequestPPDID" json:"realisasi_bon_sementara"`
	RiwayatPersetujuan    []RiwayatApproval     `gorm:"foreignKey:RequestPPDID" json:"riwayat_persetujuan"`
	Dokumen               []Dokumen             `gorm:"polymorphic:DocRef;polymorphicValue:RequestPPD" json:"dokumen_terdaftar"`
}
