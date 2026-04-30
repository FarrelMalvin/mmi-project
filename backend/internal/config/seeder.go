package config

import (
	"golang-mmi/internal/model"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedUsers(db *gorm.DB) {
	// Helper untuk hash password
	hashPassword := func(p string) string {
		h, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
		return string(h)
	}

	// 1. BUAT USER INDEPENDEN (TANPA ATASAN)

	// Direktur
	direktur := model.User{
		Nama:       "Direktur",
		Nik:        "DIR001",
		Jabatan:    "Direktur",
		Wilayah:    "Pusat",
		Departemen: "Management",
		Password:   hashPassword("12345"),
	}
	db.Where(model.User{Nik: "DIR001"}).FirstOrCreate(&direktur)

	// HRGA
	hrga := model.User{
		Nama:       "HRGA",
		Nik:        "HRG001",
		Jabatan:    "HRGA",
		Wilayah:    "Pusat",
		Departemen: "HRGA",
		Password:   hashPassword("12345"),
	}
	db.Where(model.User{Nik: "HRG001"}).FirstOrCreate(&hrga)

	// Finance
	finance := model.User{
		Nama:       "Finance",
		Nik:        "FIN001",
		Jabatan:    "Finance",
		Wilayah:    "Pusat",
		Departemen: "Finance",
		Password:   hashPassword("12345"),
	}
	db.Where(model.User{Nik: "FIN001"}).FirstOrCreate(&finance)

	// Atasan A
	atasanA := model.User{
		Nama:       "AtasanA",
		Nik:        "ATS001",
		Jabatan:    "Atasan",
		Wilayah:    "Pusat",
		Departemen: "Operational",
		Password:   hashPassword("12345"),
	}
	db.Where(model.User{Nik: "ATS001"}).FirstOrCreate(&atasanA)

	// Atasan B
	atasanB := model.User{
		Nama:       "AtasanB",
		Nik:        "ATS002",
		Jabatan:    "Atasan",
		Wilayah:    "Pusat",
		Departemen: "Operational",
		Password:   hashPassword("12345"),
	}
	db.Where(model.User{Nik: "ATS002"}).FirstOrCreate(&atasanB)

	// 2. BUAT USER PEGAWAI (DENGAN RELASI ATASAN)

	// Pegawai A -> Atasan A
	pegawaiA := model.User{
		Nama:       "PegawaiA",
		Nik:        "PGW001",
		Jabatan:    "Pegawai",
		Wilayah:    "Pusat",
		Departemen: "Operational",
		Password:   hashPassword("12345"),
		AtasanID:   &atasanA.Id, // Link ke Atasan A
	}
	db.Where(model.User{Nik: "PGW001"}).FirstOrCreate(&pegawaiA)

	// Pegawai B -> Atasan B
	pegawaiB := model.User{
		Nama:       "PegawaiB",
		Nik:        "PGW002",
		Jabatan:    "Pegawai",
		Wilayah:    "Pusat",
		Departemen: "Operational",
		Password:   hashPassword("12345"),
		AtasanID:   &atasanB.Id, // Link ke Atasan B
	}
	db.Where(model.User{Nik: "PGW002"}).FirstOrCreate(&pegawaiB)

	log.Println("✅ Seeding 7 users selesai!")
}
