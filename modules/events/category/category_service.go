package category

import (
	"errors"
	"ticket-zetu-api/modules/events/models/categories"
	"ticket-zetu-api/modules/users/authorization"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryService interface {
	GetCategories(userID string) ([]categories.Category, error)
	GetCategory(userID, id string) (*categories.Category, error)
	GetSubcategories(userID, categoryID string) ([]categories.Subcategory, error)
	CreateCategory(userID, name, description, imageURL string) (*categories.Category, error)
	UpdateCategory(userID, id, name, description, imageURL string, isActive bool) (*categories.Category, error)
	DeleteCategory(userID, id string) error
	CreateSubcategory(userID, categoryID, name, description, imageURL string) (*categories.Subcategory, error)
	UpdateSubcategory(userID, id, name, description, imageURL string, isActive bool) (*categories.Subcategory, error)
	DeleteSubcategory(userID, id string) error
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

func (s *categoryService) CreateSubcategory(userID, categoryID, name, description, imageURL string) (*categories.Subcategory, error) {
	hasPerm, err := s.HasPermission(userID, "create:subcategories")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks create:subcategories permission")
	}

	if _, err := uuid.Parse(categoryID); err != nil {
		return nil, errors.New("invalid category ID format")
	}

	// Check if category exists and is active
	var category categories.Category
	if err := s.db.Where("id = ? AND deleted_at IS NULL AND is_active = ?", categoryID, true).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	// Check if subcategory name already exists within the category
	var existingSubcategory categories.Subcategory
	if err := s.db.Where("category_id = ? AND name = ? AND deleted_at IS NULL", categoryID, name).First(&existingSubcategory).Error; err == nil {
		return nil, errors.New("subcategory name already exists in this category")
	}

	subcategory := categories.Subcategory{
		CategoryID:    categoryID,
		Name:          name,
		Description:   description,
		ImageURL:      imageURL,
		IsActive:      true,
		LastUpdatedBy: userID,
	}

	if err := s.db.Create(&subcategory).Error; err != nil {
		return nil, err
	}

	return &subcategory, nil
}

func (s *categoryService) UpdateSubcategory(userID, id, name, description, imageURL string, isActive bool) (*categories.Subcategory, error) {
	hasPerm, err := s.HasPermission(userID, "update:subcategories")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks update:subcategories permission")
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid subcategory ID format")
	}

	var subcategory categories.Subcategory
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&subcategory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("subcategory not found")
		}
		return nil, err
	}

	// Check if category exists and is active
	var category categories.Category
	if err := s.db.Where("id = ? AND deleted_at IS NULL AND is_active = ?", subcategory.CategoryID, true).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("parent category not found")
		}
		return nil, err
	}

	// Check if subcategory name already exists within the category (excluding current subcategory)
	var existingSubcategory categories.Subcategory
	if err := s.db.Where("category_id = ? AND name = ? AND id != ? AND deleted_at IS NULL", subcategory.CategoryID, name, id).First(&existingSubcategory).Error; err == nil {
		return nil, errors.New("subcategory name already exists in this category")
	}

	subcategory.Name = name
	subcategory.Description = description
	subcategory.ImageURL = imageURL
	subcategory.IsActive = isActive
	subcategory.LastUpdatedBy = userID

	if err := s.db.Save(&subcategory).Error; err != nil {
		return nil, err
	}

	return &subcategory, nil
}

func (s *categoryService) DeleteSubcategory(userID, id string) error {
	hasPerm, err := s.HasPermission(userID, "delete:subcategories")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks delete:subcategories permission")
	}

	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid subcategory ID format")
	}

	var subcategory categories.Subcategory
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&subcategory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("subcategory not found")
		}
		return err
	}

	if subcategory.IsActive {
		return errors.New("cannot delete an active subcategory")
	}

	if err := s.db.Delete(&subcategory).Error; err != nil {
		return err
	}

	return nil
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

func (s *categoryService) CreateCategory(userID, name, description, imageURL string) (*categories.Category, error) {
	hasPerm, err := s.HasPermission(userID, "create:categories")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks create:categories permission")
	}

	// Check if category name already exists
	var existingCategory categories.Category
	if err := s.db.Where("name = ? AND deleted_at IS NULL", name).First(&existingCategory).Error; err == nil {
		return nil, errors.New("category name already exists")
	}

	category := categories.Category{
		Name:          name,
		Description:   description,
		ImageURL:      imageURL,
		IsActive:      true,
		LastUpdatedBy: userID,
	}

	if err := s.db.Create(&category).Error; err != nil {
		return nil, err
	}

	return &category, nil
}

func (s *categoryService) UpdateCategory(userID, id, name, description, imageURL string, isActive bool) (*categories.Category, error) {
	hasPerm, err := s.HasPermission(userID, "update:categories")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks update:categories permission")
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid category ID format")
	}

	// Check if category name already exists (excluding current category)
	var existingCategory categories.Category
	if err := s.db.Where("name = ? AND id != ? AND deleted_at IS NULL", name, id).First(&existingCategory).Error; err == nil {
		return nil, errors.New("category name already exists")
	}

	var category categories.Category
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	category.Name = name
	category.Description = description
	category.ImageURL = imageURL
	category.IsActive = isActive
	category.LastUpdatedBy = userID

	if err := s.db.Save(&category).Error; err != nil {
		return nil, err
	}

	return &category, nil
}

func (s *categoryService) DeleteCategory(userID, id string) error {
	hasPerm, err := s.HasPermission(userID, "delete:categories")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks delete:categories permission")
	}

	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid category ID format")
	}

	var category categories.Category
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("category not found")
		}
		return err
	}

	if category.IsActive {
		return errors.New("cannot delete an active category")
	}

	if err := s.db.Delete(&category).Error; err != nil {
		return err
	}

	return nil
}
