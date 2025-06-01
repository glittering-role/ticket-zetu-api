package services

import (
	"errors"
	"ticket-zetu-api/modules/events/category/dto"
	"ticket-zetu-api/modules/events/models/categories"
	"ticket-zetu-api/modules/users/authorization"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubcategoryService interface {
	GetSubcategories(userID, categoryID string) ([]dto.SubcategoryDTO, error)
	CreateSubcategory(userID, categoryID, name, description, imageURL string) (*dto.SubcategoryDTO, error)
	UpdateSubcategory(userID, id, name, description, imageURL string, isActive bool) (*dto.SubcategoryDTO, error)
	DeleteSubcategory(userID, id string) error
}

type subcategoryService struct {
	*BaseService
}

func NewSubcategoryService(db *gorm.DB, authService authorization.PermissionService) SubcategoryService {
	return &subcategoryService{
		BaseService: NewBaseService(db, authService),
	}
}

func (s *subcategoryService) toDTO(subcategory *categories.Subcategory) *dto.SubcategoryDTO {
	var categoryName, categoryImage string
	if subcategory.CategoryID != "" {
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

	var subcategoriesDTO []dto.SubcategoryDTO
	for _, subcategory := range subcategories {
		subcategoriesDTO = append(subcategoriesDTO, *s.toDTO(&subcategory))
	}
	return subcategoriesDTO, nil
}

func (s *subcategoryService) CreateSubcategory(userID, categoryID, name, description, imageURL string) (*dto.SubcategoryDTO, error) {
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

	// Preload the category to include in the DTO
	if err := s.db.Preload("Category").First(&subcategory, "id = ?", subcategory.ID).Error; err != nil {
		return nil, err
	}

	return s.toDTO(&subcategory), nil
}

func (s *subcategoryService) UpdateSubcategory(userID, id, name, description, imageURL string, isActive bool) (*dto.SubcategoryDTO, error) {
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

	// Preload the category to include in the DTO
	if err := s.db.Preload("Category").First(&subcategory, "id = ?", subcategory.ID).Error; err != nil {
		return nil, err
	}

	return s.toDTO(&subcategory), nil
}

func (s *subcategoryService) DeleteSubcategory(userID, id string) error {
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
