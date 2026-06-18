package service

import (
    "context"
    "fmt"
    "io"
    "mime/multipart"
    "os"
    "path/filepath"
    "strings"
    "time"
)

type UploadService interface {
    UploadStruk(ctx context.Context, file *multipart.FileHeader, userID uint) (string, error)
    DeleteFileByURL(ctx context.Context, fileURL string) error
}

type UploadImpl struct{}

func (s *UploadImpl) UploadStruk(ctx context.Context, file *multipart.FileHeader, userID uint) (string, error) {
    if err := validateStrukFile(file); err != nil {
        return "", err
    }

    src, err := file.Open()
    if err != nil {
        return "", fmt.Errorf("gagal membuka file: %w", err)
    }
    defer src.Close()

    ext := strings.ToLower(filepath.Ext(file.Filename))
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
        _ = os.Remove(savePath)
        return "", fmt.Errorf("gagal menulis file: %w", err)
    }

    return fmt.Sprintf("/public/struk/%s", filename), nil
}

func (s *UploadImpl) DeleteFileByURL(ctx context.Context, fileURL string) error {
    if strings.TrimSpace(fileURL) == "" {
        return nil
    }

    if !strings.HasPrefix(fileURL, "/public/") {
        return nil
    }

    path := strings.TrimPrefix(fileURL, "/")
    path = filepath.Clean(path)

    if !strings.HasPrefix(path, filepath.Clean("public")) {
        return fmt.Errorf("path file tidak valid")
    }

    if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("gagal menghapus file: %w", err)
    }

    return nil
}

func validateStrukFile(file *multipart.FileHeader) error {
    const maxFileSize = 5 << 20

    if file == nil {
        return fmt.Errorf("file tidak boleh kosong")
    }

    if file.Size <= 0 {
        return fmt.Errorf("file kosong")
    }

    if file.Size > maxFileSize {
        return fmt.Errorf("ukuran file maksimal 5MB")
    }

    ext := strings.ToLower(filepath.Ext(file.Filename))

    allowed := map[string]bool{
        ".jpg":  true,
        ".jpeg": true,
        ".png":  true,
        ".webp": true,
        ".pdf":  true,
    }

    if !allowed[ext] {
        return fmt.Errorf("format file tidak didukung")
    }

    return nil
}