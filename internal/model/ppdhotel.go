package model

import "time"

type PPDHotel struct {
	Id           uint `gorm:"primaryKey" json:"id"`
	RequestPPDID uint `gorm:"uniqueindex" json:"request_ppd_id"`

	NamaHotel string    `gorm:"type:varchar(100)" json:"nama_hotel"`
	CheckIn   time.Time `json:"check_in"`
	CheckOut  time.Time `json:"check_out"`
	Kategori  string    `gorm:"type:varchar(20)" json:"kategori"`
	Harga     int64     `json:"harga"`
}
