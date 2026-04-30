package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"golang-mmi/internal/dto"
	"golang-mmi/internal/model"
	"golang-mmi/internal/repository"

	"github.com/disintegration/imaging"
)

type UserService interface {
	GetUserByID(ctx context.Context, userID uint) (*model.User, error)
	UpdateSignaturePath(ctx context.Context, userID uint, file *multipart.FileHeader) (string, error)
	GetDataProfile(ctx context.Context, userID uint)(dto.ProfileResponse, error)
}

type UserServiceImpl struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &UserServiceImpl{
		repo: repo,
	}
}

func (s *UserServiceImpl) GetUserByID(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *UserServiceImpl) UpdateSignaturePath(ctx context.Context, userID uint, file *multipart.FileHeader) (string, error) {
	ext := filepath.Ext(file.Filename)
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		return "", errors.New("invalid file type")
	}

	folderPath := filepath.Join("storage", "signature")

	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
    return "", fmt.Errorf("gagal menyiapkan direktori penyimpanan: %v", err)
	}

	fileName := fmt.Sprintf("ttd_%d_%d%s", userID, time.Now().Unix(), ext)
	savePath := filepath.Join(folderPath, fileName)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	if err != nil {
		return "", fmt.Errorf("gagal membaca gambar: %v", err)
	}

	img, err := imaging.Decode(src)
	if err != nil {
		return "", fmt.Errorf("gagal decode gambar: %v", err)
	}

	resizedImg := imaging.Resize(img, 300, 0, imaging.Lanczos)

	err = imaging.Save(resizedImg, savePath)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan gambar: %v", err)
	}

	dbPath := "/" + filepath.ToSlash(savePath)
	err = s.repo.UpdateSignaturePath(ctx, userID, dbPath)
	if err != nil {
		return "", err
	}

	return dbPath, nil

}

func (s *UserServiceImpl) GetDataProfile(ctx context.Context, userID uint)(dto.ProfileResponse, error){
	data, err := s.repo.GetUserDataDetail(ctx, userID)
	if err != nil{
		return dto.ProfileResponse{}, fmt.Errorf("gagal mengambil detail: %w", err)
	}

	response := dto.ProfileResponse{
		Nama: data.Nama,
		Jabatan: data.Jabatan,
		Wilayah: data.Wilayah,
		Departemen: data.Departemen,
		PathTandaTangan: data.PathTandaTangan,
	}

	fmt.Printf("---DEBUG---\n", response)

	return response, nil
}
