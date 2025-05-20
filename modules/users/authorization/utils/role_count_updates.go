package utils

import (
	"errors"

	"gorm.io/gorm"
	"ticket-zetu-api/modules/users/models/authorization"
)

// UpdateRoleUserCount updates the NumberOfUsers field for a role.
func UpdateRoleUserCount(tx *gorm.DB, roleID string, delta int) error {
	if roleID == "" {
		return errors.New("role ID cannot be empty")
	}
	result := tx.Model(&model.Role{}).
		Where("id = ?", roleID).
		Update("number_of_users", gorm.Expr("number_of_users + ?", delta))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("role not found")
	}
	return nil
}
