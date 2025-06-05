package authorization

import (
	"errors"
	"strconv"
	"strings"

	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/authorization/dto"
	"ticket-zetu-api/modules/users/authorization/service"
	"ticket-zetu-api/modules/users/models/authorization" // Import models for status constants

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionController handles permission-related HTTP requests
type PermissionController struct {
	service    authorization_service.PermissionService
	logHandler *handler.LogHandler
	validator  *validator.Validate
}

// NewPermissionController initializes the controller
func NewPermissionController(service authorization_service.PermissionService, logHandler *handler.LogHandler) *PermissionController {
	return &PermissionController{
		service:    service,
		logHandler: logHandler,
		validator:  validator.New(),
	}
}

// CreatePermission godoc
// @Summary Create a new permission
// @Description Creates a new permission
// @Tags Authorization
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.CreatePermissionDto true "Permission details"
// @Success 201 {object} dto.PermissionResponseDto
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /permissions [post]
func (c *PermissionController) CreatePermission(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Check if user has permission to create permissions
	_, err := c.service.HasPermission(userID, "create:permission")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	// if !hasPerm {
	// 	return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks create:permission permission"), fiber.StatusForbidden)
	// }

	var input dto.CreatePermissionDto
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body"), fiber.StatusBadRequest)
	}

	// Validate DTO
	if err := c.validator.Struct(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "validation failed: "+err.Error()), fiber.StatusBadRequest)
	}

	// Additional validation for field lengths
	if len(input.Description) > 255 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "description too long"), fiber.StatusBadRequest)
	}
	if len(input.Scope) > 255 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "scope too long"), fiber.StatusBadRequest)
	}

	response, err := c.service.CreatePermission(&input, userID)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, response, "Permission created successfully", true)
}

// GetPermissions godoc
// @Summary List permissions
// @Description Get list of permissions with optional filters
// @Tags Authorization
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param permission_name query string false "Filter by permission name"
// @Param status query string false "Filter by status (active/inactive)"
// @Param limit query int false "Limit results" default(100)
// @Param offset query int false "Offset results" default(0)
// @Success 200 {array} dto.PermissionResponseDto
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /permissions [get]
func (c *PermissionController) GetPermissions(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	// Check if user has permission to view permissions
	_, err := c.service.HasPermission(userID, "view:permission")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	// if !hasPerm {
	// 	return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks view:permission permission"), fiber.StatusForbidden)
	// }

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
		if status != string(model.PermissionActive) && status != string(model.PermissionInactive) {
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

// GetPermission godoc
// @Summary Get permission by ID
// @Description Retrieve a permission by its ID
// @Tags Authorization
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Permission ID"
// @Success 200 {object} dto.PermissionResponseDto
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /permissions/{id} [get]
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

// UpdatePermission godoc
// @Summary Update permission
// @Description Update a permission by its ID
// @Tags Authorization
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Permission ID"
// @Param input body dto.UpdatePermissionDto true "Permission update details"
// @Success 200 {object} dto.PermissionResponseDto
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /permissions/{id} [put]
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

	var input dto.UpdatePermissionDto
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body"), fiber.StatusBadRequest)
	}

	// Validate DTO
	if err := c.validator.Struct(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "validation failed: "+err.Error()), fiber.StatusBadRequest)
	}

	// Additional validation for field lengths and status
	if input.PermissionName != nil && len(*input.PermissionName) > 100 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "permission_name too long"), fiber.StatusBadRequest)
	}
	if input.Description != nil && len(*input.Description) > 255 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "description too long"), fiber.StatusBadRequest)
	}
	if input.Scope != nil && len(*input.Scope) > 255 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "scope too long"), fiber.StatusBadRequest)
	}
	if input.Status != nil && *input.Status != string(model.PermissionActive) && *input.Status != string(model.PermissionInactive) {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid status value"), fiber.StatusBadRequest)
	}

	response, err := c.service.UpdatePermission(id, &input, userID)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, response, "Permission updated successfully", true)
}

// DeletePermission godoc
// @Summary Delete permission
// @Description Delete a permission by its ID
// @Tags Authorization
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Permission ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /permissions/{id} [delete]
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

	if err := c.service.DeletePermission(id); err != nil {
		if err.Error() == "cannot delete active permission" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, "cannot delete active permission"), fiber.StatusConflict)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Permission deleted successfully", true)
}
