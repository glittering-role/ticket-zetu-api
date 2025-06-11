package service

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"regexp"
	"ticket-zetu-api/modules/users/members/dto"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

const (
	Argon2Time      = 1
	Argon2Memory    = 64 * 1024 // 64 MB
	Argon2Threads   = 4
	Argon2KeyLength = 32
)

func (s *userService) HashPassword(password, userID string) (string, error) {
	hashed := argon2.IDKey([]byte(password), []byte(userID), Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLength)
	return base64.RawStdEncoding.EncodeToString(hashed), nil
}

func (s *userService) SetNewPassword(userID string, userDto *dto.NewPasswordDto, updaterID string) (*dto.UserProfileResponseDto, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, errors.New("invalid user ID format")
	}

	if err := s.validator.Struct(userDto); err != nil {
		return nil, errors.New("validation failed: " + err.Error())
	}

	// Additional password strength check
	if !regexp.MustCompile(`[0-9]`).MatchString(userDto.NewPassword) || !regexp.MustCompile(`[a-zA-Z]`).MatchString(userDto.NewPassword) {
		return nil, errors.New("password must contain at least one number and one letter")
	}

	var updatedUser members.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Fetch user security attributes
		var securityAttrs members.UserSecurityAttributes
		if err := tx.Where("user_id = ?", userID).First(&securityAttrs).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user security attributes not found")
			}
			return err
		}

		// Hash new password
		hashedPassword, err := s.HashPassword(userDto.NewPassword, userID)
		if err != nil {
			return err
		}

		// Prevent reusing the same password
		if subtle.ConstantTimeCompare([]byte(hashedPassword), []byte(securityAttrs.Password)) == 1 {
			return errors.New("new password must be different from the previous password")
		}

		// Update security attributes
		updates := map[string]interface{}{
			"password":                    hashedPassword,
			"password_reset_token":        nil,
			"password_reset_token_expiry": nil,
			"failed_login_attempts":       0,
			"lock_until":                  nil,
			"last_modified_by":            updaterID,
			"updated_at":                  time.Now(),
		}

		if err := tx.Model(&members.UserSecurityAttributes{}).
			Where("user_id = ?", userID).
			Updates(updates).Error; err != nil {
			return err
		}

		// Fetch updated user with preloaded data
		if err := tx.Preload("Preferences").Preload("Location").Preload("Role").
			Where("id = ? AND deleted_at IS NULL", userID).
			First(&updatedUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return toUserProfileResponseDto(&updatedUser, true, nil), nil
}
