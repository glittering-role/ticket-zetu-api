package authentication

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"time"

	"ticket-zetu-api/modules/users/models/members"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" validate:"required"`
	Password        string `json:"password" validate:"required,min=8"`
	RememberMe      bool   `json:"remember_me"`
}

func (c *AuthController) SignIn(ctx *fiber.Ctx) error {
	var req LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return c.logHandler.LogError(ctx, errors.New("invalid request payload"), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(req); err != nil {
		return c.logHandler.LogError(ctx, errors.New("validation failed: "+err.Error()), fiber.StatusBadRequest)
	}

	user, securityAttrs, err := c.findUserAndSecurity(req.UsernameOrEmail)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	if securityAttrs.IsLocked() {
		return c.logHandler.LogError(ctx, errors.New("account temporarily locked"), fiber.StatusUnauthorized)
	}

	// Hash the input password using same params and user ID as salt
	hashed := argon2.IDKey([]byte(req.Password), []byte(user.ID), Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLength)
	encoded := base64.RawStdEncoding.EncodeToString(hashed)

	// Constant-time compare
	if subtle.ConstantTimeCompare([]byte(encoded), []byte(securityAttrs.Password)) != 1 {
		return c.handleFailedLogin(ctx, securityAttrs, user)
	}

	return c.handleSuccessfulLogin(ctx, user, securityAttrs, req.RememberMe)
}

func (c *AuthController) findUserAndSecurity(usernameOrEmail string) (*members.User, *members.UserSecurityAttributes, error) {
	type userWithSecurity struct {
		members.User
		UserSecurityAttributes members.UserSecurityAttributes `gorm:"embedded"`
	}

	var result userWithSecurity

	err := c.db.
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

func (c *AuthController) handleFailedLogin(ctx *fiber.Ctx, securityAttrs *members.UserSecurityAttributes, user *members.User) error {
	tx := c.db.Begin()
	if tx.Error != nil {
		return c.logHandler.LogError(ctx, errors.New("failed to start transaction"), fiber.StatusInternalServerError)
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
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	if err := tx.Commit().Error; err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	remainingAttempts := 5 - securityAttrs.FailedLoginAttempts
	if remainingAttempts > 0 {
		return c.logHandler.LogError(ctx, errors.New("invalid credentials"), fiber.StatusUnauthorized)
	}

	return c.logHandler.LogError(ctx, errors.New("account locked due to too many failed attempts"), fiber.StatusUnauthorized)
}

func (c *AuthController) handleSuccessfulLogin(ctx *fiber.Ctx, user *members.User, securityAttrs *members.UserSecurityAttributes, rememberMe bool) error {
	if err := c.db.Model(&members.UserSecurityAttributes{}).
		Where("user_id = ?", securityAttrs.UserID).
		Updates(map[string]interface{}{
			"failed_login_attempts": 0,
			"lock_until":            nil,
			"updated_at":            time.Now(),
		}).Error; err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	sessionDuration := SessionDuration
	if rememberMe {
		sessionDuration = LongSessionDuration
	}

	sessionToken, err := generateSecureToken(TokenLength)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}
	refreshToken, err := generateSecureToken(TokenLength)
	if err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	session := members.UserSession{
		UserID:        uuid.MustParse(user.ID),
		SessionToken:  sessionToken,
		RefreshToken:  refreshToken,
		IPAddress:     ctx.IP(),
		UserAgent:     ctx.Get("User-Agent"),
		DeviceType:    detectDeviceType(ctx.Get("User-Agent")),
		ExpiresAt:     time.Now().Add(sessionDuration),
		RefreshExpiry: time.Now().Add(sessionDuration * 2),
	}

	tx := c.db.Begin()
	if tx.Error != nil {
		return c.logHandler.LogError(ctx, errors.New("failed to start transaction"), fiber.StatusInternalServerError)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := c.userService.CreateSession(tx, &session); err != nil {
		tx.Rollback()
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	if err := tx.Commit().Error; err != nil {
		return c.logHandler.LogError(ctx, errors.New("failed to commit transaction"), fiber.StatusInternalServerError)
	}

	isProd := true // Update based on env
	ctx.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    session.SessionToken,
		Expires:  session.ExpiresAt,
		HTTPOnly: true,
		Secure:   isProd,
		SameSite: "Strict",
	})
	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    session.RefreshToken,
		Expires:  session.RefreshExpiry,
		HTTPOnly: true,
		Secure:   isProd,
		SameSite: "Strict",
	})

	return c.logHandler.LogSuccess(ctx, nil, "Login successful", true)
}
