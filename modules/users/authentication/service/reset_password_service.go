package auth_service

import (
	"context"
	"crypto/subtle"
	"errors"
	"time"

	"ticket-zetu-api/modules/users/models/members"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func (s *userService) RequestPasswordReset(ctx context.Context, c *fiber.Ctx, usernameOrEmail string) error {
	user, securityAttrs, err := s.findUserAndSecurity(usernameOrEmail)
	if err != nil {
		return err
	}

	// Check if account is locked
	if securityAttrs.IsLocked() {
		return errors.New("account temporarily locked")
	}

	// Generate reset token
	resetToken, err := s.generateSecureToken(32)
	if err != nil {
		return errors.New("failed to generate reset token")
	}

	// Set token and expiry (24 hours)
	resetTokenExpiry := time.Now().Add(24 * time.Hour)
	tx := s.db.Begin()
	if tx.Error != nil {
		return errors.New("failed to start transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&members.UserSecurityAttributes{}).
		Where("user_id = ?", securityAttrs.UserID).
		Updates(map[string]interface{}{
			"password_reset_token":        resetToken,
			"password_reset_token_expiry": resetTokenExpiry,
			"updated_at":                  time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("failed to commit transaction")
	}

	// Send password reset email
	err = s.emailService.SendPasswordResetEmail(c, user.Email, user.Username, resetToken)
	if err != nil {
		s.logHandler.LogError(c, errors.New("failed to send password reset email"), fiber.StatusInternalServerError)
	}

	return nil
}

func (s *userService) SetNewPassword(ctx context.Context, c *fiber.Ctx, resetToken, newPassword string) error {
	var securityAttrs members.UserSecurityAttributes
	err := s.db.
		Where("password_reset_token = ?", resetToken).
		Where("password_reset_token_expiry > ?", time.Now()).
		First(&securityAttrs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid or expired reset token")
		}
		return err
	}

	// Check if account is locked
	if securityAttrs.IsLocked() {
		return errors.New("account temporarily locked")
	}

	// Hash new password with user ID as salt
	hashedPassword, err := s.HashPassword(newPassword, securityAttrs.UserID.String())
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Prevent reusing the same password
	if subtle.ConstantTimeCompare([]byte(hashedPassword), []byte(securityAttrs.Password)) == 1 {
		return errors.New("new password must be different from the previous password")
	}

	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return errors.New("failed to start transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&members.UserSecurityAttributes{}).
		Where("user_id = ?", securityAttrs.UserID).
		Updates(map[string]interface{}{
			"password":                    hashedPassword,
			"password_reset_token":        nil,
			"password_reset_token_expiry": nil,
			"failed_login_attempts":       0,
			"lock_until":                  nil,
			"updated_at":                  time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("failed to commit transaction")
	}

	return nil
}
