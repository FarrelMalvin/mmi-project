package model

import "time"

type RiwayatApproval struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DocRefID   uint      `gorm:"index" json:"doc_ref_id"`
	DocRefType string    `gorm:"index;type:varchar(50)" json:"doc_ref_type"`
	UserID     uint      `gorm:"index" json:"user_id"`
	Nama       string    `gorm:"type:varchar(100)" json:"nama"`
	Jabatan    string    `gorm:"type:varchar(20)" json:"jabatan"`
	Tindakan   string    `gorm:"type:varchar(20)" json:"tindakan"`
	Catatan    string    `gorm:"type:text" json:"catatan"`
	CreatedAt  time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user"`
}
