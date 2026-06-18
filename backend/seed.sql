BEGIN;

-- =========================================================
-- CLEAN EXISTING STAGING DATA
-- =========================================================
TRUNCATE TABLE
  blacklisted_tokens,
  refresh_tokens,
  dokumens,
  riwayat_approvals,
  rbs_rincians,
  realisasi_bon_sementaras,
  ppd_transportasis,
  ppd_rincian_tambahans,
  ppd_hotels,
  request_ppds,
  users
RESTART IDENTITY CASCADE;

-- =========================================================
-- USERS
-- Password dummy bcrypt ini bisa kamu sesuaikan.
-- Role distribution:
-- 1-5     Atasan
-- 6-90    Pegawai
-- 91-94   HRGA
-- 95-97   Finance
-- 98-100  Direktur
-- =========================================================
INSERT INTO users (
  id,
  password,
  nik,
  nama,
  wilayah,
  jabatan,
  departemen,
  path_tanda_tangan,
  atasan_id,
  token_version
)
SELECT
  gs AS id,
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy' AS password,
  'NIK' || LPAD(gs::text, 5, '0') AS nik,
  CASE
    WHEN gs BETWEEN 1 AND 5 THEN 'Atasan ' || gs
    WHEN gs BETWEEN 91 AND 94 THEN 'HRGA ' || gs
    WHEN gs BETWEEN 95 AND 97 THEN 'Finance ' || gs
    WHEN gs BETWEEN 98 AND 100 THEN 'Direktur ' || gs
    ELSE 'Pegawai ' || gs
  END AS nama,
  CASE
    WHEN gs % 3 = 0 THEN 'Jakarta'
    WHEN gs % 3 = 1 THEN 'Surabaya'
    ELSE 'Bandung'
  END AS wilayah,
  CASE
    WHEN gs BETWEEN 1 AND 5 THEN 'Atasan'
    WHEN gs BETWEEN 91 AND 94 THEN 'HRGA'
    WHEN gs BETWEEN 95 AND 97 THEN 'Finance'
    WHEN gs BETWEEN 98 AND 100 THEN 'Direktur'
    ELSE 'Pegawai'
  END AS jabatan,
  CASE
    WHEN gs % 4 = 0 THEN 'GA'
    WHEN gs % 4 = 1 THEN 'Sales'
    WHEN gs % 4 = 2 THEN 'Finance'
    ELSE 'HR'
  END AS departemen,
  '/public/signature/signature_' || gs || '.png' AS path_tanda_tangan,
  CASE
    WHEN gs BETWEEN 6 AND 90 THEN ((gs % 5) + 1)
    ELSE NULL
  END AS atasan_id,
  1 AS token_version
FROM generate_series(1, 100) AS gs;

-- =========================================================
-- REQUEST PPD
-- 2.000 rows
-- =========================================================
INSERT INTO request_ppds (
  id,
  user_id,
  tujuan,
  periode_berangkat,
  periode_kembali,
  keperluan,
  status,
  total_estimasi,
  url_dokumen,
  created_at,
  updated_at
)
SELECT
  gs AS id,
  ((gs % 85) + 6) AS user_id,
  CASE
    WHEN gs % 5 = 0 THEN 'Jakarta'
    WHEN gs % 5 = 1 THEN 'Surabaya'
    WHEN gs % 5 = 2 THEN 'Bandung'
    WHEN gs % 5 = 3 THEN 'Semarang'
    ELSE 'Yogyakarta'
  END AS tujuan,
  (DATE '2026-01-01' + ((gs % 150) || ' days')::interval) AS periode_berangkat,
  (DATE '2026-01-01' + ((gs % 150) || ' days')::interval + interval '2 days') AS periode_kembali,
  'Keperluan perjalanan dinas nomor ' || gs AS keperluan,
  CASE
    WHEN gs % 10 IN (0, 1) THEN 'Menunggu Atasan'
    WHEN gs % 10 IN (2, 3) THEN 'Menunggu HRGA'
    WHEN gs % 10 = 4 THEN 'Menunggu Direktur'
    WHEN gs % 10 IN (5, 6) THEN 'Menunggu Finance'
    WHEN gs % 10 IN (7, 8) THEN 'Selesai'
    ELSE 'Ditolak Atasan'
  END AS status,
  0 AS total_estimasi,
  '' AS url_dokumen,
  NOW() - ((gs % 60) || ' days')::interval AS created_at,
  NOW() - ((gs % 30) || ' days')::interval AS updated_at
FROM generate_series(1, 2000) AS gs;

-- =========================================================
-- PPD TRANSPORTASI - KEBERANGKATAN
-- =========================================================
INSERT INTO ppd_transportasis (
  request_ppd_id,
  tipe_perjalanan,
  kota_asal,
  kota_tujuan,
  jenis_transportasi,
  nomor_kendaraan,
  harga,
  jam_berangkat
)
SELECT
  p.id AS request_ppd_id,
  'Keberangkatan' AS tipe_perjalanan,
  'Surabaya' AS kota_asal,
  p.tujuan AS kota_tujuan,
  CASE
    WHEN p.id % 4 = 0 THEN 'Pesawat'
    WHEN p.id % 4 = 1 THEN 'Kereta Api'
    WHEN p.id % 4 = 2 THEN 'Kendaraan Dinas'
    ELSE 'Bus'
  END AS jenis_transportasi,
  CASE
    WHEN p.id % 4 = 2 THEN 'L ' || (1000 + p.id)::text || ' AB'
    ELSE NULL
  END AS nomor_kendaraan,
  (150000 + ((p.id % 10) * 25000))::bigint AS harga,
  p.periode_berangkat + interval '08 hours' AS jam_berangkat
FROM request_ppds p;

-- =========================================================
-- PPD TRANSPORTASI - KEDATANGAN
-- =========================================================
INSERT INTO ppd_transportasis (
  request_ppd_id,
  tipe_perjalanan,
  kota_asal,
  kota_tujuan,
  jenis_transportasi,
  nomor_kendaraan,
  harga,
  jam_berangkat
)
SELECT
  p.id AS request_ppd_id,
  'Kedatangan' AS tipe_perjalanan,
  p.tujuan AS kota_asal,
  'Surabaya' AS kota_tujuan,
  CASE
    WHEN p.id % 4 = 0 THEN 'Pesawat'
    WHEN p.id % 4 = 1 THEN 'Kereta Api'
    WHEN p.id % 4 = 2 THEN 'Kendaraan Dinas'
    ELSE 'Bus'
  END AS jenis_transportasi,
  CASE
    WHEN p.id % 4 = 2 THEN 'L ' || (2000 + p.id)::text || ' CD'
    ELSE NULL
  END AS nomor_kendaraan,
  (150000 + ((p.id % 10) * 25000))::bigint AS harga,
  p.periode_kembali + interval '17 hours' AS jam_berangkat
FROM request_ppds p;

-- =========================================================
-- PPD HOTEL
-- 70% PPD punya hotel
-- =========================================================
INSERT INTO ppd_hotels (
  request_ppd_id,
  nama_hotel,
  check_in,
  check_out,
  harga_per_malam,
  harga_total
)
SELECT
  p.id AS request_ppd_id,
  'Hotel Staging ' || p.id AS nama_hotel,
  p.periode_berangkat AS check_in,
  p.periode_kembali AS check_out,
  (250000 + ((p.id % 8) * 50000))::bigint AS harga_per_malam,
  ((250000 + ((p.id % 8) * 50000)) * 2)::bigint AS harga_total
FROM request_ppds p
WHERE p.id % 10 <> 0;

-- =========================================================
-- PPD RINCIAN TAMBAHAN
-- 3 rows per PPD
-- =========================================================
INSERT INTO ppd_rincian_tambahans (
  request_ppd_id,
  harga,
  kuantitas,
  keterangan,
  kategori
)
SELECT
  p.id AS request_ppd_id,
  50000::bigint AS harga,
  2 AS kuantitas,
  'Konsumsi hari pertama' AS keterangan,
  'Konsumsi' AS kategori
FROM request_ppds p;

INSERT INTO ppd_rincian_tambahans (
  request_ppd_id,
  harga,
  kuantitas,
  keterangan,
  kategori
)
SELECT
  p.id AS request_ppd_id,
  30000::bigint AS harga,
  1 AS kuantitas,
  'Parkir dan tol' AS keterangan,
  'Parkir' AS kategori
FROM request_ppds p;

INSERT INTO ppd_rincian_tambahans (
  request_ppd_id,
  harga,
  kuantitas,
  keterangan,
  kategori
)
SELECT
  p.id AS request_ppd_id,
  75000::bigint AS harga,
  1 AS kuantitas,
  'Biaya lain-lain' AS keterangan,
  'Lain-lain' AS kategori
FROM request_ppds p;

-- =========================================================
-- UPDATE TOTAL ESTIMASI PPD
-- =========================================================
UPDATE request_ppds p
SET total_estimasi =
  COALESCE((
    SELECT SUM(t.harga)
    FROM ppd_transportasis t
    WHERE t.request_ppd_id = p.id
  ), 0)
  +
  COALESCE((
    SELECT SUM(h.harga_total)
    FROM ppd_hotels h
    WHERE h.request_ppd_id = p.id
  ), 0)
  +
  COALESCE((
    SELECT SUM(rt.harga * rt.kuantitas)
    FROM ppd_rincian_tambahans rt
    WHERE rt.request_ppd_id = p.id
  ), 0);

-- =========================================================
-- REALISASI BON SEMENTARA
-- 1.200 rows
-- request_ppd_id unique, pakai PPD id 1..1200
-- =========================================================
INSERT INTO realisasi_bon_sementaras (
  id,
  request_ppd_id,
  user_id,
  nomor_bon_sementara,
  total_realisasi,
  selisih,
  status,
  periode_berangkat,
  periode_kembali,
  url_bukti_transfer,
  created_at
)
SELECT
  p.id AS id,
  p.id AS request_ppd_id,
  p.user_id AS user_id,
  'BS/GA/05/2026/' || LPAD(p.id::text, 4, '0') AS nomor_bon_sementara,
  0 AS total_realisasi,
  0 AS selisih,
  CASE
    WHEN p.id % 7 IN (0, 1) THEN 'Menunggu Atasan'
    WHEN p.id % 7 = 2 THEN 'Menunggu HRGA'
    WHEN p.id % 7 = 3 THEN 'Menunggu Direktur'
    WHEN p.id % 7 IN (4, 5) THEN 'Menunggu Finance'
    ELSE 'Selesai'
  END AS status,
  p.periode_berangkat,
  p.periode_kembali,
  CASE
    WHEN p.id % 4 = 0 THEN '/public/struk/bukti_transfer_' || p.id || '.jpg'
    ELSE NULL
  END AS url_bukti_transfer,
  NOW() - ((p.id % 40) || ' days')::interval AS created_at
FROM request_ppds p
WHERE p.id BETWEEN 1 AND 1200;

-- =========================================================
-- RBS RINCIAN
-- 4 rows per RBS
-- =========================================================
INSERT INTO rbs_rincians (
  rbs_id,
  tanggal_transaksi,
  kuantitas,
  uraian,
  kategori,
  harga_unit,
  total_harga,
  url_struk
)
SELECT
  r.id AS rbs_id,
  r.periode_berangkat AS tanggal_transaksi,
  1 AS kuantitas,
  'Makan pagi' AS uraian,
  'Konsumsi' AS kategori,
  25000::bigint AS harga_unit,
  25000::bigint AS total_harga,
  '/public/struk/struk_' || r.id || '_1.jpg' AS url_struk
FROM realisasi_bon_sementaras r;

INSERT INTO rbs_rincians (
  rbs_id,
  tanggal_transaksi,
  kuantitas,
  uraian,
  kategori,
  harga_unit,
  total_harga,
  url_struk
)
SELECT
  r.id AS rbs_id,
  r.periode_berangkat + interval '1 day' AS tanggal_transaksi,
  1 AS kuantitas,
  'Makan siang' AS uraian,
  'Konsumsi' AS kategori,
  35000::bigint AS harga_unit,
  35000::bigint AS total_harga,
  '/public/struk/struk_' || r.id || '_2.jpg' AS url_struk
FROM realisasi_bon_sementaras r;

INSERT INTO rbs_rincians (
  rbs_id,
  tanggal_transaksi,
  kuantitas,
  uraian,
  kategori,
  harga_unit,
  total_harga,
  url_struk
)
SELECT
  r.id AS rbs_id,
  r.periode_berangkat + interval '1 day' AS tanggal_transaksi,
  1 AS kuantitas,
  'Transport lokal' AS uraian,
  'Transportasi' AS kategori,
  75000::bigint AS harga_unit,
  75000::bigint AS total_harga,
  '/public/struk/struk_' || r.id || '_3.jpg' AS url_struk
FROM realisasi_bon_sementaras r;

INSERT INTO rbs_rincians (
  rbs_id,
  tanggal_transaksi,
  kuantitas,
  uraian,
  kategori,
  harga_unit,
  total_harga,
  url_struk
)
SELECT
  r.id AS rbs_id,
  r.periode_kembali AS tanggal_transaksi,
  1 AS kuantitas,
  'Parkir' AS uraian,
  'Parkir' AS kategori,
  15000::bigint AS harga_unit,
  15000::bigint AS total_harga,
  '/public/struk/struk_' || r.id || '_4.jpg' AS url_struk
FROM realisasi_bon_sementaras r;

-- =========================================================
-- UPDATE TOTAL REALISASI & SELISIH
-- =========================================================
UPDATE realisasi_bon_sementaras r
SET
  total_realisasi = x.total_realisasi,
  selisih = p.total_estimasi - x.total_realisasi
FROM (
  SELECT rbs_id, SUM(total_harga) AS total_realisasi
  FROM rbs_rincians
  GROUP BY rbs_id
) x
JOIN request_ppds p ON p.id = x.rbs_id
WHERE r.id = x.rbs_id;

-- =========================================================
-- DOKUMEN PPD
-- =========================================================
INSERT INTO dokumens (
  doc_ref_id,
  doc_ref_type,
  user_id,
  nomor_dokumen,
  tipe_dokumen,
  nomor_tipe_dokumen,
  created_at
)
SELECT
  p.id AS doc_ref_id,
  'RequestPPD' AS doc_ref_type,
  p.user_id AS user_id,
  'DOC/PPD/2026/' || LPAD(p.id::text, 5, '0') AS nomor_dokumen,
  'PPD' AS tipe_dokumen,
  'PPD/GA/05/2026/' || LPAD(p.id::text, 4, '0') AS nomor_tipe_dokumen,
  p.created_at
FROM request_ppds p;

-- =========================================================
-- DOKUMEN RBS
-- =========================================================
INSERT INTO dokumens (
  doc_ref_id,
  doc_ref_type,
  user_id,
  nomor_dokumen,
  tipe_dokumen,
  nomor_tipe_dokumen,
  created_at
)
SELECT
  r.id AS doc_ref_id,
  'RealisasiBonSementara' AS doc_ref_type,
  r.user_id AS user_id,
  'DOC/RBS/2026/' || LPAD(r.id::text, 5, '0') AS nomor_dokumen,
  'RBS' AS tipe_dokumen,
  'RBS/GA/05/2026/' || LPAD(r.id::text, 4, '0') AS nomor_tipe_dokumen,
  r.created_at
FROM realisasi_bon_sementaras r;

-- =========================================================
-- RIWAYAT APPROVAL PPD
-- =========================================================
INSERT INTO riwayat_approvals (
  doc_ref_id,
  doc_ref_type,
  user_id,
  nama,
  jabatan,
  tindakan,
  catatan,
  created_at
)
SELECT
  p.id AS doc_ref_id,
  'RequestPPD' AS doc_ref_type,
  ((p.id % 5) + 1) AS user_id,
  'Atasan ' || ((p.id % 5) + 1) AS nama,
  'Atasan' AS jabatan,
  'Disetujui' AS tindakan,
  'Seed approval atasan' AS catatan,
  p.created_at + interval '1 hour' AS created_at
FROM request_ppds p
WHERE p.status IN ('Menunggu HRGA', 'Menunggu Direktur', 'Menunggu Finance', 'Selesai');

INSERT INTO riwayat_approvals (
  doc_ref_id,
  doc_ref_type,
  user_id,
  nama,
  jabatan,
  tindakan,
  catatan,
  created_at
)
SELECT
  p.id AS doc_ref_id,
  'RequestPPD' AS doc_ref_type,
  91 AS user_id,
  'HRGA 91' AS nama,
  'HRGA' AS jabatan,
  'Disetujui' AS tindakan,
  'Seed approval HRGA' AS catatan,
  p.created_at + interval '2 hours' AS created_at
FROM request_ppds p
WHERE p.status IN ('Menunggu Direktur', 'Menunggu Finance', 'Selesai');

INSERT INTO riwayat_approvals (
  doc_ref_id,
  doc_ref_type,
  user_id,
  nama,
  jabatan,
  tindakan,
  catatan,
  created_at
)
SELECT
  p.id AS doc_ref_id,
  'RequestPPD' AS doc_ref_type,
  98 AS user_id,
  'Direktur 98' AS nama,
  'Direktur' AS jabatan,
  'Disetujui' AS tindakan,
  'Seed approval direktur' AS catatan,
  p.created_at + interval '3 hours' AS created_at
FROM request_ppds p
WHERE p.status IN ('Menunggu Finance', 'Selesai');

INSERT INTO riwayat_approvals (
  doc_ref_id,
  doc_ref_type,
  user_id,
  nama,
  jabatan,
  tindakan,
  catatan,
  created_at
)
SELECT
  p.id AS doc_ref_id,
  'RequestPPD' AS doc_ref_type,
  95 AS user_id,
  'Finance 95' AS nama,
  'Finance' AS jabatan,
  'Disetujui' AS tindakan,
  'Seed approval finance' AS catatan,
  p.created_at + interval '4 hours' AS created_at
FROM request_ppds p
WHERE p.status = 'Selesai';

-- =========================================================
-- RIWAYAT APPROVAL RBS
-- =========================================================
INSERT INTO riwayat_approvals (
  doc_ref_id,
  doc_ref_type,
  user_id,
  nama,
  jabatan,
  tindakan,
  catatan,
  created_at
)
SELECT
  r.id AS doc_ref_id,
  'RealisasiBonSementara' AS doc_ref_type,
  ((r.id % 5) + 1) AS user_id,
  'Atasan ' || ((r.id % 5) + 1) AS nama,
  'Atasan' AS jabatan,
  'Disetujui' AS tindakan,
  'Seed approval RBS atasan' AS catatan,
  r.created_at + interval '1 hour' AS created_at
FROM realisasi_bon_sementaras r
WHERE r.status IN ('Menunggu HRGA', 'Menunggu Direktur', 'Menunggu Finance', 'Selesai');

INSERT INTO riwayat_approvals (
  doc_ref_id,
  doc_ref_type,
  user_id,
  nama,
  jabatan,
  tindakan,
  catatan,
  created_at
)
SELECT
  r.id AS doc_ref_id,
  'RealisasiBonSementara' AS doc_ref_type,
  91 AS user_id,
  'HRGA 91' AS nama,
  'HRGA' AS jabatan,
  'Disetujui' AS tindakan,
  'Seed approval RBS HRGA' AS catatan,
  r.created_at + interval '2 hours' AS created_at
FROM realisasi_bon_sementaras r
WHERE r.status IN ('Menunggu Direktur', 'Menunggu Finance', 'Selesai');

INSERT INTO riwayat_approvals (
  doc_ref_id,
  doc_ref_type,
  user_id,
  nama,
  jabatan,
  tindakan,
  catatan,
  created_at
)
SELECT
  r.id AS doc_ref_id,
  'RealisasiBonSementara' AS doc_ref_type,
  98 AS user_id,
  'Direktur 98' AS nama,
  'Direktur' AS jabatan,
  'Disetujui' AS tindakan,
  'Seed approval RBS direktur' AS catatan,
  r.created_at + interval '3 hours' AS created_at
FROM realisasi_bon_sementaras r
WHERE r.status IN ('Menunggu Finance', 'Selesai');

INSERT INTO riwayat_approvals (
  doc_ref_id,
  doc_ref_type,
  user_id,
  nama,
  jabatan,
  tindakan,
  catatan,
  created_at
)
SELECT
  r.id AS doc_ref_id,
  'RealisasiBonSementara' AS doc_ref_type,
  95 AS user_id,
  'Finance 95' AS nama,
  'Finance' AS jabatan,
  'Disetujui' AS tindakan,
  'Seed approval RBS finance' AS catatan,
  r.created_at + interval '4 hours' AS created_at
FROM realisasi_bon_sementaras r
WHERE r.status = 'Selesai';

-- =========================================================
-- RESET SEQUENCES
-- =========================================================
SELECT setval(pg_get_serial_sequence('users', 'id'), COALESCE((SELECT MAX(id) FROM users), 1), true);
SELECT setval(pg_get_serial_sequence('request_ppds', 'id'), COALESCE((SELECT MAX(id) FROM request_ppds), 1), true);
SELECT setval(pg_get_serial_sequence('ppd_transportasis', 'id'), COALESCE((SELECT MAX(id) FROM ppd_transportasis), 1), true);
SELECT setval(pg_get_serial_sequence('ppd_hotels', 'id'), COALESCE((SELECT MAX(id) FROM ppd_hotels), 1), true);
SELECT setval(pg_get_serial_sequence('ppd_rincian_tambahans', 'id'), COALESCE((SELECT MAX(id) FROM ppd_rincian_tambahans), 1), true);
SELECT setval(pg_get_serial_sequence('realisasi_bon_sementaras', 'id'), COALESCE((SELECT MAX(id) FROM realisasi_bon_sementaras), 1), true);
SELECT setval(pg_get_serial_sequence('rbs_rincians', 'id'), COALESCE((SELECT MAX(id) FROM rbs_rincians), 1), true);
SELECT setval(pg_get_serial_sequence('dokumens', 'id'), COALESCE((SELECT MAX(id) FROM dokumens), 1), true);
SELECT setval(pg_get_serial_sequence('riwayat_approvals', 'id'), COALESCE((SELECT MAX(id) FROM riwayat_approvals), 1), true);

COMMIT;

-- =========================================================
-- CHECK RESULT
-- =========================================================
SELECT 'users' AS table_name, COUNT(*) AS total FROM users
UNION ALL
SELECT 'request_ppds', COUNT(*) FROM request_ppds
UNION ALL
SELECT 'ppd_transportasis', COUNT(*) FROM ppd_transportasis
UNION ALL
SELECT 'ppd_hotels', COUNT(*) FROM ppd_hotels
UNION ALL
SELECT 'ppd_rincian_tambahans', COUNT(*) FROM ppd_rincian_tambahans
UNION ALL
SELECT 'realisasi_bon_sementaras', COUNT(*) FROM realisasi_bon_sementaras
UNION ALL
SELECT 'rbs_rincians', COUNT(*) FROM rbs_rincians
UNION ALL
SELECT 'dokumens', COUNT(*) FROM dokumens
UNION ALL
SELECT 'riwayat_approvals', COUNT(*) FROM riwayat_approvals;