package organizers_services

import (
	"context"
	"errors"
	"mime/multipart"
	"ticket-zetu-api/cloudinary"
	organizers "ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/users/authorization"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationImageService interface {
	AddImage(userID, organizationID string, file *multipart.FileHeader) (string, error)
	DeleteImage(userID, organizationID string) error
}

type organizationImageService struct {
	*organizerService
	cloudinary *cloudinary.CloudinaryService
}

func NewOrganizationImageService(db *gorm.DB, authService authorization.PermissionService, cloudinary *cloudinary.CloudinaryService) OrganizationImageService {
	base := NewBaseService(db, authService, cloudinary)

	return &organizationImageService{
		organizerService: base,
		cloudinary:       cloudinary,
	}
}

func (s *organizationImageService) AddImage(userID, organizationID string, file *multipart.FileHeader) (string, error) {
	// Check permission
	_, err := s.HasPermission(userID, "update:organizations")
	if err != nil {
		return "", err
	}
	// if !hasPerm {
	// 	return "", errors.New("user lacks update:organizations permission")
	// }

	// Validate organization ID
	if _, err := uuid.Parse(organizationID); err != nil {
		return "", errors.New("invalid organization ID format")
	}

	// Check if organization exists
	if err := s.db.Model(&organizers.Organizer{}).Where("id = ? AND deleted_at IS NULL", organizationID).First(&struct{}{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("organization not found")
		}
		return "", err
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

	folder := "organization_images"
	url, err := s.cloudinary.UploadFile(context.Background(), f, folder)
	if err != nil {
		return "", errors.New("failed to upload file to Cloudinary")
	}

	// Update organization with image URL
	if err := s.db.Model(&organizers.Organizer{}).Where("id = ?", organizationID).Update("image_url", url).Error; err != nil {
		return "", err
	}

	return url, nil
}

func (s *organizationImageService) DeleteImage(userID, organizationID string) error {
	// Check permission
	hasPerm, err := s.HasPermission(userID, "update:organizations")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks update:organizations permission")
	}

	// Validate organization ID
	if _, err := uuid.Parse(organizationID); err != nil {
		return errors.New("invalid organization ID format")
	}

	// Get current image URL
	var organization organizers.Organizer
	if err := s.db.Select("image_url").Where("id = ? AND deleted_at IS NULL", organizationID).First(&organization).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("organization not found")
		}
		return err
	}

	if organization.ImageURL == "" {
		return nil
	}

	// Delete from Cloudinary
	if err := s.cloudinary.DeleteFile(context.Background(), organization.ImageURL); err != nil {
		return errors.New("failed to delete file from Cloudinary")
	}

	// Update organization to remove image URL
	if err := s.db.Model(&organizers.Organizer{}).Where("id = ?", organizationID).Update("image_url", "").Error; err != nil {
		return err
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
