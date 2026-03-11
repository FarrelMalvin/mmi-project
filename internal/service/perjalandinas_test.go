package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"golang-mmi/internal/model"
	"golang-mmi/internal/repository"
	"golang-mmi/mocks"
)

func createTestContext(userID uint, jabatan string) context.Context {
	ctx := context.WithValue(context.Background(), "user_id", userID)
	ctx = context.WithValue(ctx, "jabatan", jabatan)
	return ctx
}

func TestApprovePerjalananDinas(t *testing.T) {
	tests := []struct {
		name          string
		ctx           context.Context
		ppdID         uint
		catatan       string
		mockSetup     func(repo *mocks.PerjalananDinasRepository)
		expectedError error
	}{
		{
			name:          "Gagal - Context User ID Tidak Ada",
			ctx:           context.Background(),
			ppdID:         1,
			catatan:       "OK",
			mockSetup:     func(repo *mocks.PerjalananDinasRepository) {},
			expectedError: errors.New("invalid user ID in context"),
		},
		{
			name:    "Sukses - Jabatan HRGA (Tanpa Dokumen)",
			ctx:     createTestContext(10, "HRGA"),
			ppdID:   1,
			catatan: "ACC by HRGA",
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				repo.On("GetStatusPerjalananDinas", mock.Anything, uint(1)).
					Return("Menunggu HRGA", nil).Once()

				repo.On("ApprovePerjalananDinas", mock.Anything, mock.AnythingOfType("repository.ApprovePerjalananDinasparams")).
					Return(nil).Once()
			},
			expectedError: nil,
		},

		{
			name:    "Sukses - Jabatan Atasan (Generate 1 Dokumen)",
			ctx:     createTestContext(11, "Atasan"),
			ppdID:   2,
			catatan: "ACC by Atasan",
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				repo.On("GetStatusPerjalananDinas", mock.Anything, uint(2)).
					Return("Menunggu Atasan", nil).Once()

				repo.On("GetLastNomorDokumenGeneral", mock.Anything, mock.Anything).
					Return("MMI/GA/03/2026/001", nil).Once()

				repo.On("ApprovePerjalananDinas", mock.Anything, mock.AnythingOfType("repository.ApprovePerjalananDinasparams")).
					Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:    "Sukses - Jabatan Finance (Generate 2 Dokumen)",
			ctx:     createTestContext(12, "Finance"),
			ppdID:   3,
			catatan: "ACC by Finance",
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				repo.On("GetStatusPerjalananDinas", mock.Anything, uint(3)).
					Return("Menunggu Finance", nil).Once()

				repo.On("GetLastNomorDokumenGeneral", mock.Anything, mock.Anything).
					Return("DK/GA/03/2026/001", nil).Once()

				repo.On("GetLastNomorDokumenSpecific", mock.Anything, "Bon Sementara", mock.Anything).
					Return("BS/03/2026/001", nil).Once()

				repo.On("ApprovePerjalananDinas", mock.Anything, mock.AnythingOfType("repository.ApprovePerjalananDinasparams")).
					Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:    "Gagal - Atasan Gagal Generate Nomor Umum",
			ctx:     createTestContext(11, "Atasan"),
			ppdID:   1,
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				// 1. Lolos cek status
				repo.On("GetStatusPerjalananDinas", mock.Anything, uint(1)).
					Return("Menunggu Atasan", nil).Once()

				// 2. PAKSA ERROR: Kita buat repo seolah-olah error saat ambil nomor terakhir
				repo.On("GetLastNomorDokumenGeneral", mock.Anything, mock.Anything).
					Return("", errors.New("database connection lost")).Once()
			},
			expectedError: errors.New("gagal generate nomor dokumen"),
		},
		{
			name:    "Gagal - Finance Gagal Generate Nomor Spesifik",
			ctx:     createTestContext(12, "Finance"),
			ppdID:   2,
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				repo.On("GetStatusPerjalananDinas", mock.Anything, uint(2)).
					Return("Menunggu Finance", nil).Once()

				// Nomor Umum Berhasil...
				repo.On("GetLastNomorDokumenGeneral", mock.Anything, mock.Anything).
					Return("DK/GA/03/2026/001", nil).Once()

				// ...TAPI Nomor Spesifik KITA PAKSA ERROR
				repo.On("GetLastNomorDokumenSpecific", mock.Anything, "Bon Sementara", mock.Anything).
					Return("", errors.New("sequence table locked")).Once()
			},
			expectedError: errors.New("gagal generate nomor dokumen spesifik"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := mocks.NewPerjalananDinasRepository(t)

			tt.mockSetup(mockRepo)

			svc := NewPerjalananDinasService(mockRepo)

			// Jalankan fungsi
			err := svc.ApprovePerjalananDinas(tt.ctx, tt.ppdID, tt.catatan)

			// Validasi
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreatePengajuanPerjalanaDinas(t *testing.T) {
	tests := []struct {
		name           string
		ctx            context.Context
		req            *model.RequestPPD
		mockSetup      func(repo *mocks.PerjalananDinasRepository)
		expectedError  error
		expectedStatus string
		expectedTotal  int64
	}{
		{
			name:          "Gagal - Context Jabatan Tidak Ada",
			ctx:           context.WithValue(context.Background(), "user_id", uint(1)), // Hanya user_id
			req:           &model.RequestPPD{},
			mockSetup:     func(repo *mocks.PerjalananDinasRepository) {},
			expectedError: errors.New("invalid jabatan in context"),
		},
		{
			name: "Gagal - Kategori Tambahan Tidak Valid",
			ctx:  createTestContext(1, "Pegawai"),
			req: &model.RequestPPD{
				RincianTambahan: []model.PPDRincianTambahan{
					{Kategori: "Hiburan Malam", Harga: 1000, Kuantitas: 1}, // Kategori ilegal
				},
			},
			mockSetup:     func(repo *mocks.PerjalananDinasRepository) {},
			expectedError: fmt.Errorf("kategori rincian tidak valid: Hiburan Malam"),
		},
		{
			name: "Gagal - Jabatan Tidak Diizinkan Membuat Pengajuan",
			ctx:  createTestContext(1, "Finance"), // Finance tidak ada di switch case
			req: &model.RequestPPD{
				RincianTransportasi: &model.PPDTransportasi{},
				RincianHotel:        &model.PPDHotel{},
				RincianTambahan:     []model.PPDRincianTambahan{},
			},
			mockSetup:     func(repo *mocks.PerjalananDinasRepository) {},
			expectedError: errors.New("invalid jabatan untuk membuat pengajuan"),
		},
		{
			name: "Sukses - Pegawai (Status Menunggu Atasan & Hitung Total)",
			ctx:  createTestContext(10, "Pegawai"),
			req: &model.RequestPPD{
				RincianTransportasi: &model.PPDTransportasi{JenisTransportasi: "Pesawat", Harga: 1000000},
				RincianHotel:        &model.PPDHotel{NamaHotel: "Hotel A", Harga: 500000},
				RincianTambahan: []model.PPDRincianTambahan{
					{Kategori: "BBM", Harga: 50000, Kuantitas: 2},      // 100.000
					{Kategori: "Konsumsi", Harga: 25000, Kuantitas: 4}, // 100.000
				},
			},
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				// Kita gunakan mock.MatchedBy untuk memverifikasi data yang sudah dimodifikasi backend
				repo.On("CreatePengajuanPerjalanaDinas", mock.Anything, mock.MatchedBy(func(r *model.RequestPPD) bool {
					return r.Status == "Menunggu Atasan" && r.TotalEstimasi == 1700000 && r.UserID == 10
				})).Return(nil).Once()
			},
			expectedError:  nil,
			expectedStatus: "Menunggu Atasan",
			expectedTotal:  1700000,
		},
		{
			name: "Sukses - Atasan (Status Menunggu HRGA)",
			ctx:  createTestContext(11, "Atasan"),
			req:  &model.RequestPPD{RincianTambahan: []model.PPDRincianTambahan{}},
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				repo.On("CreatePengajuanPerjalanaDinas", mock.Anything, mock.MatchedBy(func(r *model.RequestPPD) bool {
					return r.Status == "Menunggu HRGA"
				})).Return(nil).Once()
			},
			expectedError:  nil,
			expectedStatus: "Menunggu HRGA",
		},
		{
			name: "Sukses - HRGA (Status Menunggu Direktur)",
			ctx:  createTestContext(12, "HRGA"),
			req:  &model.RequestPPD{RincianTambahan: []model.PPDRincianTambahan{}},
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				repo.On("CreatePengajuanPerjalanaDinas", mock.Anything, mock.MatchedBy(func(r *model.RequestPPD) bool {
					return r.Status == "Menunggu Direktur"
				})).Return(nil).Once()
			},
			expectedError:  nil,
			expectedStatus: "Menunggu Direktur",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewPerjalananDinasRepository(t)
			tt.mockSetup(mockRepo)

			svc := NewPerjalananDinasService(mockRepo)
			err := svc.CreatePengajuanPerjalanaDinas(tt.ctx, tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, tt.req.Status)
				if tt.expectedTotal > 0 {
					assert.Equal(t, tt.expectedTotal, tt.req.TotalEstimasi)
				}
			}
		})
	}

}

func TestGetListPerjalananDinas(t *testing.T) {
	// Data dummy untuk hasil repository
	dummyResponse := []repository.RiwayatPPDResponse{
		{ID: 1, NomorDokumen: "DOC-001"},
	}

	tests := []struct {
		name          string
		ctx           context.Context
		page, limit   int
		mockSetup     func(repo *mocks.PerjalananDinasRepository)
		expectedLen   int
		expectedError error
	}{
		{
			name: "Gagal - Jabatan Tidak Ada di Context",
			ctx:  context.Background(),
			page: 1, limit: 10,
			mockSetup:     func(repo *mocks.PerjalananDinasRepository) {},
			expectedError: errors.New("invalid jabatan in context"),
		},
		{
			name: "Sukses - Jabatan HRGA/Finance/Direktur (Semua Data)",
			ctx:  createTestContext(1, "HRGA"),
			page: 1, limit: 10,
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				// Harus memanggil GetListRiwayatPerjalananDinas
				repo.On("GetListRiwayatPerjalananDinas", mock.Anything, 1, 10).
					Return(dummyResponse, int64(1), nil).Once()
			},
			expectedLen: 1,
		},
		{
			name: "Gagal - Pegawai Tanpa User ID",
			ctx:  context.WithValue(context.Background(), "jabatan", "Pegawai"),
			page: 1, limit: 10,
			mockSetup:     func(repo *mocks.PerjalananDinasRepository) {},
			expectedError: errors.New("invalid user ID in context"),
		},
		{
			name: "Sukses - Pegawai (Data Milik Sendiri)",
			ctx:  createTestContext(10, "Pegawai"),
			page: 1, limit: 10,
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				// Harus memanggil GetListPerjalananDinasByUserID
				repo.On("GetListPerjalananDinasByUserID", mock.Anything, uint(10), 1, 10).
					Return(dummyResponse, int64(1), nil).Once()
			},
			expectedLen: 1,
		},
		{
			name: "Sukses - Atasan (Data Bawahan)",
			ctx:  createTestContext(20, "Atasan"),
			page: 1, limit: 10,
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				// Harus memanggil GetListRiwayatPerjalananDinasByAtasan
				repo.On("GetListRiwayatPerjalananDinasByAtasan", mock.Anything, uint(20), 1, 10).
					Return(dummyResponse, int64(1), nil).Once()
			},
			expectedLen: 1,
		},
		{
			name: "Gagal - Jabatan Tidak Dikenali",
			ctx:  createTestContext(1, "UnknownRole"),
			page: 1, limit: 10,
			mockSetup:     func(repo *mocks.PerjalananDinasRepository) {},
			expectedError: errors.New("akses ditolak"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewPerjalananDinasRepository(t)
			tt.mockSetup(mockRepo)

			svc := NewPerjalananDinasService(mockRepo)
			res, total, err := svc.GetListPerjalananDinas(tt.ctx, tt.page, tt.limit)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLen, len(res))
				assert.GreaterOrEqual(t, total, int64(0))
			}
		})
	}
}

func TestGenerateNomorDokumenSpecific(t *testing.T) {
	now := time.Now()
	bulan := fmt.Sprintf("%02d", int(now.Month()))
	tahun := fmt.Sprintf("%d", now.Year())

	tests := []struct {
		name     string
		tipe     string
		prefix   string
		lastNo   string
		expected string
	}{
		{
			name:     "Sukses - Nomor Pertama (Urutan 001)",
			tipe:     "Bon Sementara",
			prefix:   "BS",
			lastNo:   "", // Database kosong
			expected: fmt.Sprintf("BS/GA/%s/%s/001", bulan, tahun),
		},
		{
			name:     "Sukses - Increment Urutan (005 -> 006)",
			tipe:     "Bon Sementara",
			prefix:   "BS",
			lastNo:   fmt.Sprintf("BS/GA/%s/%s/005", bulan, tahun),
			expected: fmt.Sprintf("BS/GA/%s/%s/006", bulan, tahun),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewPerjalananDinasRepository(t)

			// Pattern yang diharapkan oleh Repo
			pattern := fmt.Sprintf("%s/GA/%s/%s/%%", tt.prefix, bulan, tahun)
			mockRepo.On("GetLastNomorDokumenSpecific", mock.Anything, tt.tipe, pattern).
				Return(tt.lastNo, nil).Once()

			svc := NewPerjalananDinasService(mockRepo)
			result, err := svc.GenerateNomorDokumenSpecific(context.Background(), tt.tipe, tt.prefix)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetListPendingPerjalananDinas(t *testing.T) {
	tests := []struct {
		name          string
		ctx           context.Context
		mockSetup     func(repo *mocks.PerjalananDinasRepository)
		expectedError string
	}{
		{
			name:          "Gagal - Jabatan Tidak Ada",
			ctx:           context.Background(),
			mockSetup:     func(repo *mocks.PerjalananDinasRepository) {},
			expectedError: "unauthorized: jabatan tidak ditemukan",
		},
		{
			name:          "Gagal - Role Pegawai Dilarang (Forbidden)",
			ctx:           createTestContext(1, "Pegawai"),
			mockSetup:     func(repo *mocks.PerjalananDinasRepository) {},
			expectedError: "forbidden: pegawai tidak memiliki akses",
		},
		{
			name: "Sukses - HRGA Bisa Melihat List",
			ctx:  createTestContext(10, "HRGA"),
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				repo.On("GetListPendingPerjalananDinas", mock.Anything, "HRGA", uint(10)).
					Return([]model.RequestPPD{{Id: 1}}, nil).Once()
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewPerjalananDinasRepository(t)
			tt.mockSetup(mockRepo)

			svc := NewPerjalananDinasService(mockRepo)
			res, err := svc.GetListPendingPerjalananDinas(tt.ctx)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestGetListPerjalananDetail(t *testing.T) {
	mockRepo := mocks.NewPerjalananDinasRepository(t)
	ppdID := uint(1)

	// Data dummy rincian
	dummyRincian := []model.PPDRincianTambahan{
		{Kategori: "Tol", Harga: 20000},
	}

	// Mock Repo mengembalikan struct RequestPPD yang berisi rincian tersebut
	mockRepo.On("GetDetailPerjalananDinas", mock.Anything, ppdID).
		Return(model.RequestPPD{RincianTambahan: dummyRincian}, nil).Once()

	svc := NewPerjalananDinasService(mockRepo)
	result, err := svc.GetListPerjalananDetail(context.Background(), ppdID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Tol", result[0].Kategori)
}

func TestGetDropdownPPD(t *testing.T) {
	mockRepo := mocks.NewPerjalananDinasRepository(t)
	userID := uint(99)

	mockRepo.On("GetListPPDForRealisasi", mock.Anything, userID).
		Return([]repository.DropdownPPDResponse{{ID: 1}}, nil).Once()

	svc := NewPerjalananDinasService(mockRepo)
	res, err := svc.GetDropdownPPD(context.Background())

	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestDeclinePerjalananDinas(t *testing.T) {
	ProsesStatus = map[string][]string{
		"Menunggu Atasan": {"Menunggu HRGA", "Ditolak Atasan"},
		"Menunggu HRGA":   {"Menunggu Direktur", "Ditolak HRGA"},
	}

	tests := []struct {
		name          string
		ctx           context.Context
		ppdid         uint
		catatan       string
		mockSetup     func(repo *mocks.PerjalananDinasRepository)
		expectedError string
	}{
		{
			name:          "Gagal - User ID Tidak Ada",
			ctx:           context.WithValue(context.Background(), "jabatan", "Atasan"),
			ppdid:         1,
			mockSetup:     func(repo *mocks.PerjalananDinasRepository) {},
			expectedError: "invalid user ID in context",
		},
		{
			name:  "Gagal - Data Tidak Ditemukan",
			ctx:   createTestContext(10, "Atasan"),
			ppdid: 99,
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				repo.On("GetStatusPerjalananDinas", mock.Anything, uint(99)).
					Return("", errors.New("sql: no rows")).Once()
			},
			expectedError: "perjalanan dinas tidak ditemukan",
		},
		{
			name:  "Gagal - Jabatan Tidak Sesuai Status (Mismatch)",
			ctx:   createTestContext(10, "HRGA"),
			ppdid: 1,
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				repo.On("GetStatusPerjalananDinas", mock.Anything, uint(1)).
					Return("Menunggu Atasan", nil).Once() // Status nunggu Atasan, tapi yang nolak HRGA
			},
			expectedError: "tidak dapat menolak perjalanan dinas dengan status saat ini",
		},
		{
			name:    "Sukses - Atasan Menolak",
			ctx:     createTestContext(10, "Atasan"),
			ppdid:   1,
			catatan: "Revisi anggaran",
			mockSetup: func(repo *mocks.PerjalananDinasRepository) {
				repo.On("GetStatusPerjalananDinas", mock.Anything, uint(1)).
					Return("Menunggu Atasan", nil).Once()

				// Memastikan params yang dikirim ke repo sudah benar
				repo.On("DeclinePerjalananDinas", mock.Anything, mock.MatchedBy(func(p repository.DeclinePerjalananDinasParams) bool {
					return p.NextStatus == "Ditolak Atasan" && p.Riwayat.Tindakan == "Ditolak"
				})).Return(nil).Once()
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewPerjalananDinasRepository(t)
			tt.mockSetup(mockRepo)

			svc := NewPerjalananDinasService(mockRepo)
			err := svc.DeclinePerjalananDinas(tt.ctx, tt.ppdid, tt.catatan)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateNomorDokumenGeneral(t *testing.T) {
	now := time.Now()
	bulan := fmt.Sprintf("%02d", int(now.Month()))
	tahun := fmt.Sprintf("%d", now.Year())

	tests := []struct {
		name     string
		prefix   string
		lastNo   string
		expected string
	}{
		{
			name:     "Sukses - Nomor Pertama",
			prefix:   "MMI",
			lastNo:   "",
			expected: fmt.Sprintf("MMI/GA/%s/%s/001", bulan, tahun),
		},
		{
			name:     "Sukses - Increment dari 010 ke 011",
			prefix:   "MMI",
			lastNo:   fmt.Sprintf("MMI/GA/%s/%s/010", bulan, tahun),
			expected: fmt.Sprintf("MMI/GA/%s/%s/011", bulan, tahun),
		},
		{
			name:   "Gagal - Repo Error",
			prefix: "MMI",
			lastNo: "",
			// Skenario ini akan kita handle di mockSetup
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewPerjalananDinasRepository(t)

			pattern := fmt.Sprintf("%s/GA/%s/%s/%%", tt.prefix, bulan, tahun)

			if tt.name == "Gagal - Repo Error" {
				mockRepo.On("GetLastNomorDokumenGeneral", mock.Anything, pattern).
					Return("", errors.New("db error")).Once()

				svc := NewPerjalananDinasService(mockRepo)
				_, err := svc.GenerateNomorDokumenGeneral(context.Background(), tt.prefix)
				assert.Error(t, err)
			} else {
				mockRepo.On("GetLastNomorDokumenGeneral", mock.Anything, pattern).
					Return(tt.lastNo, nil).Once()

				svc := NewPerjalananDinasService(mockRepo)
				result, err := svc.GenerateNomorDokumenGeneral(context.Background(), tt.prefix)

				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
