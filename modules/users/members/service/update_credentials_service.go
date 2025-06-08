package service

import (
	"errors"
	"ticket-zetu-api/modules/users/members/dto"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *userService) UpdatePhone(id string, phoneDto *dto.UpdatePhoneDto, updaterID string) (*dto.UserProfileResponseDto, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid user ID format")
	}

	if err := s.validator.Struct(phoneDto); err != nil {
		return nil, errors.New("validation failed: " + err.Error())
	}

	var updatedUser members.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var user members.User
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).
			First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			return err
		}

		if err := s.checkUniqueField(tx, "phone", phoneDto.Phone, id); err != nil {
			return err
		}

		updates := map[string]interface{}{
			"phone":            phoneDto.Phone,
			"last_modified_by": updaterID,
			"updated_at":       time.Now(),
		}

		if err := tx.Model(&user).Updates(updates).Error; err != nil {
			return err
		}

		if err := tx.Preload("Preferences").Preload("Location").Preload("Role").
			Where("id = ?", id).First(&updatedUser).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return toUserProfileResponseDto(&updatedUser, true, nil), nil
}

func (s *userService) UpdateUserEmail(id string, emailDto *dto.UpdateEmailDto, updaterID string) (*dto.UserProfileResponseDto, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid user ID format")
	}

	if err := s.validator.Struct(emailDto); err != nil {
		return nil, errors.New("validation failed: " + err.Error())
	}

	var updatedUser members.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var user members.User
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).
			First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			return err
		}

		if err := s.checkUniqueField(tx, "email", emailDto.Email, id); err != nil {
			return err
		}

		var security members.UserSecurityAttributes
		if err := tx.Where("user_id = ?", id).First(&security).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				security = members.UserSecurityAttributes{
					ID:     uuid.New(),
					UserID: uuid.MustParse(id),
				}
			} else {
				return err
			}
		}

		token := uuid.New().String()
		expiry := time.Now().Add(24 * time.Hour)
		securityUpdates := map[string]interface{}{
			"pending_email":            emailDto.Email,
			"email_verification_token": token,
			"email_token_expiry":       expiry,
		}

		if security.ID == uuid.Nil {
			security.PendingEmail = emailDto.Email
			security.EmailVerificationToken = token
			security.EmailTokenExpiry = &expiry
			if err := tx.Create(&security).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&security).Updates(securityUpdates).Error; err != nil {
				return err
			}
		}

		// Stub for sending verification email
		// if err := sendVerificationEmail(emailDto.Email, token); err != nil {
		// 	return errors.New("failed to send verification email")
		// }

		if err := tx.Preload("Preferences").Preload("Location").Preload("Role").
			Where("id = ?", id).First(&updatedUser).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return toUserProfileResponseDto(&updatedUser, true, nil), nil
}
