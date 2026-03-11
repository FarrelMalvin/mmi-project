package model

type ApprovePPDRequest struct {
	Catatan string `json:"catatan"`
}

type DeclinePPDRequest struct {
	Catatan string `json:"catatan"`
}

type GetListRBSRequest struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
	Bulan int `query:"bulan"`
	Tahun int `query:"tahun"`
}

type CreateRBSRequest struct {
	RequestPPDID uint            `json:"req"`
	RBSrincian   []RBSrincianDTO `json:"rbs_rincian"`
}

type RBSrincianDTO struct {
	Keterangan string `json:"keterangan"`
	Harga      int64  `json:"harga"`
	Jumlah     int    `json:"jumlah"`
}
