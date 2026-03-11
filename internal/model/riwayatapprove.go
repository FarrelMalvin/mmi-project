package model

import "time"

type RiwayatApproval struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	RequestPPDID uint      `gorm:"index" json:"request_ppd_id"`
	UserID       uint      `gorm:"index" json:"user_id"`
	Nama         string    `gorm:"type:varchar(100)" json:"nama"`
	Jabatan      string    `gorm:"type:varchar(20)" json:"jabatan"`
	Tindakan     string    `gorm:"type:varchar(20)" json:"tindakan"`
	Catatan      string    `gorm:"type:text" json:"catatan"`
	CreatedAt    time.Time `json:"created_at"`
}
