package authorization_service

import (
	"errors"
	"ticket-zetu-api/modules/users/authorization/dto"
	model "ticket-zetu-api/modules/users/models/authorization"
	"time"

	"gorm.io/gorm"
)

type PermissionService interface {
	CreatePermission(permissionDto *dto.CreatePermissionDto, userID string) (*dto.PermissionResponseDto, error)
	GetPermissions(filters map[string]interface{}, limit, offset int) ([]dto.PermissionResponseDto, error)
	GetPermissionByID(id string) (*dto.PermissionResponseDto, error)
	GetPermissionWithRoles(id string) (*dto.PermissionResponseDto, []dto.RoleResponseDto, error)
	UpdatePermission(id string, permissionDto *dto.UpdatePermissionDto, userID string) (*dto.PermissionResponseDto, error)
	DeletePermission(id string) error
	AssignPermissionToRole(roleID, permissionID, userID string) error
	RemovePermissionFromRole(roleID, permissionID, userID string) error
	HasPermission(userID, permissionName string) (bool, error)
	GetUserRoleLevel(userID string) (int, error)
}

type permissionService struct {
	db *gorm.DB
}

// NewPermissionService initializes the service
func NewPermissionService(db *gorm.DB) PermissionService {
	return &permissionService{db: db}
}

// toPermissionResponseDto converts a models.Permission to dto.PermissionResponseDto
func toPermissionResponseDto(permission *model.Permission) *dto.PermissionResponseDto {
	return &dto.PermissionResponseDto{
		ID:             permission.ID,
		PermissionName: permission.PermissionName,
		Description:    permission.Description,
		Scope:          permission.Scope,
		Status:         string(permission.Status),
		CreatedBy:      permission.CreatedBy,
		LastModifiedBy: permission.LastModifiedBy,
		CreatedAt:      permission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      permission.UpdatedAt.Format(time.RFC3339),
	}
}

// checkPermissionExists checks if a permission exists (including soft-deleted ones) by name
func (s *permissionService) checkPermissionExists(name, excludeID string) error {
	var permission model.Permission
	query := s.db.Unscoped().Where("permission_name = ?", name)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	if err := query.First(&permission).Error; err == nil {
		if permission.DeletedAt.Valid {
			return errors.New("a permission with this name was previously deleted")
		}
		return errors.New("permission name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

// getActivePermission retrieves an active permission by ID
func (s *permissionService) getActivePermission(id string) (*model.Permission, error) {
	var permission model.Permission
	err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, err
	}
	return &permission, nil
}

// getActiveRole retrieves an active role by ID
func (s *permissionService) getActiveRole(id string) (*model.Role, error) {
	var role model.Role
	err := s.db.Where("id = ? AND status = ? AND deleted_at IS NULL", id, model.RoleActive).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found or inactive")
		}
		return nil, err
	}
	return &role, nil
}

// checkRolePermissionExists checks if a role-permission assignment exists
func (s *permissionService) checkRolePermissionExists(roleID, permissionID string) (*model.RolePermission, error) {
	var rolePermission model.RolePermission
	err := s.db.Where("role_id = ? AND permission_id = ? AND deleted_at IS NULL", roleID, permissionID).First(&rolePermission).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err == nil {
		return &rolePermission, nil
	}
	return nil, nil
}

// countRolePermissions counts active role-permission assignments for a permission
func (s *permissionService) countRolePermissions(permissionID string) (int64, error) {
	var count int64
	err := s.db.Model(&model.RolePermission{}).Where("permission_id = ? AND deleted_at IS NULL", permissionID).Count(&count).Error
	return count, err
}

func (s *permissionService) CreatePermission(permissionDto *dto.CreatePermissionDto, userID string) (*dto.PermissionResponseDto, error) {
	if err := s.checkPermissionExists(permissionDto.PermissionName, ""); err != nil {
		return nil, err
	}

	permission := &model.Permission{
		PermissionName: permissionDto.PermissionName,
		Description:    permissionDto.Description,
		Scope:          permissionDto.Scope,
		Status:         model.PermissionActive,
		CreatedBy:      userID,
		LastModifiedBy: userID,
	}

	if err := s.db.Create(permission).Error; err != nil {
		return nil, err
	}

	return toPermissionResponseDto(permission), nil
}

func (s *permissionService) GetPermissions(filters map[string]interface{}, limit, offset int) ([]dto.PermissionResponseDto, error) {
	var permissions []model.Permission
	query := s.db.Model(&model.Permission{}).Where("deleted_at IS NULL")
	for key, value := range filters {
		query = query.Where(key, value)
	}
	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&permissions).Error
	if err != nil {
		return nil, err
	}

	response := make([]dto.PermissionResponseDto, len(permissions))
	for i, perm := range permissions {
		response[i] = *toPermissionResponseDto(&perm)
	}
	return response, nil
}

func (s *permissionService) GetPermissionByID(id string) (*dto.PermissionResponseDto, error) {
	permission, err := s.getActivePermission(id)
	if err != nil {
		return nil, err
	}
	return toPermissionResponseDto(permission), nil
}

func (s *permissionService) GetPermissionWithRoles(id string) (*dto.PermissionResponseDto, []dto.RoleResponseDto, error) {
	permission, err := s.getActivePermission(id)
	if err != nil {
		return nil, nil, err
	}

	var roles []model.Role
	err = s.db.Table("roles").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Where("role_permissions.permission_id = ? AND role_permissions.deleted_at IS NULL AND roles.deleted_at IS NULL", id).
		Find(&roles).Error
	if err != nil {
		return nil, nil, err
	}

	roleDtos := make([]dto.RoleResponseDto, len(roles))
	for i, role := range roles {
		roleDtos[i] = *toRoleResponseDto(&role)
	}

	return toPermissionResponseDto(permission), roleDtos, nil
}

func (s *permissionService) UpdatePermission(id string, permissionDto *dto.UpdatePermissionDto, userID string) (*dto.PermissionResponseDto, error) {
	permission, err := s.getActivePermission(id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if permissionDto.PermissionName != nil {
		if err := s.checkPermissionExists(*permissionDto.PermissionName, id); err != nil {
			return nil, err
		}
		updates["permission_name"] = *permissionDto.PermissionName
	}
	if permissionDto.Description != nil {
		updates["description"] = *permissionDto.Description
	}
	if permissionDto.Scope != nil {
		updates["scope"] = *permissionDto.Scope
	}
	if permissionDto.Status != nil {
		updates["status"] = *permissionDto.Status
	}
	updates["last_modified_by"] = userID
	updates["updated_at"] = time.Now()

	if status, ok := updates["status"]; ok && status == string(model.PermissionInactive) {
		count, err := s.countRolePermissions(id)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("cannot change status of permission assigned to roles")
		}
	}

	if len(updates) == 0 {
		return toPermissionResponseDto(permission), nil
	}

	if err := s.db.Model(&model.Permission{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}

	updatedPermission, err := s.getActivePermission(id)
	if err != nil {
		return nil, err
	}
	return toPermissionResponseDto(updatedPermission), nil
}

func (s *permissionService) DeletePermission(id string) error {
	permission, err := s.getActivePermission(id)
	if err != nil {
		return err
	}

	if permission.Status == model.PermissionActive {
		return errors.New("cannot delete active permission")
	}

	count, err := s.countRolePermissions(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("cannot delete permission assigned to roles")
	}

	return s.db.Delete(&model.Permission{}, "id = ?", id).Error
}

func (s *permissionService) AssignPermissionToRole(roleID, permissionID, userID string) error {
	if _, err := s.getActiveRole(roleID); err != nil {
		return err
	}
	if _, err := s.getActivePermission(permissionID); err != nil {
		return err
	}

	if existing, err := s.checkRolePermissionExists(roleID, permissionID); err != nil {
		return err
	} else if existing != nil {
		return errors.New("permission already assigned to role")
	}

	rolePermission := model.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
		CreatedBy:    userID,
	}
	return s.db.Create(&rolePermission).Error
}

func (s *permissionService) HasPermission(userID, permissionName string) (bool, error) {
	var count int64
	err := s.db.Table("user_profiles").
		Joins("JOIN roles ON user_profiles.role_id = roles.id").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("user_profiles.id = ? AND permissions.permission_name = ? AND roles.status = ? AND permissions.status = ? AND user_profiles.deleted_at IS NULL AND roles.deleted_at IS NULL AND permissions.deleted_at IS NULL AND role_permissions.deleted_at IS NULL",
			userID, permissionName, model.RoleActive, model.PermissionActive).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *permissionService) GetUserRoleLevel(userID string) (int, error) {
	var role model.Role
	err := s.db.Table("user_profiles").
		Joins("JOIN roles ON user_profiles.role_id = roles.id").
		Where("user_profiles.id = ? AND roles.status = ? AND user_profiles.deleted_at IS NULL AND roles.deleted_at IS NULL", userID, model.RoleActive).
		Select("roles.level").First(&role).Error
	if err != nil {
		return 0, errors.New("user has no active role")
	}
	return role.Level, nil
}

func (s *permissionService) RemovePermissionFromRole(roleID, permissionID string, userID string) error {
	userMaxLevel, err := s.GetUserRoleLevel(userID)
	if err != nil {
		return err
	}

	role, err := s.getActiveRole(roleID)
	if err != nil {
		return err
	}
	if role.Level > userMaxLevel {
		return errors.New("cannot remove permission from role with higher level")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		rolePermission, err := s.checkRolePermissionExists(roleID, permissionID)
		if err != nil {
			return err
		}
		if rolePermission == nil {
			return errors.New("role-permission assignment not found")
		}

		if err := tx.Model(&model.Role{}).Where("id = ?", roleID).Update("last_modified_by", userID).Error; err != nil {
			return err
		}

		return tx.Delete(rolePermission).Error
	})
}
