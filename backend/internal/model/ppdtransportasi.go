package model

import "time"

type PPDTransportasi struct {
    Id           uint `gorm:"primaryKey" json:"id"`
    RequestPPDID uint `gorm:"index;not null" json:"request_ppd_id"`

    TipePerjalanan    string    `gorm:"type:varchar(30);not null" json:"tipe_perjalanan"`
    KotaAsal          string    `gorm:"type:varchar(100);not null" json:"kota_asal"`
    KotaTujuan        string    `gorm:"type:varchar(100);not null" json:"kota_tujuan"`
    JenisTransportasi string    `gorm:"type:varchar(50);not null" json:"jenis_transportasi"`
    NomorKendaraan    *string   `gorm:"type:varchar(30)" json:"nomor_kendaraan"`
    Harga             int64     `gorm:"not null;default:0" json:"harga"`
    JamBerangkat      time.Time `json:"jam_berangkat"`
}