package model

type PPDRincianTambahan struct {
	Id           uint `gorm:"primaryKey" json:"id"`
	RequestPPDID uint `gorm:"foreignKey" json:"request_ppd_id"`

	Harga      int64  `json:"harga"`
	Kuantitas  int    `gorm:"default:1" json:"kuantitas"`
	Keterangan string `gorm:"type:varchar(100)" json:"keterangan"`
	Kategori   string `gorm:"type:varchar(20)" json:"kategori"`
}
