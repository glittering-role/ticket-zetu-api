package service

import (
	"context"
	"errors"
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ImageService interface {
	UploadProfileImage(ctx context.Context, userID string, file interface{}, contentType string) (string, error)
	DeleteProfileImage(userID string) error
}

type imageService struct {
	db         *gorm.DB
	cloudinary *cloudinary.CloudinaryService
}

func NewImageService(db *gorm.DB, cloudinary *cloudinary.CloudinaryService) ImageService {
	return &imageService{
		db:         db,
		cloudinary: cloudinary,
	}
}

func (s *imageService) UploadProfileImage(ctx context.Context, userID string, file interface{}, contentType string) (string, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return "", errors.New("invalid user ID format")
	}

	var user members.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("user not found")
		}
		return "", err
	}

	if !s.isValidFileType(contentType) {
		return "", errors.New("invalid file type. Only images are allowed")
	}

	url, err := s.cloudinary.UploadFile(ctx, file, "profile_images")
	if err != nil {
		return "", errors.New("failed to upload file to Cloudinary")
	}

	user.AvatarURL = url
	user.UpdatedAt = time.Now()

	if err := s.db.Save(&user).Error; err != nil {
		return "", err
	}

	return url, nil
}

func (s *imageService) DeleteProfileImage(userID string) error {
	if _, err := uuid.Parse(userID); err != nil {
		return errors.New("invalid user ID format")
	}

	var user members.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	if user.AvatarURL == "" {
		return errors.New("no profile image to delete")
	}

	if err := s.cloudinary.DeleteFile(context.Background(), user.AvatarURL); err != nil {
		return errors.New("failed to delete file from Cloudinary")
	}

	user.AvatarURL = ""
	user.UpdatedAt = time.Now()

	if err := s.db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

func (s *imageService) isValidFileType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}
