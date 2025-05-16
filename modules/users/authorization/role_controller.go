package authorization

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"ticket-zetu-api/logs/handler"
	models "ticket-zetu-api/modules/users/models/authorization"
)

type RoleController struct {
	service    RoleService
	logHandler *handler.LogHandler
}

func NewRoleController(service RoleService, logHandler *handler.LogHandler) *RoleController {
	return &RoleController{
		service:    service,
		logHandler: logHandler,
	}
}

// CreateRole creates a new role with validation
func (c *RoleController) CreateRole(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string) // Assume user_id is set by auth middleware

	// Check if user has permission to create roles
	hasPerm, err := c.service.HasPermission(userID, "create:role")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks create:role permission"), fiber.StatusForbidden)
	}

	var role models.Role
	if err := ctx.BodyParser(&role); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	// Validate role input
	if err := validateRoleInput(&role); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	// Ensure the role level is valid based on the user's max role level
	userMaxLevel, err := c.service.GetUserRoleLevel(userID)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if role.Level > userMaxLevel {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "cannot create role with higher level"), fiber.StatusForbidden)
	}

	// Set CreatedBy
	role.CreatedBy = userID
	role.LastModifiedBy = userID

	if err := c.service.CreateRole(&role); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, role, "Role created successfully", true)
}

// GetRoles retrieves roles with optional filters
func (c *RoleController) GetRoles(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Check if user has permission to view roles
	hasPerm, err := c.service.HasPermission(userID, "view:role")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks view:role permission"), fiber.StatusForbidden)
	}

	filters := make(map[string]interface{})

	// Validate and add role_name filter
	if roleName := strings.TrimSpace(ctx.Query("role_name")); roleName != "" {
		if len(roleName) > 100 {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "role_name too long"), fiber.StatusBadRequest)
		}
		filters["role_name = ?"] = roleName
	}

	// Validate and add status filter
	if status := strings.TrimSpace(ctx.Query("status")); status != "" {
		if status != string(models.RoleActive) && status != string(models.RoleInactive) && status != string(models.RoleArchived) {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid status value"), fiber.StatusBadRequest)
		}
		filters["status = ?"] = status
	}

	// Validate pagination parameters
	limit, err := strconv.Atoi(ctx.Query("limit", "100"))
	if err != nil || limit < 1 || limit > 1000 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid limit value"), fiber.StatusBadRequest)
	}

	offset, err := strconv.Atoi(ctx.Query("offset", "0"))
	if err != nil || offset < 0 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid offset value"), fiber.StatusBadRequest)
	}

	roles, err := c.service.GetRoles(filters, limit, offset)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, roles, "Roles retrieved successfully", true)
}

// GetRole retrieves a single role by ID
func (c *RoleController) GetRole(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Check if user has permission to view roles
	hasPerm, err := c.service.HasPermission(userID, "view:role")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks view:role permission"), fiber.StatusForbidden)
	}

	id := ctx.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid ID format"), fiber.StatusBadRequest)
	}

	role, err := c.service.GetRoleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, "role not found"), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, role, "Role retrieved successfully", true)
}

// UpdateRole updates an existing role
func (c *RoleController) UpdateRole(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Check if user has permission to update roles
	hasPerm, err := c.service.HasPermission(userID, "update:role")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks update:role permission"), fiber.StatusForbidden)
	}

	id := ctx.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid ID format"), fiber.StatusBadRequest)
	}

	var updates map[string]interface{}
	if err := ctx.BodyParser(&updates); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	// Validate update fields
	if err := validateRoleUpdates(updates); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	// If updating level, ensure it's not higher than the user's max role level
	if level, ok := updates["level"]; ok {
		userMaxLevel, err := c.service.GetUserRoleLevel(userID)
		if err != nil {
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
		if l, ok := level.(float64); ok && int(l) > userMaxLevel {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "cannot update role to higher level"), fiber.StatusForbidden)
		}
	}

	// Set LastModifiedBy
	updates["last_modified_by"] = userID

	if err := c.service.UpdateRole(id, updates); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	role, err := c.service.GetRoleByID(id)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, role, "Role updated successfully", true)
}

// DeleteRole deletes a role if it's not active
func (c *RoleController) DeleteRole(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Check if user has permission to delete roles
	hasPerm, err := c.service.HasPermission(userID, "delete:role")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks delete:role permission"), fiber.StatusForbidden)
	}

	id := ctx.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid ID format"), fiber.StatusBadRequest)
	}

	role, err := c.service.GetRoleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, "role not found"), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	// Prevent deletion of active roles
	if role.Status == models.RoleActive {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, "cannot delete active role"), fiber.StatusConflict)
	}

	// Prevent deletion of roles with assigned users
	if role.NumberOfUsers > 0 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, "cannot delete role with assigned users"), fiber.StatusConflict)
	}

	if err := c.service.DeleteRole(id); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Role deleted successfully", true)
}

// AssignPermissionToRole assigns a permission to a role
func (c *RoleController) AssignPermissionToRole(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Check if user has permission to assign permissions
	hasPerm, err := c.service.HasPermission(userID, "assign:permission")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks assign:permission permission"), fiber.StatusForbidden)
	}

	var input struct {
		RoleID       string `json:"role_id"`
		PermissionID string `json:"permission_id"`
	}
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	if _, err := uuid.Parse(input.RoleID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid role ID format"), fiber.StatusBadRequest)
	}
	if _, err := uuid.Parse(input.PermissionID); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid permission ID format"), fiber.StatusBadRequest)
	}

	// Check if the role level is valid for the user's max role level
	userMaxLevel, err := c.service.GetUserRoleLevel(userID)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	role, err := c.service.GetRoleByID(input.RoleID)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusNotFound)
	}
	if role.Level > userMaxLevel {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "cannot assign permission to role with higher level"), fiber.StatusForbidden)
	}

	// Assign permission
	if err := c.service.AssignPermissionToRole(input.RoleID, input.PermissionID, userID); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Permission assigned to role successfully", true)
}

// validateRoleInput validates role creation input
func validateRoleInput(role *models.Role) error {
	if strings.TrimSpace(role.RoleName) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "role name is required")
	}
	if len(role.RoleName) > 100 {
		return fiber.NewError(fiber.StatusBadRequest, "role name too long")
	}
	if role.Status != models.RoleActive && role.Status != models.RoleInactive && role.Status != models.RoleArchived {
		return fiber.NewError(fiber.StatusBadRequest, "invalid status value")
	}
	if role.Level < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "level must be positive")
	}
	if role.Description != "" && len(role.Description) > 255 {
		return fiber.NewError(fiber.StatusBadRequest, "description too long")
	}
	return nil
}

// validateRoleUpdates validates role update fields
func validateRoleUpdates(updates map[string]interface{}) error {
	for field, value := range updates {
		switch field {
		case "role_name":
			name, ok := value.(string)
			if !ok || strings.TrimSpace(name) == "" {
				return fiber.NewError(fiber.StatusBadRequest, "invalid role name value")
			}
			if len(name) > 100 {
				return fiber.NewError(fiber.StatusBadRequest, "role name too long")
			}
		case "status":
			status, ok := value.(string)
			if !ok || (status != string(models.RoleActive) && status != string(models.RoleInactive) && status != string(models.RoleArchived)) {
				return fiber.NewError(fiber.StatusBadRequest, "invalid status value")
			}
		case "description":
			if desc, ok := value.(string); ok && len(desc) > 255 {
				return fiber.NewError(fiber.StatusBadRequest, "description too long")
			}
		case "level":
			if level, ok := value.(float64); !ok || level < 1 {
				return fiber.NewError(fiber.StatusBadRequest, "invalid level value")
			}
		case "is_system_role":
			if _, ok := value.(bool); !ok {
				return fiber.NewError(fiber.StatusBadRequest, "invalid is_system_role value")
			}
		case "last_modified_by":
			if _, err := uuid.Parse(value.(string)); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "invalid last_modified_by ID format")
			}
		default:
			return fiber.NewError(fiber.StatusBadRequest, "invalid field: "+field)
		}
	}
	return nil
}
