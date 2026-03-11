package service

import (
	"context"
	"errors"

	"golang-mmi/internal/model"
	"golang-mmi/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	ValidateCredentials(ctx context.Context, password string, nama string) (*model.User, error)
	GetUserByID(ctx context.Context, userID uint) (*model.User, error)
}

type UserServiceImpl struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &UserServiceImpl{
		repo: repo,
	}
}

func (s *UserServiceImpl) ValidateCredentials(ctx context.Context, password string, nama string) (*model.User, error) {
	user, err := s.repo.GetByName(ctx, nama)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *UserServiceImpl) GetUserByID(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}
