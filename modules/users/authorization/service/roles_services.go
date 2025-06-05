package authorization_service

import (
	"errors"
	"ticket-zetu-api/modules/users/authorization/dto"
	"ticket-zetu-api/modules/users/authorization/utils"
	models "ticket-zetu-api/modules/users/models/authorization"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleService interface {
	CreateRole(roleDto *dto.CreateRoleDto, userID string) (*dto.RoleResponseDto, error)
	GetRoles(filters map[string]interface{}, limit, offset int) ([]dto.RoleResponseDto, error)
	GetRoleByID(id string) (*dto.RoleResponseDto, error)
	GetRoleWithPermissions(id string) (*dto.RoleResponseDto, []dto.PermissionResponseDto, error)
	UpdateRole(id string, roleDto *dto.UpdateRoleDto, userID string) (*dto.RoleResponseDto, error)
	DeleteRole(id string, userID string) error
	HasPermission(userID, permissionName string) (bool, error)
	GetUserRoleLevel(userID string) (int, error)
	AssignRoleToUser(userID, roleID, callerID string) error
}

type roleService struct {
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) RoleService {
	return &roleService{db: db}
}

// toRoleResponseDto converts a models.Role to dto.RoleResponseDto
func toRoleResponseDto(role *models.Role) *dto.RoleResponseDto {
	return &dto.RoleResponseDto{
		ID:             role.ID,
		RoleName:       role.RoleName,
		Description:    role.Description,
		Level:          role.Level,
		Status:         string(role.Status),
		IsSystemRole:   role.IsSystemRole,
		NumberOfUsers:  role.NumberOfUsers,
		CreatedBy:      role.CreatedBy,
		LastModifiedBy: role.LastModifiedBy,
		CreatedAt:      role.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      role.UpdatedAt.Format(time.RFC3339),
		Version:        int64(role.Version),
	}
}

func (s *roleService) CreateRole(roleDto *dto.CreateRoleDto, userID string) (*dto.RoleResponseDto, error) {
	var existingRole models.Role
	if err := s.db.Unscoped().Where("role_name = ?", roleDto.RoleName).First(&existingRole).Error; err == nil {
		if existingRole.DeletedAt.Valid {
			return nil, errors.New("a role with this name was previously deleted")
		}
		return nil, errors.New("role name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	role := &models.Role{
		ID:             uuid.New().String(),
		RoleName:       roleDto.RoleName,
		Description:    roleDto.Description,
		Level:          int(roleDto.Level),
		Status:         models.RoleActive,
		IsSystemRole:   roleDto.IsSystemRole,
		NumberOfUsers:  0,
		CreatedBy:      userID,
		LastModifiedBy: userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Version:        1,
	}

	if err := s.db.Create(role).Error; err != nil {
		return nil, err
	}

	return toRoleResponseDto(role), nil
}

func (s *roleService) GetRoles(filters map[string]interface{}, limit, offset int) ([]dto.RoleResponseDto, error) {
	var roles []models.Role
	query := s.db.Model(&models.Role{}).Where("deleted_at IS NULL")
	for key, value := range filters {
		query = query.Where(key, value)
	}
	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&roles).Error
	if err != nil {
		return nil, err
	}

	response := make([]dto.RoleResponseDto, len(roles))
	for i, role := range roles {
		response[i] = *toRoleResponseDto(&role)
	}
	return response, nil
}

func (s *roleService) GetRoleByID(id string) (*dto.RoleResponseDto, error) {
	var role models.Role
	err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return toRoleResponseDto(&role), nil
}

func (s *roleService) GetRoleWithPermissions(id string) (*dto.RoleResponseDto, []dto.PermissionResponseDto, error) {
	var role models.Role
	if err := s.db.Preload("RolePermissions.Permission").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("role not found")
		}
		return nil, nil, err
	}

	permDtos := make([]dto.PermissionResponseDto, 0, len(role.RolePermissions))
	for _, rp := range role.RolePermissions {
		if rp.Permission.DeletedAt.Valid {
			continue
		}
		permDtos = append(permDtos, *toPermissionResponseDto(&rp.Permission))
	}

	return toRoleResponseDto(&role), permDtos, nil
}

func (s *roleService) UpdateRole(id string, roleDto *dto.UpdateRoleDto, userID string) (*dto.RoleResponseDto, error) {
	var updatedRole models.Role
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var role models.Role
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("role not found")
			}
			return err
		}

		if role.IsSystemRole {
			return errors.New("cannot update system role")
		}

		updates := make(map[string]interface{})
		if roleDto.RoleName != nil {
			var existingRole models.Role
			if err := tx.Unscoped().Where("role_name = ? AND id != ?", *roleDto.RoleName, id).First(&existingRole).Error; err == nil {
				if existingRole.DeletedAt.Valid {
					return errors.New("a role with this name was previously deleted")
				}
				return errors.New("role name already exists")
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			updates["role_name"] = *roleDto.RoleName
		}
		if roleDto.Description != nil {
			updates["description"] = *roleDto.Description
		}
		if roleDto.Level != nil {
			updates["level"] = *roleDto.Level
		}
		if roleDto.Status != nil {
			updates["status"] = *roleDto.Status
		}
		if roleDto.IsSystemRole != nil {
			updates["is_system_role"] = *roleDto.IsSystemRole
		}
		updates["last_modified_by"] = userID
		updates["updated_at"] = time.Now()
		updates["version"] = role.Version + 1

		if status, ok := updates["status"]; ok && (status == string(models.RoleInactive) || status == string(models.RoleArchived)) {
			var userCount int64
			if err := tx.Table("user_profiles").Where("role_id = ? AND deleted_at IS NULL", id).Count(&userCount).Error; err != nil {
				return err
			}
			if userCount > 0 {
				return errors.New("cannot change status of role with assigned users")
			}
		}

		if len(updates) == 0 {
			updatedRole = role
			return nil
		}

		if err := tx.Model(&models.Role{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}

		if err := tx.Where("id = ?", id).First(&updatedRole).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return toRoleResponseDto(&updatedRole), nil
}

func (s *roleService) DeleteRole(id string, userID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var role models.Role
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("role not found")
			}
			return err
		}

		if role.IsSystemRole {
			return errors.New("cannot delete system role")
		}
		if role.Status == models.RoleActive {
			return errors.New("cannot delete active role")
		}
		var userCount int64
		if err := tx.Table("user_profiles").Where("role_id = ? AND deleted_at IS NULL", id).Count(&userCount).Error; err != nil {
			return err
		}
		if userCount > 0 {
			return errors.New("cannot delete role with assigned users")
		}

		updates := map[string]interface{}{
			"last_modified_by": userID,
			"updated_at":       time.Now(),
			"version":          role.Version + 1,
		}
		if err := tx.Model(&models.Role{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Role{}, "id = ?", id).Error
	})
}

func (s *roleService) HasPermission(userID, permissionName string) (bool, error) {
	var count int64
	err := s.db.Table("user_profiles").
		Joins("JOIN roles ON user_profiles.role_id = roles.id").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("user_profiles.id = ? AND permissions.permission_name = ? AND roles.status = ? AND permissions.status = ? AND user_profiles.deleted_at IS NULL AND roles.deleted_at IS NULL AND permissions.deleted_at IS NULL AND role_permissions.deleted_at IS NULL",
			userID, permissionName, models.RoleActive, models.PermissionActive).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *roleService) GetUserRoleLevel(userID string) (int, error) {
	var role models.Role
	err := s.db.Table("user_profiles").
		Joins("JOIN roles ON user_profiles.role_id = roles.id").
		Where("user_profiles.id = ? AND roles.status = ? AND user_profiles.deleted_at IS NULL AND roles.deleted_at IS NULL", userID, models.RoleActive).
		Select("roles.level").First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("user has no active role")
		}
		return 0, err
	}
	return role.Level, nil
}

func (s *roleService) AssignRoleToUser(userID, roleID, callerID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var role models.Role
		if err := tx.Where("id = ? AND status = ? AND deleted_at IS NULL", roleID, models.RoleActive).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("role not found or inactive")
			}
			return err
		}

		var user members.User
		if err := tx.Table("user_profiles").Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			return err
		}

		callerLevel, err := s.GetUserRoleLevel(callerID)
		if err != nil {
			return err
		}
		if role.Level > callerLevel {
			return errors.New("cannot assign role with higher level than caller's")
		}

		if user.RoleID != "" && user.RoleID != roleID {
			var oldRole models.Role
			if err := tx.Where("id = ? AND deleted_at IS NULL", user.RoleID).First(&oldRole).Error; err == nil {
				if oldRole.IsSystemRole {
					return errors.New("cannot modify user with system role")
				}
				if err := tx.Model(&models.Role{}).
					Where("id = ?", user.RoleID).
					Updates(map[string]interface{}{
						"number_of_users":  gorm.Expr("number_of_users - 1"),
						"last_modified_by": callerID,
						"updated_at":       time.Now(),
						"version":          oldRole.Version + 1,
					}).Error; err != nil {
					return err
				}
			}
		}

		if role.IsSystemRole && user.RoleID != roleID {
			return errors.New("cannot assign system role directly")
		}

		if err := utils.UpdateRoleUserCount(tx, roleID, 1); err != nil {
			return err
		}

		if err := tx.Table("user_profiles").
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"role_id":          roleID,
				"last_modified_by": callerID,
				"updated_at":       time.Now(),
			}).Error; err != nil {
			return err
		}

		return nil
	})
}
