export interface PPDTransportasi {
  id?: number;
  request_ppd_id?: number;
  tipe_perjalanan: string;
  kota_asal: string;
  kota_tujuan: string;
  jenis_transportasi: string;
  nomor_kendaraan?: string | null;
  harga: number;
  kategori: string;
  jam_berangkat: string;  
}

export interface PPDRincianTambahan {
  id?: number;
  request_ppd_id?: number;
  harga: number;
  kuantitas: number;
  keterangan: string;
  kategori: string;
}

export interface PPDHotel {
  id?: number;
  request_ppd_id?: number;
  nama_hotel: string;
  periode_hotel: string; 
  check_in: string;      
  check_out: string;     
  kategori: string;
  harga: number;
}

export interface RBSItemRequest {
  uraian: string;
  tanggal: string;
  qty: number;
  harga_unit: number;
  kategori: string;
  total: number;
  url_struk: string;
}

// REQUEST DTO

export interface CreatePPDRequest {
  tujuan: string;
  tanggal_berangkat: string; 
  tanggal_kembali: string;   
  keperluan: string;
  url_dokumen?: string;
  rincian_tambahan?: PPDRincianTambahan[];
  rincian_transportasi?: PPDTransportasi[];
  rincian_hotel?: PPDHotel | null;
}

export interface CreateRBSRequest {
  request_ppd_id: number;
  total_realisasi: number;
  selisih: number;
  periode_berangkat: string; 
  nomor_bon_sementara: string;
  periode_kembali: string;   
  url_bukti_transfer?: string | null; 
  items: RBSItemRequest[];
}

// RESPONSE
export interface PerjalananDinas {
  id: number;
  nama?: string;             
  nomor_tipe_dokumen?: string; 
  tujuan: string;
  keperluan: string;
  total_estimasi: number;
  status: string;
  periode_berangkat: string;   
  is_downloadable: boolean;
}