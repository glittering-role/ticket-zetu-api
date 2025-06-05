package services

import (
	"errors"
	"ticket-zetu-api/cloudinary"
	"ticket-zetu-api/modules/users/authorization/service"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseService struct {
	db                   *gorm.DB
	authorizationService authorization_service.PermissionService
	cloudinary           *cloudinary.CloudinaryService
}

func NewBaseService(db *gorm.DB, authService authorization_service.PermissionService) *BaseService {
	return &BaseService{
		db:                   db,
		authorizationService: authService,
	}
}

func (s *BaseService) HasPermission(userID, permission string) (bool, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return false, errors.New("invalid user ID format")
	}
	hasPerm, err := s.authorizationService.HasPermission(userID, permission)
	if err != nil {
		return false, err
	}
	return hasPerm, nil
}
