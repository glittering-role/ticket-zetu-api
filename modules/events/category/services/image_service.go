package services

import (
	"context"
	"errors"
	"mime/multipart"
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/modules/events/models/categories"
	"ticket-zetu-api/modules/users/authorization"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ImageService interface {
	AddImage(userID, entityType, entityID string, file *multipart.FileHeader) (string, error)
	DeleteImage(userID, entityType, entityID string) error
}

type imageService struct {
	*BaseService
}

func NewImageService(db *gorm.DB, authService authorization.PermissionService, cloudinary *cloudinary.CloudinaryService) ImageService {
	base := NewBaseService(db, authService)
	base.cloudinary = cloudinary
	return &imageService{
		BaseService: base,
	}
}

func (s *imageService) AddImage(userID, entityType, entityID string, file *multipart.FileHeader) (string, error) {
	// Validate entity type
	if entityType != "category" && entityType != "subcategory" {
		return "", errors.New("invalid entity type")
	}

	// Check permission based on entity type
	permission := "update:" + entityType + "s"
	_, err := s.HasPermission(userID, permission)
	if err != nil {
		return "", err
	}
	// if !hasPerm {
	// 	return "", errors.New("user lacks " + permission + " permission")
	// }

	// Validate entity ID
	if _, err := uuid.Parse(entityID); err != nil {
		return "", errors.New("invalid " + entityType + " ID format")
	}

	// Check if entity exists
	var exists bool
	switch entityType {
	case "category":
		exists = s.db.Model(&categories.Category{}).Where("id = ? AND deleted_at IS NULL", entityID).First(&struct{}{}).Error == nil
	case "subcategory":
		exists = s.db.Model(&categories.Subcategory{}).Where("id = ? AND deleted_at IS NULL", entityID).First(&struct{}{}).Error == nil
	}
	if !exists {
		return "", errors.New(entityType + " not found")
	}

	// Validate file
	if file.Size > 10*1024*1024 { // 10MB
		return "", errors.New("file size exceeds 10MB limit")
	}

	contentType := file.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		return "", errors.New("invalid file type. Only images are allowed")
	}

	// Upload to Cloudinary
	f, err := file.Open()
	if err != nil {
		return "", errors.New("failed to open file")
	}
	defer f.Close()

	folder := entityType + "_images"
	url, err := s.cloudinary.UploadFile(context.Background(), f, folder)
	if err != nil {
		return "", errors.New("failed to upload file to Cloudinary")
	}

	// Update entity with image URL
	switch entityType {
	case "category":
		if err := s.db.Model(&categories.Category{}).Where("id = ?", entityID).Update("image_url", url).Error; err != nil {
			return "", err
		}
	case "subcategory":
		if err := s.db.Model(&categories.Subcategory{}).Where("id = ?", entityID).Update("image_url", url).Error; err != nil {
			return "", err
		}
	}

	return url, nil
}

func (s *imageService) DeleteImage(userID, entityType, entityID string) error {
	// Validate entity type
	if entityType != "category" && entityType != "subcategory" {
		return errors.New("invalid entity type")
	}

	// Check permission based on entity type
	permission := "update:" + entityType + "s"
	hasPerm, err := s.HasPermission(userID, permission)
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks " + permission + " permission")
	}

	// Validate entity ID
	if _, err := uuid.Parse(entityID); err != nil {
		return errors.New("invalid " + entityType + " ID format")
	}

	// Get current image URL
	var imageURL string
	switch entityType {
	case "category":
		var category categories.Category
		if err := s.db.Select("image_url").Where("id = ? AND deleted_at IS NULL", entityID).First(&category).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New(entityType + " not found")
			}
			return err
		}
		imageURL = category.ImageURL
	case "subcategory":
		var subcategory categories.Subcategory
		if err := s.db.Select("image_url").Where("id = ? AND deleted_at IS NULL", entityID).First(&subcategory).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New(entityType + " not found")
			}
			return err
		}
		imageURL = subcategory.ImageURL
	}

	if imageURL == "" {
		return nil
	}

	// Delete from Cloudinary
	if err := s.cloudinary.DeleteFile(context.Background(), imageURL); err != nil {
		return errors.New("failed to delete file from Cloudinary")
	}

	// Update entity to remove image URL
	switch entityType {
	case "category":
		if err := s.db.Model(&categories.Category{}).Where("id = ?", entityID).Update("image_url", "").Error; err != nil {
			return err
		}
	case "subcategory":
		if err := s.db.Model(&categories.Subcategory{}).Where("id = ?", entityID).Update("image_url", "").Error; err != nil {
			return err
		}
	}

	return nil
}

func isValidImageType(contentType string) bool {
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
