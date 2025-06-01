package dto

// category dto
type CategoryDTO struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	ImageURL      string           `json:"image_url"`
	IsActive      bool             `json:"is_active"`
	LastUpdatedBy string           `json:"last_updated_by"`
	CreatedAt     string           `json:"created_at"`
	UpdatedAt     string           `json:"updated_at"`
	DeletedAt     *string          `json:"deleted_at,omitempty"`
	Subcategories []SubcategoryDTO `json:"subcategories,omitempty"`
}
