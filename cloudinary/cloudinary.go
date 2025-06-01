package cloudinary

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Config holds Cloudinary configuration
type Config struct {
	CloudName string
	APIKey    string
	APISecret string
}

// CloudinaryService wraps the Cloudinary client and configuration
type CloudinaryService struct {
	Client    *cloudinary.Cloudinary
	CloudName string
}

// NewCloudinaryService initializes the Cloudinary client
func NewCloudinaryService(config Config) (*CloudinaryService, error) {
	// Validate configuration
	if config.CloudName == "" {
		return nil, errors.New("cloudinary cloud name is required")
	}
	if config.APIKey == "" {
		return nil, errors.New("cloudinary API key is required")
	}
	if config.APISecret == "" {
		return nil, errors.New("cloudinary API secret is required")
	}

	cld, err := cloudinary.NewFromParams(config.CloudName, config.APIKey, config.APISecret)
	if err != nil {
		log.Printf("Failed to initialize Cloudinary for cloud %s: %v", config.CloudName, err)
		return nil, err
	}
	return &CloudinaryService{
		Client:    cld,
		CloudName: config.CloudName,
	}, nil
}

// UploadFile uploads a file to Cloudinary and returns the secure URL
func (s *CloudinaryService) UploadFile(ctx context.Context, file interface{}, folder string) (string, error) {
	resp, err := s.Client.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folder,
	})
	if err != nil {
		log.Printf("Failed to upload file to Cloudinary in folder %s: %v", folder, err)
		return "", err
	}
	return resp.SecureURL, nil
}

// DeleteFile deletes a file from Cloudinary using the public ID and resource type
func (s *CloudinaryService) DeleteFile(ctx context.Context, url string) error {
	publicID, resourceType := extractPublicID(url, s.CloudName)
	if publicID == "" {
		log.Printf("No valid public ID extracted from URL: %s", url)
		return nil // No error, just nothing to delete
	}

	_, err := s.Client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: resourceType,
	})
	if err != nil {
		log.Printf("Failed to delete file from Cloudinary (public ID: %s, resource type: %s): %v", publicID, resourceType, err)
		return err
	}
	return nil
}

// extractPublicID extracts the public ID and resource type from a Cloudinary URL
func extractPublicID(url, cloudName string) (string, string) {

	var resourceType string
	var prefix string

	// Check for image or video prefix
	imagePrefix := "https://res.cloudinary.com/" + cloudName + "/image/upload/"
	videoPrefix := "https://res.cloudinary.com/" + cloudName + "/video/upload/"
	if strings.HasPrefix(url, imagePrefix) {
		resourceType = "image"
		prefix = imagePrefix
	} else if strings.HasPrefix(url, videoPrefix) {
		resourceType = "video"
		prefix = videoPrefix
	} else {
		return "", ""
	}

	// Remove prefix
	publicID := strings.TrimPrefix(url, prefix)

	// Handle versioned URLs (e.g., /v1234567890/)
	if strings.HasPrefix(publicID, "v") {
		parts := strings.SplitN(publicID, "/", 2)
		if len(parts) == 2 {
			publicID = parts[1]
		}
	}

	// Remove file extension
	parts := strings.Split(publicID, ".")
	if len(parts) < 2 {
		return "", ""
	}
	publicID = strings.Join(parts[:len(parts)-1], ".")

	return publicID, resourceType
}
