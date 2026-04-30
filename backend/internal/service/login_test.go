package service

import (
	"context"
	"errors"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"golang-mmi/internal/model"
	"golang-mmi/mocks"
)

func TestValidateCredentials(t *testing.T) {
	// Persiapan password yang sudah di-hash untuk mock
	plainPassword := "secret123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		nama          string
		password      string
		mockSetup     func(repo *mocks.UserRepository)
		expectedError string
	}{
		{
			name:     "Sukses - Kredensial Valid",
			nama:     "JohnDoe",
			password: plainPassword,
			mockSetup: func(repo *mocks.UserRepository) {
				repo.On("GetByName", mock.Anything, "JohnDoe").
					Return(&model.User{Nama: "JohnDoe", Password: string(hashedPassword)}, nil).Once()
			},
			expectedError: "",
		},
		{
			name:     "Gagal - User Tidak Ditemukan",
			nama:     "Unknown",
			password: plainPassword,
			mockSetup: func(repo *mocks.UserRepository) {
				repo.On("GetByName", mock.Anything, "Unknown").
					Return(nil, errors.New("not found")).Once()
			},
			expectedError: "invalid credentials",
		},
		{
			name:     "Gagal - Password Salah",
			nama:     "JohnDoe",
			password: "wrongpassword",
			mockSetup: func(repo *mocks.UserRepository) {
				repo.On("GetByName", mock.Anything, "JohnDoe").
					Return(&model.User{Nama: "JohnDoe", Password: string(hashedPassword)}, nil).Once()
			},
			expectedError: "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewUserRepository(t)
			tt.mockSetup(mockRepo)

			svc := NewUserService(mockRepo)
			user, err := svc.ValidateCredentials(context.Background(), tt.password, tt.nama)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.nama, user.Nama)
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint
		mockSetup     func(repo *mocks.UserRepository)
		expectedError string
	}{
		{
			name:   "Sukses - User Ditemukan",
			userID: 1,
			mockSetup: func(repo *mocks.UserRepository) {
				repo.On("GetByID", mock.Anything, uint(1)).
					Return(&model.User{Id: 1, Nama: "John"}, nil).Once()
			},
			expectedError: "",
		},
		{
			name:   "Gagal - User Tidak Ada",
			userID: 99,
			mockSetup: func(repo *mocks.UserRepository) {
				repo.On("GetByID", mock.Anything, uint(99)).
					Return(nil, errors.New("sql: no rows")).Once()
			},
			expectedError: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewUserRepository(t)
			tt.mockSetup(mockRepo)

			svc := NewUserService(mockRepo)
			user, err := svc.GetUserByID(context.Background(), tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, user.Id)
			}
		})
	}
}