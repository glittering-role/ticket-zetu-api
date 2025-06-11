package service

import (
	"errors"
	"fmt"
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
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var updatedUser members.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var user members.User
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).
			First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			return fmt.Errorf("failed to fetch user: %w", err)
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
				return fmt.Errorf("failed to fetch security attributes: %w", err)
			}
		}

		// Generate verification code
		verificationCode, err := s.emailService.GenerateAndSendVerificationCode(nil, emailDto.Email, user.Username, id)
		if err != nil {
			return fmt.Errorf("failed to generate and send verification code: %w", err)
		}

		token := verificationCode
		expiry := time.Now().Add(24 * time.Hour)
		securityUpdates := map[string]interface{}{
			"pending_email":            emailDto.Email,
			"email_verification_token": token,
			"email_token_expiry":       expiry,
			"updated_at":               time.Now(),
		}

		if security.ID == uuid.Nil {
			security.PendingEmail = emailDto.Email
			security.EmailVerificationToken = token
			security.EmailTokenExpiry = &expiry
			security.CreatedAt = time.Now()
			security.UpdatedAt = time.Now()
			if err := tx.Create(&security).Error; err != nil {
				return fmt.Errorf("failed to create security attributes: %w", err)
			}
		} else {
			if err := tx.Model(&security).Updates(securityUpdates).Error; err != nil {
				return fmt.Errorf("failed to update security attributes: %w", err)
			}
		}

		if err := tx.Preload("Preferences").Preload("Location").Preload("Role").
			Where("id = ?", id).First(&updatedUser).Error; err != nil {
			return fmt.Errorf("failed to fetch updated user: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return toUserProfileResponseDto(&updatedUser, true, nil), nil
}

func (s *userService) UpdateUsername(id string, usernameDto *dto.UpdateUsernameDto, updaterID string) (*dto.UserProfileResponseDto, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid user ID format")
	}

	if err := s.validator.Struct(usernameDto); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate username and check availability
	available, suggestions, err := s.usernameCheck.ValidateAndCheckUsername(usernameDto.Username)
	if err != nil {
		return nil, fmt.Errorf("username validation failed: %w", err)
	}
	if !available {
		return nil, fmt.Errorf("username already taken, suggestions: %v", suggestions)
	}

	var updatedUser members.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var user members.User
		if err := tx.Where("id = ? AND deleted_at IS NULL", id).
			First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			return fmt.Errorf("failed to fetch user: %w", err)
		}

		updates := map[string]interface{}{
			"username":         usernameDto.Username,
			"last_modified_by": updaterID,
			"updated_at":       time.Now(),
		}

		if err := tx.Model(&user).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update username: %w", err)
		}

		if err := tx.Preload("Preferences").Preload("Location").Preload("Role").
			Where("id = ?", id).First(&updatedUser).Error; err != nil {
			return fmt.Errorf("failed to fetch updated user: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return toUserProfileResponseDto(&updatedUser, true, nil), nil
}
