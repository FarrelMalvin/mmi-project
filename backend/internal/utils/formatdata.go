package utils

import (
	"fmt"
	"time"
	"strings"
	"encoding/base64"
	"os"
	"path/filepath"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func FormatTanggal(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	bulan := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	return fmt.Sprintf("%02d %s %d", t.Day(), bulan[t.Month()], t.Year())
}

func FormatRupiah(nominal int64) string {
	p := message.NewPrinter(language.Indonesian)
	return p.Sprintf("Rp %d", nominal)
}

func FormatNominal(nominal int64) string {
	p := message.NewPrinter(language.Indonesian)
	return p.Sprintf("%d", nominal)
}


func getBase64Image(filePath string) string {
	if filePath == "" {
		return ""
	}

	cleanPath := strings.TrimLeft(filePath, "/") 
	imgBytes, err := os.ReadFile(cleanPath)
	if err != nil {
		fmt.Printf("Gagal membaca file gambar %s: %v\n", cleanPath, err)
		return ""
	}
	return base64.StdEncoding.EncodeToString(imgBytes)
}

func GetBase64Image(filePath string) string {
    if filePath == "" {
        return ""
    }
	
    cleanPath := strings.TrimPrefix(filePath, "/")

    pwd, _ := os.Getwd()

    absPath := filepath.Join(pwd, cleanPath)

    imgData, err := os.ReadFile(absPath)
    if err != nil {
        fmt.Printf("DEBUG ERROR: Gagal membaca file di: %s | Error: %v\n", absPath, err)
        return ""
    }

    return base64.StdEncoding.EncodeToString(imgData)
}