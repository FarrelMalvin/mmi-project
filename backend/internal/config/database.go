package config

import (
	"fmt"
	"golang-mmi/internal/model"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan variabel environment bawaan OS")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, password, dbname, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}

	log.Println("Koneksi ke database PostgreSQL berhasil!")

	err = db.AutoMigrate(
		&model.User{},
		&model.RequestPPD{},
		&model.RealisasiBonSementara{},
		&model.BlacklistedToken{},
		&model.Dokumen{},
		&model.PPDHotel{},
		&model.PPDRincianTambahan{},
		&model.PPDTransportasi{},
		&model.RBSrincian{},
		&model.RefreshToken{},
		&model.RiwayatApproval{},
	)

	if err != nil {
		log.Fatalf("error migrate: %v", err)
	}

	return db
}
