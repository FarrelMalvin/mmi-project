package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"golang-mmi/internal/model"
)

type userRepository struct {
	db *gorm.DB
}

type UserRepository interface {
	GetByName(ctx context.Context, nama string) (*model.User, error)
	GetByID(ctx context.Context, userID uint) (*model.User, error)
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) GetByName(ctx context.Context, nama string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("nama = ?", nama).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
	}
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
	}
	return &user, nil
}
