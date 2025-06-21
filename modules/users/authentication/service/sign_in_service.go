package auth_service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"ticket-zetu-api/modules/users/helpers"
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

// UserSessionData matches middleware's cached user data
type UserSessionData struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type LogHandler interface {
	LogError(c *fiber.Ctx, err error, statusCode int) error
}

type EmailService interface {
	GenerateAndSendVerificationCode(c *fiber.Ctx, email, username, userID string) (string, error)
	SendLoginWarning(c *fiber.Ctx, email, username, deviceDescription, ipAddress, country, state string, loginTime time.Time, warningType string) error
}

func (s *userService) Authenticate(ctx context.Context, c *fiber.Ctx, usernameOrEmail, password string, rememberMe bool) (*members.User, *members.UserSession, error) {
	// Normalize input for consistent querying
	usernameOrEmail = strings.TrimSpace(strings.ToLower(usernameOrEmail))

	// Get IP and location from GeolocationMiddleware
	var location *helpers.Location
	if loc, ok := c.Locals("user_location").(*helpers.Location); ok && loc != nil {
		location = loc
	} else {
		return nil, nil, errors.New("location data not available")
	}

	// Get device info from DeviceDetectionMiddleware
	var deviceInfo *helpers.DeviceInfo
	if dev, ok := c.Locals("device_info").(*helpers.DeviceInfo); ok && dev != nil {
		deviceInfo = dev
	} else {
		return nil, nil, errors.New("device info not available")
	}

	user, securityAttrs, err := s.findUserAndSecurity(ctx, usernameOrEmail)
	if err != nil {
		return nil, nil, err
	}

	// Check if email is verified
	if !securityAttrs.EmailVerified {
		_, err := s.emailService.GenerateAndSendVerificationCode(c, user.Email, user.Username, user.ID)
		if err != nil {
			s.logHandler.LogError(c, errors.New("failed to resend verification email"), fiber.StatusInternalServerError)
		}
		return nil, nil, errors.New("email not verified, verification email resent")
	}

	// Check if account is locked
	if securityAttrs.IsLocked() {
		err := s.emailService.SendLoginWarning(c, user.Email, user.Username, deviceInfo.Description, location.IPAddress, location.Country, location.State, time.Now(), "account_locked")
		if err != nil {
			s.logHandler.LogError(c, errors.New("failed to send lockout warning email"), fiber.StatusInternalServerError)
		}
		return nil, nil, errors.New("account temporarily locked")
	}

	// Hash the input password using Argon2 with user ID as salt
	hashed := argon2.IDKey([]byte(password), []byte(user.ID), Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLength)
	encoded := base64.RawStdEncoding.EncodeToString(hashed)

	// Constant-time compare
	if subtle.ConstantTimeCompare([]byte(encoded), []byte(securityAttrs.Password)) != 1 {
		return nil, nil, s.handleFailedLogin(ctx, c, securityAttrs, user, location, deviceInfo)
	}

	return s.handleSuccessfulLogin(ctx, c, user, securityAttrs, rememberMe, location, deviceInfo)
}

func (s *userService) HashPassword(password, userID string) (string, error) {
	hashed := argon2.IDKey([]byte(password), []byte(userID), Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLength)
	return base64.RawStdEncoding.EncodeToString(hashed), nil
}

func (s *userService) findUserAndSecurity(ctx context.Context, usernameOrEmail string) (*members.User, *members.UserSecurityAttributes, error) {
	type userWithSecurity struct {
		members.User
		UserSecurityAttributes members.UserSecurityAttributes `gorm:"embedded"`
	}

	var result userWithSecurity

	// Ensure index on user_profiles(username, email) and user_security_attributes(user_id)
	err := s.db.WithContext(ctx).
		Table("user_profiles").
		Select("user_profiles.*, user_security_attributes.*").
		Joins("JOIN user_security_attributes ON user_security_attributes.user_id = user_profiles.id").
		Where("LOWER(user_profiles.username) = ? OR LOWER(user_profiles.email) = ?", usernameOrEmail, usernameOrEmail).
		First(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("invalid credentials")
		}
		return nil, nil, err
	}

	return &result.User, &result.UserSecurityAttributes, nil
}

func (s *userService) handleFailedLogin(ctx context.Context, c *fiber.Ctx, securityAttrs *members.UserSecurityAttributes, user *members.User, location *helpers.Location, deviceInfo *helpers.DeviceInfo) error {
	// Single update to reduce locking
	securityAttrs.IncrementFailedAttempts(5)
	err := s.db.WithContext(ctx).Model(&members.UserSecurityAttributes{}).
		Where("user_id = ?", securityAttrs.UserID).
		Updates(map[string]interface{}{
			"failed_login_attempts": securityAttrs.FailedLoginAttempts,
			"lock_until":            securityAttrs.LockUntil,
			"updated_at":            time.Now(),
		}).Error
	if err != nil {
		return err
	}

	remainingAttempts := 5 - securityAttrs.FailedLoginAttempts
	if remainingAttempts <= 0 {
		err := s.emailService.SendLoginWarning(c, user.Email, user.Username, deviceInfo.Description, location.IPAddress, location.Country, location.State, time.Now(), "lockout_failed_attempts")
		if err != nil {
			s.logHandler.LogError(c, errors.New("failed to send lockout warning email"), fiber.StatusInternalServerError)
		}
		return errors.New("account locked due to too many failed attempts")
	}

	return errors.New("invalid credentials")
}

func (s *userService) handleSuccessfulLogin(ctx context.Context, c *fiber.Ctx, user *members.User, securityAttrs *members.UserSecurityAttributes, rememberMe bool, location *helpers.Location, deviceInfo *helpers.DeviceInfo) (*members.User, *members.UserSession, error) {
	// Send login warning email for new login
	err := s.emailService.SendLoginWarning(c, user.Email, user.Username, deviceInfo.Description, location.IPAddress, location.Country, location.State, time.Now(), "new_login")
	if err != nil {
		s.logHandler.LogError(c, errors.New("failed to send login warning email"), fiber.StatusInternalServerError)
	}

	// Reset failed login attempts and store IP
	err = s.db.WithContext(ctx).Model(&members.UserSecurityAttributes{}).
		Where("user_id = ?", securityAttrs.UserID).
		Updates(map[string]interface{}{
			"failed_login_attempts": 0,
			"lock_until":            nil,
			"updated_at":            time.Now(),
		}).Error
	if err != nil {
		return nil, nil, err
	}

	// Create session
	sessionDuration := time.Hour * 24
	if rememberMe {
		sessionDuration = time.Hour * 24 * 7
	}

	sessionToken, err := s.generateSecureToken(32)
	if err != nil {
		return nil, nil, err
	}
	refreshToken, err := s.generateSecureToken(32)
	if err != nil {
		return nil, nil, err
	}

	session := &members.UserSession{
		UserID:        uuid.MustParse(user.ID),
		SessionToken:  sessionToken,
		RefreshToken:  refreshToken,
		IPAddress:     location.IPAddress,
		UserAgent:     c.Get("User-Agent"),
		DeviceType:    deviceInfo.DeviceType,
		DeviceName:    deviceInfo.Description,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(sessionDuration),
		RefreshExpiry: time.Now().Add(sessionDuration * 2),
	}

	// Store session in database
	err = s.CreateSession(s.db.WithContext(ctx), session)
	if err != nil {
		return nil, nil, err
	}

	// Cache session in Redis (same format as middleware)
	cacheKey := fmt.Sprintf("session:%s", session.SessionToken)
	userData := UserSessionData{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
	userDataJSON, err := json.Marshal(userData)
	if err != nil {
		s.logHandler.LogError(c, err, fiber.StatusInternalServerError)
	} else {
		err = s.redisClient.Set(ctx, cacheKey, userDataJSON, sessionDuration).Err()
		if err != nil {
			s.logHandler.LogError(c, err, fiber.StatusInternalServerError)
		}
	}

	return user, session, nil
}

func (s *userService) generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

func (s *userService) CreateSession(tx *gorm.DB, session *members.UserSession) error {
	if session == nil {
		return errors.New("session data is required")
	}
	return tx.Create(session).Error
}
