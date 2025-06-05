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

type CategoryService interface {
	GetCategories(userID string) ([]dto.CategoryDTO, error)
	GetCategory(userID, id string) (*dto.CategoryDTO, error)
	CreateCategory(userID, name, description string) (*dto.CategoryDTO, error)
	UpdateCategory(userID, id, name, description string) (*dto.CategoryDTO, error)
	DeleteCategory(userID, id string) error
	ToggleCategoryStatus(userID, id string, isActive bool) error
	GetAllCategoriesWithTheirSubCategories(userID string) ([]dto.CategoryDTO, error)
}

type categoryService struct {
	*BaseService
}

func NewCategoryService(db *gorm.DB, authService authorization_service.PermissionService) CategoryService {
	return &categoryService{
		BaseService: NewBaseService(db, authService),
	}
}

func (s *categoryService) toDTO(category *categories.Category) *dto.CategoryDTO {
	var subcategoriesDTO []dto.SubcategoryDTO
	for _, sub := range category.Subcategories {
		subcategoriesDTO = append(subcategoriesDTO, dto.SubcategoryDTO{
			ID:          sub.ID,
			Name:        sub.Name,
			Description: sub.Description,
			ImageURL:    sub.ImageURL,
			IsActive:    sub.IsActive,
			CategoryID:  sub.CategoryID,
			CreatedAt:   sub.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   sub.UpdatedAt.Format(time.RFC3339),
			DeletedAt:   formatTimePointer(getDeletedAtTime(sub.DeletedAt)),
		})
	}

	return &dto.CategoryDTO{
		ID:            category.ID,
		Name:          category.Name,
		Description:   category.Description,
		ImageURL:      category.ImageURL,
		IsActive:      category.IsActive,
		LastUpdatedBy: category.LastUpdatedBy,
		CreatedAt:     category.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     category.UpdatedAt.Format(time.RFC3339),
		DeletedAt:     formatTimePointer(getDeletedAtTime(category.DeletedAt)),
		Subcategories: subcategoriesDTO,
	}
}

func formatTimePointer(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := t.Format(time.RFC3339)
	return &formatted
}

func getDeletedAtTime(deletedAt gorm.DeletedAt) *time.Time {
	if deletedAt.Valid {
		return &deletedAt.Time
	}
	return nil
}

// services/category_service.go

func (s *categoryService) GetAllCategoriesWithTheirSubCategories(userID string) ([]dto.CategoryDTO, error) {
	// Add permission check if needed
	// if err := s.authService.CheckPermission(userID, "view:categories"); err != nil {
	//     return nil, err
	// }

	var categories []categories.Category
	if err := s.db.
		Preload("Subcategories"). // Preload subcategories
		Where("deleted_at IS NULL AND is_active = ?", true).
		Find(&categories).Error; err != nil {
		return nil, err
	}

	var categoriesDTO []dto.CategoryDTO
	for _, category := range categories {
		categoriesDTO = append(categoriesDTO, *s.toDTO(&category))
	}

	return categoriesDTO, nil
}

func (s *categoryService) GetCategories(userID string) ([]dto.CategoryDTO, error) {
	// Add permission check if needed
	// if err := s.authService.CheckPermission(userID, "read:categories"); err != nil {
	//     return nil, err
	// }

	var categories []categories.Category
	if err := s.db.
		Where("deleted_at IS NULL AND is_active = ?", true).
		Find(&categories).Error; err != nil {
		return nil, err
	}

	var categoriesDTO []dto.CategoryDTO
	for _, category := range categories {
		categoriesDTO = append(categoriesDTO, *s.toDTO(&category))
	}
	return categoriesDTO, nil
}

func (s *categoryService) GetCategory(userID, id string) (*dto.CategoryDTO, error) {
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
	return s.toDTO(&category), nil
}

func (s *categoryService) CreateCategory(userID, name, description string) (*dto.CategoryDTO, error) {
	// Add permission check
	_, err := s.HasPermission(userID, "create:categories")
	if err != nil {
		return nil, err
	}

	// if !hasPerm {
	// 	return nil, errors.New("user lacks create:categories permission")
	// }

	var existingCategory categories.Category
	if err := s.db.Where("name = ? AND deleted_at IS NULL", name).First(&existingCategory).Error; err == nil {
		return nil, errors.New("category name already exists")
	}

	category := categories.Category{
		Name:          name,
		Description:   description,
		IsActive:      true,
		LastUpdatedBy: userID,
	}

	if err := s.db.Create(&category).Error; err != nil {
		return nil, err
	}

	return s.toDTO(&category), nil
}

func (s *categoryService) UpdateCategory(userID, id, name, description string) (*dto.CategoryDTO, error) {
	// Add permission check
	_, err := s.HasPermission(userID, "update:categories")
	if err != nil {
		return nil, err
	}

	// if !hasPerm {
	// 	return nil, errors.New("user lacks update:categories permission")
	// }

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid category ID format")
	}

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
	category.LastUpdatedBy = userID

	if err := s.db.Save(&category).Error; err != nil {
		return nil, err
	}

	return s.toDTO(&category), nil
}

func (s *categoryService) DeleteCategory(userID, id string) error {
	// Add permission check
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

func (s *categoryService) ToggleCategoryStatus(userID, id string, isActive bool) error {
	// Add permission check
	_, err := s.HasPermission(userID, "update:categories")
	if err != nil {
		return err
	}
	// if !hasPerm {
	// 	return errors.New("user lacks update:categories permission")
	// }

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

	category.IsActive = isActive
	category.LastUpdatedBy = userID

	if err := s.db.Save(&category).Error; err != nil {
		return err
	}

	return nil
}
