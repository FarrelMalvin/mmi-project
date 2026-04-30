package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	// Sesuaikan path import ini dengan nama module kamu
	"golang-mmi/internal/config"
	"golang-mmi/internal/model"

	"gorm.io/gorm"
)

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) config.TokenStore {
	return &tokenRepository{db: db}
}

func parseUserID(userID string) (uint, error) {
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return 0, errors.New("format user ID tidak valid")
	}
	return uint(id), nil
}

func (r *tokenRepository) StoreRefreshToken(ctx context.Context, tokenID, userID string, expiresAt time.Time) error {
	uid, err := parseUserID(userID)
	if err != nil {
		return err
	}

	token := model.RefreshToken{
		TokenID:   tokenID,
		UserID:    uid,
		ExpiresAt: expiresAt,
		Revoked:   false,
	}
	return r.db.WithContext(ctx).Create(&token).Error
}

func (r *tokenRepository) GetRefreshToken(ctx context.Context, tokenID string) (*config.RefreshTokenData, error) {
	var token model.RefreshToken
	err := r.db.WithContext(ctx).Where("token_id = ?", tokenID).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &config.RefreshTokenData{
		TokenID:   token.TokenID,
		UserID:    strconv.Itoa(int(token.UserID)),
		ExpiresAt: token.ExpiresAt,
		Revoked:   token.Revoked,
		CreatedAt: token.CreatedAt,
	}, nil
}

func (r *tokenRepository) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	return r.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("token_id = ?", tokenID).Update("revoked", true).Error
}

func (r *tokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	uid, err := parseUserID(userID)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("user_id = ?", uid).Update("revoked", true).Error
}

func (r *tokenRepository) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.BlacklistedToken{}).Where("token_id = ?", tokenID).Count(&count).Error
	return count > 0, err
}

func (r *tokenRepository) BlacklistToken(ctx context.Context, tokenID string, expiresAt time.Time) error {
	blacklist := model.BlacklistedToken{
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
	}
	return r.db.WithContext(ctx).Create(&blacklist).Error
}

func (r *tokenRepository) GetUserTokenVersion(ctx context.Context, userID string) (int, error) {
	uid, err := parseUserID(userID)
	if err != nil {
		return 0, err
	}

	var version int
	err = r.db.WithContext(ctx).Model(&model.User{}).Select("token_version").Where("id = ?", uid).Scan(&version).Error
	return version, err
}

func (r *tokenRepository) IncrementUserTokenVersion(ctx context.Context, userID string) (int, error) {
	uid, err := parseUserID(userID)
	if err != nil {
		return 0, err
	}

	var newVersion int
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).Where("id = ?", uid).UpdateColumn("token_version", gorm.Expr("token_version + ?", 1)).Error; err != nil {
			return err
		}
		return tx.Model(&model.User{}).Select("token_version").Where("id = ?", uid).Scan(&newVersion).Error
	})

	return newVersion, err
}
