package dto

type CreateUserRequest struct {
	Nama           string `json:"nama"`
	Jabatan        string `json:"jabatan"`
	Wilayah        string `json:"wilayah"`
	Departemen     string `json:"departemen"`
	Password       string `json:"password"`
	JabatanRequest string `json:"-"`
}
