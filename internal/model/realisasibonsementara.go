package model

import "time"

type RealisasiBonSementara struct {
	Id           uint `gorm:"primarykey" json:"id"`
	RequestPPDID uint `gorm:"uniqueindex" json:"req"`
	UserID       uint `gorm:"index" json:"user_id"`

	TotalRealisasi   int64     `json:"total_realisasi"`
	Selisih          int64     `json:"selisih"`
	Status           string    `gorm:"type:varchar(20)" json:"status"`
	Periode          time.Time `json:"periode"`
	UrlBuktiTransfer *string   `gorm:"type:varchar(255)" json:"url_bukti_transfer"`

	RBSrincian []RBSrincian `gorm:"foreignKey:RBSID" json:"rbs_rincian"`
	Dokumen    []Dokumen    `gorm:"polymorphic:DocRef;polymorphicValue:RealisasiBonSementara" json:"dokumen_terdaftar"`
}
