package model

import "time"

type PPDTransportasi struct {
	Id           uint `gorm:"primaryKey" json:"id"`
	RequestPPDID uint `gorm:"uniqueindex" json:"request_ppd_id"`

	TipePerjalanan    string    `gorm:"type:varchar(10)" json:"tipe_perjalanan"`
	KotaAsal          string    `gorm:"type:varchar(20)" json:"kota_asal"`
	KotaTujuan        string    `gorm:"type:varchar(20)" json:"kota_tujuan"`
	JenisTransportasi string    `gorm:"type:varchar(20)" json:"jenis_transportasi"`
	NomorKendaraan    *string   `gorm:"type:varchar(20)" json:"nomor_kendaraan"`
	Harga             int64     `json:"harga"`
	Kategori          string    `gorm:"type:varchar(20)" json:"kategori"`
	Jamberangkat      time.Time `json:"jam_berangkat"`
}
