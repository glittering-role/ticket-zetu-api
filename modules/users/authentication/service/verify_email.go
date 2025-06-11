package auth_service

import (
	"errors"
	"strings"
	"time"

	"ticket-zetu-api/modules/users/models/members"

	"gorm.io/gorm"
)

func (s *userService) VerifyEmailCode(tx *gorm.DB, userID, code string) error {
	var sa members.UserSecurityAttributes
	if err := tx.
		Where("user_id = ? AND is_deleted = ?", userID, false).
		First(&sa).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found or account is deleted")
		}
		return errors.New("failed to fetch user security attributes")
	}

	now := time.Now()
	if err := tx.Model(&members.UserSecurityAttributes{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"email_verified":           true,
			"email_verified_at":        now,
			"email_verification_token": "",
			"email_token_expiry":       nil,
			"pending_email":            "",
			"updated_at":               now,
		}).Error; err != nil {
		return errors.New("failed to update email verification status")
	}

	var u members.User
	if err := tx.Select("email").
		Where("id = ?", userID).
		First(&u).Error; err != nil {
		return errors.New("failed to fetch current user email")
	}

	if strings.TrimSpace(u.Email) == "" {
		if err := tx.Model(&members.User{}).
			Where("id = ?", userID).
			Update("email", sa.PendingEmail).Error; err != nil {
			return errors.New("failed to update user email")
		}
	}

	return nil
}
