package dto

import (
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
	Catatan      string `json:"catatan" validate:"required"`
	UserID       uint   `json:"-"`
	Jabatan      string `json:"-"`
}

type CreatePPDRequest struct {
	Tujuan              string                    `json:"tujuan" validate:"required"`
	TanggalBerangkat    time.Time                 `json:"tanggal_berangkat" validate:"required"`
	TanggalKembali      time.Time                 `json:"tanggal_kembali" validate:"required"`
	Keperluan           string                    `json:"keperluan" validate:"required"`
	UrlDokumen          string                    `json:"url_dokumen"`
	RincianTambahan     []PPDRincianTambahan      `json:"rincian_tambahan"`
	RincianTransportasi []PPDRincianTransportasi `json:"rincian_transportasi"`
	RincianHotel        *PPDRincianHotel          `json:"rincian_hotel"`
	UserID              uint                      `json:"-"`
	Jabatan             string                    `json:"-"`
}

type UpdatePPDRequest struct {
	RincianTambahan     []PPDRincianTambahan      `json:"rincian_tambahan"`
	RincianTransportasi []PPDRincianTransportasi `json:"rincian_transportasi"`
	RincianHotel        *PPDRincianHotel          `json:"rincian_hotel"`
}

type PPDRincianTambahan struct {
	ID         uint   `json:"id"`
	Harga      int64  `json:"harga"`
	Kuantitas  int    `json:"kuantitas"`
	Keterangan string `json:"keterangan"`
	Kategori   string `json:"kategori"`
}

type PPDRincianTransportasi struct {
	ID                uint      `json:"id"`
	TipePerjalanan    string    `json:"tipe_perjalanan"`
	Harga             int64     `json:"harga"`
	Kategori          string    `json:"kategori"`
	KotaAsal          string    `json:"kota_asal"`
	KotaTujuan        string    `json:"kota_tujuan"`
	JamBerangkat      time.Time `json:"jam_berangkat"`
	JenisTransportasi string    `json:"jenis_transportasi"`
	NomorKendaraan    *string   `json:"nomor_kendaraan"`
}

type PPDRincianHotel struct {
	ID          uint      `json:"id"`
	NamaHotel   string    `json:"nama_hotel"`
	LokasiHotel string    `json:"lokasi_hotel"`
	HargaPerMalam int64     `json:"harga_per_malam"`
	HargaTotal    int64     `json:"harga_total"`
	CheckIn     time.Time `json:"check_in"`
	CheckOut    time.Time `json:"check_out"`
}
