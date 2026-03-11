package model

import "time"

//RBS = Realisasi Bon Sementara
type RBSrincian struct {
	Id    uint `gorm:"primaryKey" json:"id"`
	RBSID uint `gorm:"index" json:"rbs_id"`

	TanggalTransaksi time.Time `json:"tanggal_transaksi"`
	Kuantitas        int       `gorm:"default:1" json:"kuantitas"`
	HargaUnit        int64     `json:"harga_unit"`
	TotalHarga       int64     `json:"total_harga"`
	UrlStruk         string    `gorm:"type:varchar(255)" json:"url_struk"`
}
