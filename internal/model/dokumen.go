package model

import "time"

type Dokumen struct {
	Id         uint   `gorm:"primaryKey" json:"id"`
	DocRefID   uint   `gorm:"index" json:"doc_ref_id"`
	DocRefType string `gorm:"index;type:varchar(50)" json:"doc_ref_type"`
	UserID     uint   `gorm:"index" json:"user_id"`

	NomorDokumen     string    `gorm:"type:varchar(17)" json:"nomor_dokumen"`
	TipeDokumen      string    `gorm:"type:varchar(20)" json:"tipe_dokumen"`
	NomorTipeDokumen string    `gorm:"type:varchar(20)" json:"nomor_tipe_dokumen"`
	CreatedAt        time.Time `json:"created_at"`
}
