package authorization

import (
	"errors"

	"gorm.io/gorm"
	models "ticket-zetu-api/modules/users/models/authorization"
)

type PermissionService interface {
	CreatePermission(permission *models.Permission) error
	GetPermissions(filters map[string]interface{}, limit, offset int) ([]models.Permission, error)
	GetPermissionByID(id string) (*models.Permission, error)
	GetPermissionWithRoles(id string) (*models.Permission, []models.Role, error) // New method
	UpdatePermission(id string, updates map[string]interface{}) error
	DeletePermission(id string) error
	AssignPermissionToRole(roleID, permissionID, userID string) error
	RemovePermissionFromRole(roleID, permissionID, userID string) error
	HasPermission(userID, permissionName string) (bool, error)
	GetUserRoleLevel(userID string) (int, error)
}

type permissionService struct {
	db *gorm.DB
}

func NewPermissionService(db *gorm.DB) PermissionService {
	return &permissionService{db: db}
}

func (s *permissionService) CreatePermission(permission *models.Permission) error {
	// Ensure permission is not soft-deleted with the same name
	var existingPermission models.Permission
	if err := s.db.Unscoped().Where("permission_name = ?", permission.PermissionName).First(&existingPermission).Error; err == nil {
		if existingPermission.DeletedAt.Valid {
			return errors.New("a permission with this name was previously deleted")
		}
		return errors.New("permission name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Create the permission
	return s.db.Create(permission).Error
}

func (s *permissionService) GetPermissions(filters map[string]interface{}, limit, offset int) ([]models.Permission, error) {
	var permissions []models.Permission
	query := s.db.Model(&models.Permission{}).Where("deleted_at IS NULL")

	for key, value := range filters {
		query = query.Where(key, value)
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&permissions).Error
	return permissions, err
}

func (s *permissionService) GetPermissionByID(id string) (*models.Permission, error) {
	var permission models.Permission
	err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// New method to get permission with its roles
func (s *permissionService) GetPermissionWithRoles(id string) (*models.Permission, []models.Role, error) {
	var permission models.Permission
	err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&permission).Error
	if err != nil {
		return nil, nil, err
	}

	// Get related roles through role_permissions table
	var roles []models.Role
	err = s.db.Table("roles").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Where("role_permissions.permission_id = ? AND role_permissions.deleted_at IS NULL AND roles.deleted_at IS NULL", id).
		Find(&roles).Error

	return &permission, roles, err
}

func (s *permissionService) UpdatePermission(id string, updates map[string]interface{}) error {
	// Check if permission exists and is not soft-deleted
	var permission models.Permission
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&permission).Error; err != nil {
		return err
	}

	// If updating permission_name, ensure it's unique
	if name, ok := updates["permission_name"]; ok {
		var existingPermission models.Permission
		if err := s.db.Unscoped().Where("permission_name = ? AND id != ?", name, id).First(&existingPermission).Error; err == nil {
			if existingPermission.DeletedAt.Valid {
				return errors.New("a permission with this name was previously deleted")
			}
			return errors.New("permission name already exists")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	// Check if permission is assigned to any roles if status is changing to inactive
	if status, ok := updates["status"]; ok && status == string(models.PermissionInactive) {
		var roleCount int64
		if err := s.db.Model(&models.RolePermission{}).Where("permission_id = ? AND deleted_at IS NULL", id).Count(&roleCount).Error; err != nil {
			return err
		}
		if roleCount > 0 {
			return errors.New("cannot change status of permission assigned to roles")
		}
	}

	// Perform the update
	return s.db.Model(&models.Permission{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).Error
}

func (s *permissionService) DeletePermission(id string) error {
	// Check if permission exists and is not soft-deleted
	var permission models.Permission
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&permission).Error; err != nil {
		return err
	}

	// Prevent deletion of active permissions
	if permission.Status == models.PermissionActive {
		return errors.New("cannot delete active permission")
	}

	// Prevent deletion of permissions assigned to roles
	var roleCount int64
	if err := s.db.Model(&models.RolePermission{}).Where("permission_id = ? AND deleted_at IS NULL", id).Count(&roleCount).Error; err != nil {
		return err
	}
	if roleCount > 0 {
		return errors.New("cannot delete permission assigned to roles")
	}

	// Soft delete the permission
	return s.db.Delete(&models.Permission{}, "id = ?", id).Error
}

func (s *permissionService) AssignPermissionToRole(roleID, permissionID, userID string) error {
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

func (s *permissionService) RemovePermissionFromRole(roleID, permissionID, userID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Check if the role-permission assignment exists
		var rolePermission models.RolePermission
		if err := tx.Where("role_id = ? AND permission_id = ? AND deleted_at IS NULL", roleID, permissionID).First(&rolePermission).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("role-permission assignment not found")
			}
			return err
		}

		// Update LastModifiedBy on the role
		if err := tx.Model(&models.Role{}).Where("id = ?", roleID).Update("last_modified_by", userID).Error; err != nil {
			return err
		}

		// Soft delete the role-permission assignment
		return tx.Delete(&rolePermission).Error
	})
}

func (s *permissionService) HasPermission(userID, permissionName string) (bool, error) {
	var count int64
	err := s.db.Table("user_profiles"). // Use correct table name
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

func (s *permissionService) GetUserRoleLevel(userID string) (int, error) {
	var role models.Role
	err := s.db.Table("user_profiles"). // Use correct table name
						Joins("JOIN roles ON user_profiles.role_id = roles.id").
						Where("user_profiles.id = ? AND roles.status = ? AND user_profiles.deleted_at IS NULL AND roles.deleted_at IS NULL", userID, models.RoleActive).
						Select("roles.level").First(&role).Error
	if err != nil {
		return 0, errors.New("user has no active role")
	}
	return role.Level, nil
}
