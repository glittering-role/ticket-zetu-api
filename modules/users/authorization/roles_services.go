package authorization

import (
	"errors"
	"ticket-zetu-api/modules/users/authorization/utils"
	models "ticket-zetu-api/modules/users/models/authorization"
	users "ticket-zetu-api/modules/users/models/members"

	"gorm.io/gorm"
)

type RoleService interface {
	CreateRole(role *models.Role) error
	GetRoles(filters map[string]interface{}, limit, offset int) ([]models.Role, error)
	GetRoleByID(id string) (*models.Role, error)
	GetRoleWithPermissions(id string) (*models.Role, []models.Permission, error)
	UpdateRole(id string, updates map[string]interface{}) error
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

func (s *roleService) CreateRole(role *models.Role) error {
	var existingRole models.Role
	if err := s.db.Unscoped().Where("role_name = ?", role.RoleName).First(&existingRole).Error; err == nil {
		if existingRole.DeletedAt.Valid {
			return errors.New("a role with this name was previously deleted")
		}
		return errors.New("role name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.db.Create(role).Error
}

func (s *roleService) GetRoles(filters map[string]interface{}, limit, offset int) ([]models.Role, error) {
	var roles []models.Role
	query := s.db.Model(&models.Role{}).Where("deleted_at IS NULL")
	for key, value := range filters {
		query = query.Where(key, value)
	}
	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&roles).Error
	return roles, err
}

func (s *roleService) GetRoleByID(id string) (*models.Role, error) {
	var role models.Role
	err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// New method to get role with its permissions
func (s *roleService) GetRoleWithPermissions(id string) (*models.Role, []models.Permission, error) {
	var role models.Role
	err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&role).Error
	if err != nil {
		return nil, nil, err
	}

	// Get related permissions through role_permissions table
	var permissions []models.Permission
	err = s.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND role_permissions.deleted_at IS NULL AND permissions.deleted_at IS NULL", id).
		Find(&permissions).Error

	return &role, permissions, err
}

func (s *roleService) UpdateRole(id string, updates map[string]interface{}) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var role models.Role
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&role).Error; err != nil {
			return err
		}
		if name, ok := updates["role_name"]; ok {
			var existingRole models.Role
			if err := tx.Unscoped().Where("role_name = ? AND id != ?", name, id).First(&existingRole).Error; err == nil {
				if existingRole.DeletedAt.Valid {
					return errors.New("a role with this name was previously deleted")
				}
				return errors.New("role name already exists")
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		if status, ok := updates["status"]; ok && (status == string(models.RoleInactive) || status == string(models.RoleArchived)) {
			var userCount int64
			if err := tx.Table("user_profiles").Where("role_id = ? AND deleted_at IS NULL", id).Count(&userCount).Error; err != nil {
				return err
			}
			if userCount > 0 {
				return errors.New("cannot change status of role with assigned users")
			}
		}
		return tx.Model(&models.Role{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).Error
	})
}

func (s *roleService) DeleteRole(id string, userID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var role models.Role
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&role).Error; err != nil {
			return err
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
		}
		if err := tx.Model(&models.Role{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Role{}, "id = ?", id).Error
	})
}

func (s *roleService) HasPermission(userID, permissionName string) (bool, error) {
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

func (s *roleService) GetUserRoleLevel(userID string) (int, error) {
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

func (s *roleService) AssignRoleToUser(userID, roleID, callerID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Check if the role exists and is active
		var role models.Role
		if err := tx.Where("id = ? AND status = ? AND deleted_at IS NULL", roleID, models.RoleActive).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("role not found or inactive")
			}
			return err
		}

		// Check if the user exists
		var user users.User // Use shared.User instead of users.User
		if err := tx.Table("user_profiles").Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			return err
		}

		// Get the caller's role level
		callerLevel, err := s.GetUserRoleLevel(callerID)
		if err != nil {
			return err
		}

		// Ensure caller can't assign a role with higher level
		if role.Level > callerLevel {
			return errors.New("cannot assign role with higher level than caller's")
		}

		// If user has an existing role, decrement its number_of_users
		if user.RoleID != "" && user.RoleID != roleID {
			var oldRole models.Role
			if err := tx.Where("id = ? AND deleted_at IS NULL", user.RoleID).First(&oldRole).Error; err == nil {
				if err := tx.Model(&models.Role{}).
					Where("id = ?", user.RoleID).
					Updates(map[string]interface{}{
						"number_of_users":  gorm.Expr("number_of_users - 1"),
						"last_modified_by": callerID,
					}).Error; err != nil {
					return err
				}
			}
		}

		// Update number_of_users for the new role using utils
		if err := utils.UpdateRoleUserCount(tx, roleID, 1); err != nil {
			return err
		}

		// Update user's role_id and last_modified_by
		if err := tx.Table("user_profiles").
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"role_id":          roleID,
				"last_modified_by": callerID,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}
