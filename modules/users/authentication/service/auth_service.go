package auth_service

import (
	"context"
	"errors"
	"strings"
	"time"

	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/authentication/dto"
	"ticket-zetu-api/modules/users/authentication/mail"
	"ticket-zetu-api/modules/users/models/members"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserService interface {
	SignUp(ctx context.Context, req dto.SignUpRequest, userID, encodedHash string) (*members.User, error)
	Authenticate(ctx context.Context, c *fiber.Ctx, usernameOrEmail, encodedHash string, rememberMe bool, ipAddress, userAgent string) (*members.User, *members.UserSession, error)
	ValidateUserExists(username, email string) error
	CreateSession(tx *gorm.DB, session *members.UserSession) error
	VerifyEmailCode(tx *gorm.DB, userID, code string) error
	UpdateVerificationCode(ctx context.Context, userID, verificationCode string) error
	RequestPasswordReset(ctx context.Context, c *fiber.Ctx, usernameOrEmail string) error
	SetNewPassword(ctx context.Context, c *fiber.Ctx, resetToken, newPassword string) error
}

type userService struct {
	db           *gorm.DB
	logHandler   *handler.LogHandler
	emailService mail_service.EmailService
}

func NewUserService(db *gorm.DB, logHandler *handler.LogHandler, emailService mail_service.EmailService) UserService {
	return &userService{
		db:           db,
		logHandler:   logHandler,
		emailService: emailService,
	}
}

func (s *userService) UpdateVerificationCode(ctx context.Context, userID, verificationCode string) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return errors.New("failed to start transaction")
	}

	if err := tx.Model(&members.UserSecurityAttributes{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"email_verification_token": verificationCode,
			"email_token_expiry":       time.Now().Add(24 * time.Hour),
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("failed to commit transaction")
	}

	return nil
}

func (s *userService) VerifyEmailCode(tx *gorm.DB, userID, code string) error {
	var securityAttrs members.UserSecurityAttributes
	if err := tx.Where("user_id = ? AND is_deleted = ?", userID, false).First(&securityAttrs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found or account is deleted")
		}
		return errors.New("failed to fetch user security attributes: " + err.Error())
	}

	if securityAttrs.EmailVerificationToken != code {
		return errors.New("invalid verification code")
	}

	if securityAttrs.EmailTokenExpiry != nil && securityAttrs.EmailTokenExpiry.Before(time.Now()) {
		return errors.New("verification code has expired")
	}

	verifiedAt := time.Now()
	if err := tx.Model(&members.UserSecurityAttributes{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"email_verified":           true,
			"email_verified_at":        verifiedAt,
			"email_verification_token": "",
			"email_token_expiry":       nil,
			"updated_at":               verifiedAt,
		}).Error; err != nil {
		return errors.New("failed to update email verification status: " + err.Error())
	}

	return nil
}

func (s *userService) ValidateUserExists(username, email string) error {
	var user members.User

	if err := s.db.Where("username = ?", username).First(&user).Error; err == nil {
		return errors.New("username already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("failed to check username existence")
	}

	if err := s.db.Where("email = ?", email).First(&user).Error; err == nil {
		return errors.New("email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("failed to check email existence")
	}

	return nil
}

func isDuplicateKeyError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Error 1062")
}

func getDuplicateKeyMessage(err error) error {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "idx_user_profiles_username"):
		return errors.New("username already exists")
	case strings.Contains(msg, "idx_user_profiles_email"):
		return errors.New("email already exists")
	case strings.Contains(msg, "idx_user_profiles_phone"):
		return errors.New("phone number already exists")
	default:
		return errors.New("duplicate entry")
	}
}

func isUserOver16(dob time.Time) bool {
	return time.Since(dob).Hours()/24/365 >= 16
}
