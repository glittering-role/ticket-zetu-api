package category

import (
	"errors"
	"ticket-zetu-api/modules/tickets/models/categories"
	"ticket-zetu-api/modules/users/authorization"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryService interface {
	GetCategories(userID string) ([]categories.Category, error)
	GetCategory(userID, id string) (*categories.Category, error)
	GetSubcategories(userID, categoryID string) ([]categories.Subcategory, error)
	HasPermission(userID, permission string) (bool, error)
}

type categoryService struct {
	db                   *gorm.DB
	authorizationService authorization.PermissionService
}

func NewCategoryService(db *gorm.DB, authService authorization.PermissionService) CategoryService {
	return &categoryService{
		db:                   db,
		authorizationService: authService,
	}
}

func (s *categoryService) HasPermission(userID, permission string) (bool, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return false, errors.New("invalid user ID format")
	}
	hasPerm, err := s.authorizationService.HasPermission(userID, permission)
	if err != nil {
		return false, err
	}
	return hasPerm, nil
}

func (s *categoryService) GetCategories(userID string) ([]categories.Category, error) {
	hasPerm, err := s.HasPermission(userID, "view:categories")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks view:categories permission")
	}

	var categories []categories.Category
	if err := s.db.
		Preload("Subcategories").
		Where("deleted_at IS NULL AND is_active = ?", true).
		Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (s *categoryService) GetCategory(userID, id string) (*categories.Category, error) {
	hasPerm, err := s.HasPermission(userID, "view:categories")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks view:categories permission")
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid category ID format")
	}
	var category categories.Category
	if err := s.db.
		Preload("Subcategories").
		Where("id = ? AND deleted_at IS NULL AND is_active = ?", id, true).
		First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	return &category, nil
}

func (s *categoryService) GetSubcategories(userID, categoryID string) ([]categories.Subcategory, error) {
	hasPerm, err := s.HasPermission(userID, "view:categories")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks view:categories permission")
	}

	if _, err := uuid.Parse(categoryID); err != nil {
		return nil, errors.New("invalid category ID format")
	}
	var subcategories []categories.Subcategory
	if err := s.db.
		Preload("Category").
		Where("category_id = ? AND deleted_at IS NULL AND is_active = ?", categoryID, true).
		Find(&subcategories).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("subcategories not found")
		}
		return nil, err
	}
	return subcategories, nil
}
