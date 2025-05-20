package authorization

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"ticket-zetu-api/logs/handler"
	models "ticket-zetu-api/modules/users/models/authorization"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

func (c *RoleController) CreateRole(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)
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
	if err := validateRoleInput(&role); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}
	userMaxLevel, err := c.service.GetUserRoleLevel(userID)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if role.Level > userMaxLevel {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "cannot create role with higher level"), fiber.StatusForbidden)
	}
	role.CreatedBy = userID
	role.LastModifiedBy = userID
	if err := c.service.CreateRole(&role); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, role, "Role created successfully", true)
}

func (c *RoleController) GetRoles(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(string)

	log.Printf("GetRoles called by User ID: %s", userID)

	hasPerm, err := c.service.HasPermission(userID, "view:role")
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if !hasPerm {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "user lacks view:role permission"), fiber.StatusForbidden)
	}
	filters := make(map[string]interface{})
	if roleName := strings.TrimSpace(ctx.Query("role_name")); roleName != "" {
		if len(roleName) > 100 {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusBadRequest, "role_name too long"), fiber.StatusBadRequest)
		}
		filters["role_name = ?"] = roleName
	}
	if status := strings.TrimSpace(ctx.Query("status")); status != "" {
		if status != string(models.RoleActive) && status != string(models.RoleInactive) && status != string(models.RoleArchived) {
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
	return c.logHandler.LogSuccess(ctx, roles, "Roles retrieved successfully", true)
}

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, "role not found"), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, role, "Role retrieved successfully", true)
}

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
	var updates map[string]interface{}
	if err := ctx.BodyParser(&updates); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}
	if err := validateRoleUpdates(updates); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}
	if level, ok := updates["level"]; ok {
		userMaxLevel, err := c.service.GetUserRoleLevel(userID)
		if err != nil {
			return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
		}
		if l, ok := level.(float64); ok && int(l) > userMaxLevel {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusForbidden, "cannot update role to higher level"), fiber.StatusForbidden)
		}
	}
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
	role, err := c.service.GetRoleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusNotFound, "role not found"), fiber.StatusNotFound)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	if role.Status == models.RoleActive {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, "cannot delete active role"), fiber.StatusConflict)
	}
	if role.NumberOfUsers > 0 {
		return c.logHandler.LogError(ctx, fiber.NewError(fiber.StatusConflict, "cannot delete role with assigned users"), fiber.StatusConflict)
	}
	if err := c.service.DeleteRole(id, userID); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	return c.logHandler.LogSuccess(ctx, nil, "Role deleted successfully", true)
}

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
