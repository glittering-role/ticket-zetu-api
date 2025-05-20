package authentication

import (
	"errors"
	"strings"
	"time"

	"ticket-zetu-api/modules/users/models/members"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService interface {
	CreateUser(tx *gorm.DB, user *members.User, encodedHashedPassword string, prefs *members.UserPreferences, session *members.UserSession) error
	ValidateUserExists(username, email string) error
	CreateSession(tx *gorm.DB, session *members.UserSession) error
}

type userService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) UserService {
	return &userService{db: db}
}

func (s *userService) CreateUser(
	tx *gorm.DB,
	user *members.User,
	encodedHashedPassword string,
	prefs *members.UserPreferences,
	session *members.UserSession,
) error {
	// Attempt to create the user
	if err := tx.Create(user).Error; err != nil {
		if isDuplicateKeyError(err) {
			return getDuplicateKeyMessage(err)
		}
		return err
	}

	// Create related security attributes
	securityAttrs := members.UserSecurityAttributes{
		UserID:           uuid.MustParse(user.ID),
		Password:         encodedHashedPassword,
		AuthType:         members.AuthTypePassword,
		TwoFactorEnabled: false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := tx.Create(&securityAttrs).Error; err != nil {
		return err
	}

	// Create user preferences (optional)
	if prefs != nil {
		if err := tx.Create(prefs).Error; err != nil {
			return err
		}
	}

	// Create session (optional)
	if session != nil {
		if err := tx.Create(session).Error; err != nil {
			return err
		}
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

func (s *userService) CreateSession(tx *gorm.DB, session *members.UserSession) error {
	if session == nil {
		return errors.New("session data is required")
	}
	return tx.Create(session).Error
}

// isDuplicateKeyError checks for MySQL duplicate entry error 1062.
func isDuplicateKeyError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Error 1062")
}

// getDuplicateKeyMessage extracts a user-friendly message based on the constraint.
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
