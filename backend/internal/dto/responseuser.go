package dto

type ProfileResponse struct{
	Nama string `json:"nama"`
	Wilayah string `json:"wilayah"`
	Jabatan string `json:"jabatan"`
	Departemen string `json:"departemen"`
	PathTandaTangan string `json:"path_tanda_tangan"`
}