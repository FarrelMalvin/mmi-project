package dto

import (
	"golang-mmi/internal/model"
	"time"
)

type ApprovePPDRequest struct {
	RequestPPDID uint   `json:"id"`
	Catatan      string `json:"catatan"`
	UserID       uint   `json:"-"`
	Jabatan      string `json:"-"`
}

type DeclinePPDRequest struct {
	RequestPPDID uint   `json:"id"`
	Catatan      string `json:"catata" validate:"required"`
	UserID       uint   `json:"-"`
	Jabatan      string `json:"-"`
}

type CreatePPDRequest struct {
	Tujuan              string                     `json:"tujuan" validate:"required"` 
	TanggalBerangkat    time.Time                  `json:"tanggal_berangkat" validate:"required"`
	TanggalKembali      time.Time                  `json:"tanggal_kembali" validate:"required"`
	Keperluan           string                     `json:"keperluan" validate:"required"`
	UrlDokumen          string                     `json:"url_dokumen"`
	RincianTambahan     []model.PPDRincianTambahan `json:"rincian_tambahan"`
	RincianTransportasi []model.PPDTransportasi     `json:"rincian_transportasi"`
	RincianHotel        *model.PPDHotel            `json:"rincian_hotel"`
	UserID              uint                       `json:"-"`
	Jabatan             string                     `json:"-"`
}
