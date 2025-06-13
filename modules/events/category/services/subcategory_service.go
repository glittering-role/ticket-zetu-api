package services

import (
	"errors"
	"ticket-zetu-api/modules/events/category/dto"
	"ticket-zetu-api/modules/events/models/categories"
	"ticket-zetu-api/modules/users/authorization/service"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubcategoryService interface {
	GetSubcategories(userID, categoryID string) ([]dto.SubcategoryDTO, error)
	CreateSubcategory(userID, categoryID, name, description string) (*dto.SubcategoryDTO, error)
	UpdateSubcategory(userID, id, name, description string) (*dto.SubcategoryDTO, error)
	DeleteSubcategory(userID, id string) error
	ToggleSubCategoryStatus(userID, id string, isActive bool) error
}

type subcategoryService struct {
	*BaseService
}

func NewSubcategoryService(db *gorm.DB, authService authorization_service.PermissionService) SubcategoryService {
	return &subcategoryService{
		BaseService: NewBaseService(db, authService),
	}
}

func (s *subcategoryService) toDTO(subcategory *categories.Subcategory) *dto.SubcategoryDTO {
	var categoryName, categoryImage string
	if subcategory.Category.ID != "" {
		categoryName = subcategory.Category.Name
		categoryImage = subcategory.Category.ImageURL
	}

	return &dto.SubcategoryDTO{
		ID:            subcategory.ID,
		Name:          subcategory.Name,
		Description:   subcategory.Description,
		ImageURL:      subcategory.ImageURL,
		IsActive:      subcategory.IsActive,
		LastUpdatedBy: subcategory.LastUpdatedBy,
		CreatedAt:     subcategory.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     subcategory.UpdatedAt.Format(time.RFC3339),
		DeletedAt:     formatTimePointer(getDeletedAtTime(subcategory.DeletedAt)),
		CategoryID:    subcategory.CategoryID,
		CategoryName:  categoryName,
		CategoryImage: categoryImage,
	}
}

func (s *subcategoryService) GetSubcategories(userID, categoryID string) ([]dto.SubcategoryDTO, error) {

	_, err := s.HasPermission(userID, "read:subcategories")
	if err != nil {
		return nil, err
	}

	// if !hasPerm {
	// 	return nil, errors.New("user lacks read:subcategories permission")
	// }

	// Validate category ID format
	if _, err := uuid.Parse(categoryID); err != nil {
		return nil, errors.New("invalid category ID format")
	}

	// Check if category exists
	var category categories.Category
	if err := s.db.Where("id = ? AND deleted_at IS NULL", categoryID).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	var subcategories []categories.Subcategory
	if err := s.db.
		Preload("Category").
		Where("category_id = ? AND deleted_at IS NULL", categoryID).
		Find(&subcategories).Error; err != nil {
		return nil, err
	}

	if len(subcategories) == 0 {
		return []dto.SubcategoryDTO{}, nil
	}

	var subcategoriesDTO []dto.SubcategoryDTO
	for _, subcategory := range subcategories {
		subcategoriesDTO = append(subcategoriesDTO, *s.toDTO(&subcategory))
	}
	return subcategoriesDTO, nil
}

func (s *subcategoryService) CreateSubcategory(userID, categoryID, name, description string) (*dto.SubcategoryDTO, error) {
	_, err := s.HasPermission(userID, "create:subcategories")
	if err != nil {
		return nil, err
	}

	// if !hasPerm {
	// 	return nil, errors.New("user lacks create:subcategories permission")
	// }

	if _, err := uuid.Parse(categoryID); err != nil {
		return nil, errors.New("invalid category ID format")
	}

	// Check if category exists and is active
	var category categories.Category
	if err := s.db.Where("id = ? AND deleted_at IS NULL AND is_active = ?", categoryID, true).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found or inactive")
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
		IsActive:      true,
		LastUpdatedBy: userID,
	}

	if err := s.db.Create(&subcategory).Error; err != nil {
		return nil, err
	}

	// Preload the category to include in the DTO
	if err := s.db.Preload("Category").First(&subcategory, "id = ?", subcategory.ID).Error; err != nil {
		return nil, err
	}

	return s.toDTO(&subcategory), nil
}

func (s *subcategoryService) UpdateSubcategory(userID, id, name, description string) (*dto.SubcategoryDTO, error) {
	// hasPerm, err := s.HasPermission(userID, "update:subcategories")
	// if err != nil {
	// 	return nil, err
	// }

	// if !hasPerm {
	// 	return nil, errors.New("user lacks update:subcategories permission")
	// }

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid subcategory ID format")
	}

	var subcategory categories.Subcategory
	if err := s.db.Preload("Category").Where("id = ? AND deleted_at IS NULL", id).First(&subcategory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("subcategory not found")
		}
		return nil, err
	}

	// Check if category is active
	if !subcategory.Category.IsActive {
		return nil, errors.New("parent category is inactive")
	}

	// Check if subcategory name already exists within the category (excluding current subcategory)
	var existingSubcategory categories.Subcategory
	if err := s.db.Where("category_id = ? AND name = ? AND id != ? AND deleted_at IS NULL", subcategory.CategoryID, name, id).First(&existingSubcategory).Error; err == nil {
		return nil, errors.New("subcategory name already exists in this category")
	}

	subcategory.Name = name
	subcategory.Description = description
	subcategory.LastUpdatedBy = userID

	if err := s.db.Save(&subcategory).Error; err != nil {
		return nil, err
	}

	return s.toDTO(&subcategory), nil
}

func (s *subcategoryService) DeleteSubcategory(userID, id string) error {
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

	// Use soft delete
	if err := s.db.Model(&subcategory).Update("deleted_at", time.Now()).Error; err != nil {
		return err
	}

	return nil
}

func (s *subcategoryService) ToggleSubCategoryStatus(userID, id string, isActive bool) error {
	// Optional permission check
	// hasPerm, err := s.HasPermission(userID, "update:subcategories")
	// if err != nil {
	// 	return err
	// }
	// if !hasPerm {
	// 	return errors.New("user lacks update:subcategories permission")
	// }

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid subcategory ID format")
	}

	// Find subcategory
	var subcategory categories.Subcategory
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&subcategory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("subcategory not found")
		}
		return err
	}

	// ✅ Early return if the status is already set
	if subcategory.IsActive == isActive {
		return errors.New("subcategory status already set")
	}

	// ✅ Ensure parent category is active before activating subcategory
	if isActive {
		var category categories.Category
		if err := s.db.Where("id = ? AND deleted_at IS NULL", subcategory.CategoryID).First(&category).Error; err != nil {
			return errors.New("parent category not found")
		}
		if !category.IsActive {
			return errors.New("cannot activate subcategory with inactive parent category")
		}
	}

	// Update subcategory status
	subcategory.IsActive = isActive
	subcategory.LastUpdatedBy = userID

	if err := s.db.Save(&subcategory).Error; err != nil {
		return err
	}

	return nil
}
