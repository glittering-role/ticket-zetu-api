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

type PermissionController struct {
	service    PermissionService
	logHandler *handler.LogHandler
	db         *gorm.DB // Added for role level check in AssignPermissionToRole
}

func NewPermissionController(service PermissionService, logHandler *handler.LogHandler, db *gorm.DB) *PermissionController {
	return &PermissionController{
		service:    service,
		logHandler: logHandler,
		db:         db,
	}
}

// CreatePermission creates a new permission with validation
func (c *PermissionController) CreatePermission(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string) // Assume user_id is set by auth middleware

	// Check if user has permission to create permissions
	// hasPerm, err := c.service.HasPermission(userID, "create:permission")
	// if err != nil {
	// 	return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	// }
	// if !hasPerm {
	// 	return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks create:permission permission"), fiber.StatusForbidden)
	// }

	var permission models.Permission
	if err := ctx.BodyParser(&permission); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	// Validate permission input
	if err := validatePermissionInput(&permission); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	// Set CreatedBy
	permission.CreatedBy = userID
	permission.LastModifiedBy = userID

	if err := c.service.CreatePermission(&permission); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, permission, "Permission created successfully", true)
}

// GetPermissions retrieves permissions with optional filters
func (c *PermissionController) GetPermissions(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Check if user has permission to view permissions
	hasPerm, err := c.service.HasPermission(userID, "view:permission")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks view:permission permission"), fiber.StatusForbidden)
	}

	filters := make(map[string]interface{})

	// Validate and add permission_name filter
	if permissionName := strings.TrimSpace(ctx.Query("permission_name")); permissionName != "" {
		if len(permissionName) > 100 {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "permission_name too long"), fiber.StatusBadRequest)
		}
		filters["permission_name = ?"] = permissionName
	}

	// Validate and add status filter
	if status := strings.TrimSpace(ctx.Query("status")); status != "" {
		if status != string(models.PermissionActive) && status != string(models.PermissionInactive) {
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

	permissions, err := c.service.GetPermissions(filters, limit, offset)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, permissions, "Permissions retrieved successfully", true)
}

// GetPermission retrieves a single permission by ID
func (c *PermissionController) GetPermission(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Check if user has permission to view permissions
	hasPerm, err := c.service.HasPermission(userID, "view:permission")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks view:permission permission"), fiber.StatusForbidden)
	}

	id := ctx.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid ID format"), fiber.StatusBadRequest)
	}

	permission, err := c.service.GetPermissionByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, "permission not found"), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, permission, "Permission retrieved successfully", true)
}

// UpdatePermission updates an existing permission
func (c *PermissionController) UpdatePermission(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Check if user has permission to update permissions
	hasPerm, err := c.service.HasPermission(userID, "update:permission")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks update:permission permission"), fiber.StatusForbidden)
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
	if err := validatePermissionUpdates(updates); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	// Set LastModifiedBy
	updates["last_modified_by"] = userID

	if err := c.service.UpdatePermission(id, updates); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	permission, err := c.service.GetPermissionByID(id)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, permission, "Permission updated successfully", true)
}

// DeletePermission deletes a permission if it's not active
func (c *PermissionController) DeletePermission(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Check if user has permission to delete permissions
	hasPerm, err := c.service.HasPermission(userID, "delete:permission")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks delete:permission permission"), fiber.StatusForbidden)
	}

	id := ctx.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid ID format"), fiber.StatusBadRequest)
	}

	permission, err := c.service.GetPermissionByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, "permission not found"), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	// Prevent deletion of active permissions
	if permission.Status == models.PermissionActive {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, "cannot delete active permission"), fiber.StatusConflict)
	}

	if err := c.service.DeletePermission(id); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Permission deleted successfully", true)
}



// validatePermissionInput validates permission creation input
func validatePermissionInput(permission *models.Permission) error {
	if strings.TrimSpace(permission.PermissionName) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "permission name is required")
	}
	if len(permission.PermissionName) > 100 {
		return fiber.NewError(fiber.StatusBadRequest, "permission name too long")
	}
	if permission.Status != models.PermissionActive && permission.Status != models.PermissionInactive {
		return fiber.NewError(fiber.StatusBadRequest, "invalid status value")
	}
	if permission.Description != "" && len(permission.Description) > 255 {
		return fiber.NewError(fiber.StatusBadRequest, "description too long")
	}
	if permission.Scope != "" && len(permission.Scope) > 255 {
		return fiber.NewError(fiber.StatusBadRequest, "scope too long")
	}
	return nil
}

// validatePermissionUpdates validates permission update fields
func validatePermissionUpdates(updates map[string]interface{}) error {
	for field, value := range updates {
		switch field {
		case "permission_name":
			name, ok := value.(string)
			if !ok || strings.TrimSpace(name) == "" {
				return fiber.NewError(fiber.StatusBadRequest, "invalid permission name value")
			}
			if len(name) > 100 {
				return fiber.NewError(fiber.StatusBadRequest, "permission name too long")
			}
		case "status":
			status, ok := value.(string)
			if !ok || (status != string(models.PermissionActive) && status != string(models.PermissionInactive)) {
				return fiber.NewError(fiber.StatusBadRequest, "invalid status value")
			}
		case "description":
			if desc, ok := value.(string); ok && len(desc) > 255 {
				return fiber.NewError(fiber.StatusBadRequest, "description too long")
			}
		case "scope":
			if scope, ok := value.(string); ok && len(scope) > 255 {
				return fiber.NewError(fiber.StatusBadRequest, "scope too long")
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
