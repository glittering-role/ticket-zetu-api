package dto

// CreatePermissionDto defines the structure for creating a permission
type CreatePermissionDto struct {
	PermissionName string `json:"permission_name" validate:"required"`
	Description    string `json:"description,omitempty"`
	Scope          string `json:"scope,omitempty"`
}

// UpdatePermissionDto defines the structure for updating a permission
type UpdatePermissionDto struct {
	PermissionName *string `json:"permission_name,omitempty"`
	Description    *string `json:"description,omitempty"`
	Scope          *string `json:"scope,omitempty"`
	Status         *string `json:"status,omitempty"`
}

// PermissionResponseDto defines the structure for permission response
type PermissionResponseDto struct {
	ID             string `json:"id"`
	PermissionName string `json:"permission_name"`
	Description    string `json:"description"`
	Scope          string `json:"scope"`
	Status         string `json:"status"`
	CreatedBy      string `json:"created_by"`
	LastModifiedBy string `json:"last_modified_by"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type PermissionAssignmentDto struct {
	RoleID       string `json:"role_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	PermissionID string `json:"permission_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440001"`
}
