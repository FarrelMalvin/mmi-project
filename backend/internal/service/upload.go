package service

import(
	"context"
	"os"
	"fmt"
	"time"
	"io"
	"mime/multipart"
	"path/filepath"

)

type UploadService interface {
    UploadStruk(ctx context.Context, file *multipart.FileHeader, userID uint) (string, error)
}

type UploadImpl struct{}

func (s *UploadImpl) UploadStruk(ctx context.Context, file *multipart.FileHeader, userID uint) (string, error) {
    src, err := file.Open()
    if err != nil {
        return "", fmt.Errorf("gagal membuka file: %w", err)
    }
    defer src.Close()

    ext := filepath.Ext(file.Filename)
    filename := fmt.Sprintf("struk_%d_%d%s", userID, time.Now().UnixNano(), ext)
    savePath := filepath.Join("public", "struk", filename)

    if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
        return "", fmt.Errorf("gagal membuat direktori: %w", err)
    }

    dst, err := os.Create(savePath)
    if err != nil {
        return "", fmt.Errorf("gagal menyimpan file: %w", err)
    }
    defer dst.Close()

    if _, err := io.Copy(dst, src); err != nil {
        return "", fmt.Errorf("gagal menulis file: %w", err)
    }

    url := fmt.Sprintf("/public/struk/%s", filename)
    return url, nil
}