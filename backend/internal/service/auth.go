package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"golang-mmi/internal/model"
	"golang-mmi/internal/repository"
)

type AuthService interface {
	ValidateCredentials(ctx context.Context, password string, nama string) (*model.User, error)
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error 
}

type AuthServiceImpl struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) AuthService {
	return &AuthServiceImpl{
		repo: repo,
	}
}

var (
	ErrInvalidCredetial   = errors.New("nama pengguna atau password salah")
	ErrUserNotFound       = errors.New("user tidak ditemukan")
	ErrInvalidOldPassword = errors.New("password lama salah")
	ErrHashFailed         = errors.New("gagal mengenkripsi password")
	ErrUpdateFailed       = errors.New("gagal menyimpan password ke database")
)

func (s *AuthServiceImpl) ValidateCredentials(ctx context.Context, password string, nama string) (*model.User, error) {
	user, err := s.repo.GetByName(ctx, nama)
	if err != nil {
		return nil,ErrInvalidCredetial
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil,ErrInvalidCredetial
	}

	return user, nil
}

func (s *AuthServiceImpl) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return ErrInvalidOldPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return ErrHashFailed
	}

	err = s.repo.ChangePassword(ctx, userID, string(hashedPassword))
	if err != nil {
		return ErrUpdateFailed
	}

	return nil
}
