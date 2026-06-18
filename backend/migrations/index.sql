CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_request_ppds_periode_berangkat_desc
ON request_ppds (periode_berangkat DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_request_ppds_status_periode
ON request_ppds (status, periode_berangkat DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_request_ppds_user_status_periode
ON request_ppds (user_id, status, periode_berangkat DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_dokumens_ref_type_tipe_refid
ON dokumens (doc_ref_type, tipe_dokumen, doc_ref_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_rbs_periode_berangkat_desc
ON realisasi_bon_sementaras (periode_berangkat DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_rbs_status_periode
ON realisasi_bon_sementaras (status, periode_berangkat DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_rbs_user_status_periode

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_rbs_rincians_rbs_id
ON rbs_rincians (rbs_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ppd_transportasis_request_ppd_id
ON ppd_transportasis (request_ppd_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ppd_rincian_tambahans_request_ppd_id
ON ppd_rincian_tambahans (request_ppd_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ppd_hotels_request_ppd_id
ON ppd_hotels (request_ppd_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_riwayat_approvals_doc_ref
ON riwayat_approvals (doc_ref_type, doc_ref_id);