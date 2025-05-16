package authorization

import (
	"errors"

	"gorm.io/gorm"
	models "ticket-zetu-api/modules/users/models/authorization"
)

type RoleService interface {
	CreateRole(role *models.Role) error
	GetRoles(filters map[string]interface{}, limit, offset int) ([]models.Role, error)
	GetRoleByID(id string) (*models.Role, error)
	UpdateRole(id string, updates map[string]interface{}) error
	DeleteRole(id string) error
	HasPermission(userID, permissionName string) (bool, error)
	GetUserRoleLevel(userID string) (int, error)
	AssignPermissionToRole(roleID, permissionID, userID string) error
}

type roleService struct {
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) RoleService {
	return &roleService{db: db}
}

func (s *roleService) CreateRole(role *models.Role) error {
	// Ensure role is not soft-deleted with the same name
	var existingRole models.Role
	if err := s.db.Unscoped().Where("role_name = ?", role.RoleName).First(&existingRole).Error; err == nil {
		if existingRole.DeletedAt.Valid {
			return errors.New("a role with this name was previously deleted")
		}
		return errors.New("role name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Create the role
	return s.db.Create(role).Error
}

func (s *roleService) GetRoles(filters map[string]interface{}, limit, offset int) ([]models.Role, error) {
	var roles []models.Role
	query := s.db.Model(&models.Role{}).Preload("Permissions").Where("deleted_at IS NULL")

	for key, value := range filters {
		query = query.Where(key, value)
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&roles).Error
	return roles, err
}

func (s *roleService) GetRoleByID(id string) (*models.Role, error) {
	var role models.Role
	err := s.db.Preload("Permissions").Where("id = ? AND deleted_at IS NULL", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (s *roleService) UpdateRole(id string, updates map[string]interface{}) error {
	// Check if role exists and is not soft-deleted
	var role models.Role
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&role).Error; err != nil {
		return err
	}

	// If updating role_name, ensure it’s unique
	if name, ok := updates["role_name"]; ok {
		var existingRole models.Role
		if err := s.db.Unscoped().Where("role_name = ? AND id != ?", name, id).First(&existingRole).Error; err == nil {
			if existingRole.DeletedAt.Valid {
				return errors.New("a role with this name was previously deleted")
			}
			return errors.New("role name already exists")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	// Update NumberOfUsers if status changes to inactive or archived
	if status, ok := updates["status"]; ok && (status == string(models.RoleInactive) || status == string(models.RoleArchived)) {
		var userCount int64
		if err := s.db.Model(&models.User{}).Where("role_id = ? AND deleted_at IS NULL", id).Count(&userCount).Error; err != nil {
			return err
		}
		if userCount > 0 {
			return errors.New("cannot change status of role with assigned users")
		}
	}

	// Perform the update
	return s.db.Model(&models.Role{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).Error
}

func (s *roleService) DeleteRole(id string) error {
	// Check if role exists and is not soft-deleted
	var role models.Role
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&role).Error; err != nil {
		return err
	}

	// Prevent deletion of active roles
	if role.Status == models.RoleActive {
		return errors.New("cannot delete active role")
	}

	// Prevent deletion of roles with assigned users
	var userCount int64
	if err := s.db.Model(&models.User{}).Where("role_id = ? AND deleted_at IS NULL", id).Count(&userCount).Error; err != nil {
		return err
	}
	if userCount > 0 {
		return errors.New("cannot delete role with assigned users")
	}

	// Soft delete the role
	return s.db.Delete(&models.Role{}, "id = ?", id).Error
}

func (s *roleService) HasPermission(userID, permissionName string) (bool, error) {
	var count int64
	err := s.db.Model(&models.User{}).
		Joins("JOIN roles ON users.role_id = roles.id").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("users.id = ? AND permissions.permission_name = ? AND roles.status = ? AND permissions.status = ? AND users.deleted_at IS NULL AND roles.deleted_at IS NULL AND permissions.deleted_at IS NULL",
			userID, permissionName, models.RoleActive, models.PermissionActive).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *roleService) GetUserRoleLevel(userID string) (int, error) {
	var role models.Role
	err := s.db.Model(&models.User{}).
		Joins("JOIN roles ON users.role_id = roles.id").
		Where("users.id = ? AND roles.status = ? AND users.deleted_at IS NULL AND roles.deleted_at IS NULL", userID, models.RoleActive).
		Select("roles.level").First(&role).Error
	if err != nil {
		return 0, errors.New("user has no active role")
	}
	return role.Level, nil
}

func (s *roleService) AssignPermissionToRole(roleID, permissionID, userID string) error {
	// Check if role exists and is active
	var role models.Role
	if err := s.db.Where("id = ? AND status = ? AND deleted_at IS NULL", roleID, models.RoleActive).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found or inactive")
		}
		return err
	}

	// Check if permission exists and is active
	var permission models.Permission
	if err := s.db.Where("id = ? AND status = ? AND deleted_at IS NULL", permissionID, models.PermissionActive).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("permission not found or inactive")
		}
		return err
	}

	// Check if the role-permission assignment already exists
	var existingRolePermission models.RolePermission
	if err := s.db.Where("role_id = ? AND permission_id = ? AND deleted_at IS NULL", roleID, permissionID).First(&existingRolePermission).Error; err == nil {
		return errors.New("permission already assigned to role")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Create the role-permission assignment
	rolePermission := models.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
		CreatedBy:    userID,
	}
	return s.db.Create(&rolePermission).Error
}
