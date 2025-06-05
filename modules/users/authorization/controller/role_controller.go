package authorization

import (
	"strconv"
	"strings"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/authorization/dto"
	"ticket-zetu-api/modules/users/authorization/service"
	model "ticket-zetu-api/modules/users/models/authorization"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RoleController handles role-related HTTP requests
type RoleController struct {
	service    authorization_service.RoleService
	logHandler *handler.LogHandler
	validator  *validator.Validate
}

// NewRoleController initializes the controller
func NewRoleController(service authorization_service.RoleService, logHandler *handler.LogHandler) *RoleController {
	return &RoleController{
		service:    service,
		logHandler: logHandler,
		validator:  validator.New(),
	}
}

// CreateRole godoc
// @Summary Create a new role
// @Description Creates a new role if the user has sufficient permissions
// @Tags Roles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body dto.CreateRoleDto true "Role details"
// @Success 201 {object} SuccessResponse{data=dto.RoleResponseDto}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /roles [post]
func (c *RoleController) CreateRole(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	_, err := c.service.HasPermission(userID, "create:role")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	// if !hasPerm {
	// 	return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks create:role permission"), fiber.StatusForbidden)
	// }

	var input dto.CreateRoleDto
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "validation failed: "+err.Error()), fiber.StatusBadRequest)
	}

	userMaxLevel, err := c.service.GetUserRoleLevel(userID)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if input.Level > int64(userMaxLevel) {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "cannot create role with higher level"), fiber.StatusForbidden)
	}

	role, err := c.service.CreateRole(&input, userID)
	if err != nil {
		if err.Error() == "role name already exists" || err.Error() == "a role with this name was previously deleted" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return ctx.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Message: "Role created successfully",
		Data:    role,
	})
}

// GetRoles godoc
// @Summary List roles
// @Description Retrieves a list of roles with optional filters
// @Tags Roles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role_name query string false "Filter by role name" example:"Admin"
// @Param status query string false "Filter by status (active/inactive/archived)" example:"active"
// @Param limit query int false "Limit results" example:"100"
// @Param offset query int false "Offset results" example:"0"
// @Success 200 {object} SuccessResponse{data=[]dto.RoleResponseDto}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /roles [get]
func (c *RoleController) GetRoles(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	_, err := c.service.HasPermission(userID, "view:role")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	// if !hasPerm {
	// 	return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks view:role permission"), fiber.StatusForbidden)
	// }

	filters := make(map[string]interface{})
	if roleName := strings.TrimSpace(ctx.Query("role_name")); roleName != "" {
		if len(roleName) > 100 {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "role_name too long"), fiber.StatusBadRequest)
		}
		filters["role_name = ?"] = roleName
	}
	if status := strings.TrimSpace(ctx.Query("status")); status != "" {
		if status != string(model.RoleActive) && status != string(model.RoleInactive) && status != string(model.RoleArchived) {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid status value"), fiber.StatusBadRequest)
		}
		filters["status = ?"] = status
	}

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

	return c.logHandler.LogSuccess(ctx, roles, "Permission created successfully", true)

}

// GetRole godoc
// @Summary Get a role by ID
// @Description Retrieves a role by its ID
// @Tags Roles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Role ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success 200 {object} SuccessResponse{data=dto.RoleResponseDto}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /roles/{id} [get]
func (c *RoleController) GetRole(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

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
		if err.Error() == "role not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, role, "Role retrieved successfully", true)

}

// UpdateRole godoc
// @Summary Update a role
// @Description Updates a role by its ID if the user has sufficient permissions
// @Tags Roles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Role ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Param input body dto.UpdateRoleDto true "Role update details"
// @Success 200 {object} SuccessResponse{data=dto.RoleResponseDto}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /roles/{id} [put]
func (c *RoleController) UpdateRole(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

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

	var input dto.UpdateRoleDto
	if err := ctx.BodyParser(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "invalid request body"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(&input); err != nil {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "validation failed: "+err.Error()), fiber.StatusBadRequest)
	}

	if input.Level != nil {
		userMaxLevel, err := c.service.GetUserRoleLevel(userID)
		if err != nil {
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
		if *input.Level > int64(userMaxLevel) {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "cannot update role to higher level"), fiber.StatusForbidden)
		}
	}

	_, err = c.service.UpdateRole(id, &input, userID)
	if err != nil {
		if err.Error() == "role not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "role name already exists" || err.Error() == "a role with this name was previously deleted" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		}
		if err.Error() == "cannot update system role" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, err.Error()), fiber.StatusForbidden)
		}
		if err.Error() == "cannot change status of role with assigned users" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Role updated successfully", true)
}

// DeleteRole godoc
// @Summary Delete a role
// @Description Deletes a role by its ID if the user has sufficient permissions
// @Tags Roles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Role ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /roles/{id} [delete]
func (c *RoleController) DeleteRole(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

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

	err = c.service.DeleteRole(id, userID)
	if err != nil {
		if err.Error() == "role not found" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, err.Error()), fiber.StatusNotFound)
		}
		if err.Error() == "cannot delete system role" || err.Error() == "cannot delete active role" || err.Error() == "cannot delete role with assigned users" {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, err.Error()), fiber.StatusConflict)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Role deleted successfully", true)
}
