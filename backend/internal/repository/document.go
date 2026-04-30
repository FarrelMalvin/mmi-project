// repository/document.go
package repository

import (
    "context"
    "errors"
    "fmt"

    "golang-mmi/internal/model"
    "gorm.io/gorm"
)

type DocumentRepository interface {
    GetLastNomorDokumenGeneral(ctx context.Context, pattern string) (string, error)
    GetLastNomorDokumenSpecific(ctx context.Context, tipe string, pattern string) (string, error)
    CreateDokumen(ctx context.Context, dokumen *model.Dokumen) error
    RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Document struct {
    db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
    return &Document{db: db}
}

type contextKey string
const txKey contextKey = "tx"

func (r *Document) getDB(ctx context.Context) *gorm.DB {
    if tx, ok := ctx.Value(txKey).(*gorm.DB); ok && tx != nil {
        return tx.WithContext(ctx)
    }
    return r.db.WithContext(ctx)
}

func (r *Document) GetLastNomorDokumenGeneral(ctx context.Context, pattern string) (string, error) {
    var lastDoc model.Dokumen

    err := r.getDB(ctx).
        Set("gorm:query_option", "FOR UPDATE").
        Where("nomor_dokumen LIKE ?", pattern).
        Order("id DESC").
        First(&lastDoc).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return "", nil
        }
        return "", fmt.Errorf("gagal query nomor dokumen terakhir: %w", err)
    }

    return lastDoc.NomorDokumen, nil
}

func (r *Document) GetLastNomorDokumenSpecific(ctx context.Context, tipe string, pattern string) (string, error) {
    var lastDoc model.Dokumen

    err := r.getDB(ctx).
        Set("gorm:query_option", "FOR UPDATE").
        Where("tipe_dokumen = ? AND nomor_tipe_dokumen LIKE ?", tipe, pattern).
        Order("id DESC").
        First(&lastDoc).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return "", nil
        }
        return "", fmt.Errorf("gagal query nomor tipe dokumen terakhir: %w", err)
    }

    return lastDoc.NomorTipeDokumen, nil
}

func (r *Document) CreateDokumen(ctx context.Context, dokumen *model.Dokumen) error {
    if err := r.getDB(ctx).Create(dokumen).Error; err != nil {
        return fmt.Errorf("gagal menyimpan dokumen: %w", err)
    }
    return nil
}

func (r *Document) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        txCtx := context.WithValue(ctx, txKey, tx)
        return fn(txCtx)
    })
}