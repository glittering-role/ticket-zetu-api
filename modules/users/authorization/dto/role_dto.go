package dto

// RoleResponseDto defines the structure for role response
type RoleResponseDto struct {
	ID             string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoleName       string `json:"role_name" example:"Admin"`
	Description    string `json:"description" example:"Administrator role"`
	Level          int    `json:"level" example:"10"`
	Status         string `json:"status" example:"active"`
	IsSystemRole   bool   `json:"is_system_role" example:"false"`
	NumberOfUsers  int64  `json:"number_of_users" example:"5"`
	CreatedBy      string `json:"created_by" example:"550e8400-e29b-41d4-a716-446655440001"`
	LastModifiedBy string `json:"last_modified_by" example:"550e8400-e29b-41d4-a716-446655440001"`
	CreatedAt      string `json:"created_at" example:"2025-06-05T10:00:00Z"`
	UpdatedAt      string `json:"updated_at" example:"2025-10-01T12:00:00Z"`
	Version        int64  `json:"version" example:"1"`
}

// CreateRoleDto defines the structure for creating a role
type CreateRoleDto struct {
	RoleName     string `json:"role_name" validate:"required,max=100"`
	Description  string `json:"description" validate:"max=255"`
	Level        int64  `json:"level" validate:"gte=0"`
	IsSystemRole bool   `json:"is_system_role" validate:"boolean"`
}

// UpdateRoleDto defines the structure for updating a role
type UpdateRoleDto struct {
	RoleName     *string `json:"role_name" validate:"omit,max=100"`
	Description  *string `json:"description" validate:"omit,max=255"`
	Level        *int64  `json:"level" validate:"omit,gte=0"`
	Status       *string `json:"status" validate:"omit,oneof=active inactive archived"`
	IsSystemRole *bool   `json:"is_system_role" validate:"omit,boolean"`
}

// RoleAssignmentDto represents the request body for assigning a role to a user
type RoleAssignmentDto struct {
	UserID string `json:"user_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoleID string `json:"role_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440001"`
}
