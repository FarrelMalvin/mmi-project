package model

type User struct {
	Id             uint   `gorm:"primaryKey" json:"id"`
	Password       string `gorm:"type:varchar(255)" json:"password"`
	Nik            string `gorm:"type:varchar(50)" json:"nik"`
	Nama           string `gorm:"type:varchar(100)" json:"nama"`
	Wilayah        string `gorm:"type:varchar(20)" json:"wilayah"`
	Jabatan        string `gorm:"type:varchar(20)" json:"jabatan"`
	Departemen     string `gorm:"type:varchar(20)" json:"departemen"`
	UrlTandaTangan string `gorm:"type:varchar(255)" json:"url_tanda_tangan"`
	AtasanID       *uint  `gorm:"index" json:"atasan_id"`
	Atasan         *User  `gorm:"foreignKey:AtasanID" json:"atasan,omitempty"`
	TokenVersion   int    `gorm:"default:1" json:"-"`
}
