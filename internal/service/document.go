// service/document.go
package service

import (
    "context"
    "fmt"
    "strconv"
    "strings"
    "time"

    "golang-mmi/internal/constant"
    "golang-mmi/internal/model"
    "golang-mmi/internal/repository"
)

type DocumentService interface {
    GenerateNomorDokumenGeneral(ctx context.Context, dept string) (string, error)
    GenerateNomorDokumenSpecific(ctx context.Context, tipe string, prefix string) (string, error)
    SaveDokumen(ctx context.Context, dokumen *model.Dokumen) error
}

type DocumentImpl struct {
    repo repository.DocumentRepository
}

func NewDocumentService(repo repository.DocumentRepository) DocumentService {
    return &DocumentImpl{repo: repo}
}

func (s *DocumentImpl) GenerateNomorDokumenGeneral(ctx context.Context, dept string) (string, error) {
    var result string

    err := s.repo.RunInTransaction(ctx, func(txCtx context.Context) error {
        pattern := fmt.Sprintf("MMI/%s/%%", dept)

        lastno, err := s.repo.GetLastNomorDokumenGeneral(txCtx, pattern)
        if err != nil {
            return err
        }

        newSeq := 1
        if lastno != "" {
            parts := strings.Split(lastno, "/")
            if len(parts) >= 3 {
                // ✅ Handle error Atoi
                lastSeq, err := strconv.Atoi(parts[len(parts)-1])
                if err != nil {
                    return fmt.Errorf("format nomor tidak valid '%s': %w", parts[len(parts)-1], err)
                }
                newSeq = lastSeq + 1
            }
        }

        result = fmt.Sprintf("MMI/%s/%03d", dept, newSeq)
        return nil
    })

    if err != nil {
        return "", fmt.Errorf("gagal generate nomor dokumen general: %w", err)
    }

    return result, nil
}

func (s *DocumentImpl) GenerateNomorDokumenSpecific(ctx context.Context, tipe string, prefix string) (string, error) {
    var result string

    err := s.repo.RunInTransaction(ctx, func(txCtx context.Context) error {
        now := time.Now()
        bulan := fmt.Sprintf("%02d", int(now.Month()))
        tahun := fmt.Sprintf("%d", now.Year())

        pattern := fmt.Sprintf("%s/%s/%s/%s/%%", prefix, constant.KodeDeptGA, bulan, tahun)

        lastno, err := s.repo.GetLastNomorDokumenSpecific(txCtx, tipe, pattern)
        if err != nil {
            return err
        }

        newSeq := 1
        if lastno != "" {
            parts := strings.Split(lastno, "/")
            if len(parts) >= 5 {
                lastSeq, err := strconv.Atoi(parts[len(parts)-1])
                if err != nil {
                    return fmt.Errorf("format nomor tidak valid '%s': %w", parts[len(parts)-1], err)
                }
                newSeq = lastSeq + 1
            }
        }

        result = fmt.Sprintf("%s/%s/%s/%s/%03d", prefix, constant.KodeDeptGA, bulan, tahun, newSeq)
        return nil
    })  

    if err != nil {
        return "", fmt.Errorf("gagal generate nomor dokumen spesifik: %w", err)
    }

    return result, nil
}

func (s *DocumentImpl) SaveDokumen(ctx context.Context, dokumen *model.Dokumen) error {
    return s.repo.CreateDokumen(ctx, dokumen)
}