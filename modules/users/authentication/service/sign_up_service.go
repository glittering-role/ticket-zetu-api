package auth_service

import (
	"context"
	"errors"
	"time"

	"ticket-zetu-api/modules/users/authentication/dto"
	authorization "ticket-zetu-api/modules/users/models/authorization"
	"ticket-zetu-api/modules/users/models/members"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *userService) SignUp(ctx context.Context, req dto.SignUpRequest, userID, encodedHash string) (*members.User, error) {
	var dob *time.Time
	if req.DateOfBirth != "" {
		parsedDob, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			return nil, errors.New("invalid date_of_birth format, use YYYY-MM-DD")
		}
		if !isUserOver16(parsedDob) {
			return nil, errors.New("user must be at least 16 years old")
		}
		dob = &parsedDob
	}

	var role authorization.Role
	if err := s.db.Where("role_name = ?", "guest").First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("default guest role not found")
		}
		return nil, err
	}

	if err := s.ValidateUserExists(req.Username, req.Email); err != nil {
		return nil, err
	}

	user := &members.User{
		ID:          userID,
		Username:    req.Username,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Phone:       req.Phone,
		DateOfBirth: dob,
		RoleID:      role.ID,
	}

	prefs := &members.UserPreferences{
		UserID:   uuid.MustParse(user.ID),
		Language: "en",
		Theme:    "light",
		Timezone: "UTC",
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, errors.New("failed to start transaction")
	}

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		if isDuplicateKeyError(err) {
			return nil, getDuplicateKeyMessage(err)
		}
		return nil, err
	}

	expiry := time.Now().Add(24 * time.Hour)
	securityAttrs := members.UserSecurityAttributes{
		UserID:           uuid.MustParse(user.ID),
		Password:         encodedHash,
		AuthType:         members.AuthTypePassword,
		TwoFactorEnabled: false,
		EmailTokenExpiry: &expiry,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := tx.Create(&securityAttrs).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Create(prefs).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("failed to commit transaction")
	}

	return user, nil
}
