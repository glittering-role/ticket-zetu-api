package auth_service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"ticket-zetu-api/modules/users/models/members"

	"github.com/gofiber/fiber/v2"
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

type LogHandler interface {
	LogError(c *fiber.Ctx, err error, statusCode int) error
}

func (s *userService) Authenticate(ctx context.Context, c *fiber.Ctx, usernameOrEmail, password string, rememberMe bool, ipAddress, userAgent string) (*members.User, *members.UserSession, error) {
	user, securityAttrs, err := s.findUserAndSecurity(usernameOrEmail)
	if err != nil {
		return nil, nil, err
	}

	// Check if email is verified
	if !securityAttrs.EmailVerified {
		_, err := s.emailService.GenerateAndSendVerificationCode(c, user.Email, user.Username, user.ID)
		if err != nil {
			s.logHandler.LogError(c, fmt.Errorf("failed to resend verification email: %v", err), fiber.StatusInternalServerError)
		}
		return nil, nil, errors.New("email not verified, verification email resent")
	}

	// Check if account is locked
	if securityAttrs.IsLocked() {
		err := s.emailService.SendLoginWarning(c, user.Email, user.Username, userAgent, ipAddress, time.Now(), "account_locked")
		if err != nil {
			s.logHandler.LogError(c, fmt.Errorf("failed to send lockout warning email: %v", err), fiber.StatusInternalServerError)
		}
		return nil, nil, errors.New("account temporarily locked")
	}

	// Hash the input password using Argon2 with user ID as salt
	hashed := argon2.IDKey([]byte(password), []byte(user.ID), Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLength)
	encoded := base64.RawStdEncoding.EncodeToString(hashed)

	// Constant-time compare
	if subtle.ConstantTimeCompare([]byte(encoded), []byte(securityAttrs.Password)) != 1 {
		return nil, nil, s.handleFailedLogin(c, securityAttrs, user)
	}

	return s.handleSuccessfulLogin(c, user, securityAttrs, rememberMe, ipAddress, userAgent)
}

func (s *userService) HashPassword(password, userID string) (string, error) {
	hashed := argon2.IDKey([]byte(password), []byte(userID), Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLength)
	return base64.RawStdEncoding.EncodeToString(hashed), nil
}

func (s *userService) findUserAndSecurity(usernameOrEmail string) (*members.User, *members.UserSecurityAttributes, error) {
	type userWithSecurity struct {
		members.User
		UserSecurityAttributes members.UserSecurityAttributes `gorm:"embedded"`
	}

	var result userWithSecurity

	err := s.db.
		Table("user_profiles").
		Select("user_profiles.*, user_security_attributes.*").
		Joins("JOIN user_security_attributes ON user_security_attributes.user_id = user_profiles.id").
		Where("user_profiles.username = ? OR user_profiles.email = ?", usernameOrEmail, usernameOrEmail).
		First(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("invalid credentials")
		}
		return nil, nil, err
	}

	return &result.User, &result.UserSecurityAttributes, nil
}

func (s *userService) handleFailedLogin(c *fiber.Ctx, securityAttrs *members.UserSecurityAttributes, user *members.User) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return errors.New("failed to start transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	securityAttrs.IncrementFailedAttempts(5)
	if err := tx.Model(&members.UserSecurityAttributes{}).
		Where("user_id = ?", securityAttrs.UserID).
		Updates(map[string]interface{}{
			"failed_login_attempts": securityAttrs.FailedLoginAttempts,
			"lock_until":            securityAttrs.LockUntil,
			"updated_at":            time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("failed to commit transaction")
	}

	remainingAttempts := 5 - securityAttrs.FailedLoginAttempts
	if remainingAttempts <= 0 {
		err := s.emailService.SendLoginWarning(c, user.Email, user.Username, c.Get("User-Agent"), c.IP(), time.Now(), "lockout_failed_attempts")
		if err != nil {
			s.logHandler.LogError(c, fmt.Errorf("failed to send lockout warning email: %v", err), fiber.StatusInternalServerError)
		}
		return errors.New("account locked due to too many failed attempts")
	}

	return errors.New("invalid credentials")
}

func (s *userService) handleSuccessfulLogin(c *fiber.Ctx, user *members.User, securityAttrs *members.UserSecurityAttributes, rememberMe bool, ipAddress, userAgent string) (*members.User, *members.UserSession, error) {
	// Send login warning email for new login
	err := s.emailService.SendLoginWarning(c, user.Email, user.Username, userAgent, ipAddress, time.Now(), "new_login")
	if err != nil {
		s.logHandler.LogError(c, fmt.Errorf("failed to send login warning email: %v", err), fiber.StatusInternalServerError)
		// Continue login despite email failure
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, nil, errors.New("failed to start transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Reset failed login attempts
	if err := tx.Model(&members.UserSecurityAttributes{}).
		Where("user_id = ?", securityAttrs.UserID).
		Updates(map[string]interface{}{
			"failed_login_attempts": 0,
			"lock_until":            nil,
			"updated_at":            time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// Create session
	sessionDuration := time.Hour * 24 // Match SessionDuration
	if rememberMe {
		sessionDuration = time.Hour * 24 * 7 // Match LongSessionDuration
	}

	sessionToken, err := s.generateSecureToken(32)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	refreshToken, err := s.generateSecureToken(32)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	session := &members.UserSession{
		UserID:        uuid.MustParse(user.ID),
		SessionToken:  sessionToken,
		RefreshToken:  refreshToken,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		DeviceType:    s.detectDeviceType(userAgent),
		ExpiresAt:     time.Now().Add(sessionDuration),
		RefreshExpiry: time.Now().Add(sessionDuration * 2),
	}

	if err := s.CreateSession(tx, session); err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, nil, errors.New("failed to commit transaction")
	}

	return nil, session, nil
}

func (s *userService) generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

func (s *userService) detectDeviceType(userAgent string) members.SessionDeviceType {
	userAgent = strings.ToLower(userAgent)
	switch {
	case strings.Contains(userAgent, "mobile"):
		return members.DeviceMobile
	case strings.Contains(userAgent, "tablet"):
		return members.DeviceTablet
	case strings.Contains(userAgent, "windows") || strings.Contains(userAgent, "macintosh"):
		return members.DeviceDesktop
	default:
		return members.DeviceUnknown
	}
}

func (s *userService) CreateSession(tx *gorm.DB, session *members.UserSession) error {
	if session == nil {
		return errors.New("session data is required")
	}
	return tx.Create(session).Error
}
