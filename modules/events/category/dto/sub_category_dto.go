package dto

// sub category dto
type SubcategoryDTO struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	ImageURL      string  `json:"image_url"`
	IsActive      bool    `json:"is_active"`
	LastUpdatedBy string  `json:"last_updated_by"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	DeletedAt     *string `json:"deleted_at,omitempty"`
	CategoryID    string  `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	CategoryImage string  `json:"category_image"`
}

type CreateSubcategoryDto struct {
	CategoryID  string `json:"category_id" example:"7d9eeb25-5b88-4c51-a3de-45a4dfd5f0f2" validate:"required,uuid"`
	Name        string `json:"name" example:"Live Concerts" validate:"required,min=2,max=50"`
	Description string `json:"description,omitempty" example:"Events involving live musical performances" validate:"omitempty,max=255"`
}

type UpdateSubSubcategoryDto struct {
	Name        string `json:"name" example:"Live Concerts" validate:"required,min=2,max=50"`
	Description string `json:"description,omitempty" example:"Events involving live musical performances" validate:"omitempty,max=255"`
}

type ToggleCategoryStatus struct {
	IsActive bool `json:"is_active" validate:"required"`
}
