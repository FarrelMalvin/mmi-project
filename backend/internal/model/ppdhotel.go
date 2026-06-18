package model

import "time"

type PPDHotel struct {
    Id           uint `gorm:"primaryKey" json:"id"`
    RequestPPDID uint `gorm:"uniqueIndex;not null" json:"request_ppd_id"`

    NamaHotel     string    `gorm:"type:varchar(100);not null" json:"nama_hotel"`
    CheckIn       time.Time `gorm:"not null" json:"check_in"`
    CheckOut      time.Time `gorm:"not null" json:"check_out"`
    HargaPerMalam int64     `gorm:"not null;default:0" json:"harga_per_malam"`
    HargaTotal    int64     `gorm:"not null;default:0" json:"harga_total"`
}
