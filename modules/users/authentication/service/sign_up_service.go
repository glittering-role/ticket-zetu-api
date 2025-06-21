package auth_service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ticket-zetu-api/modules/users/authentication/dto"
	"ticket-zetu-api/modules/users/helpers"
	authorization "ticket-zetu-api/modules/users/models/authorization"
	"ticket-zetu-api/modules/users/models/members"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UsernameExistsError is a custom error type for username conflicts
type UsernameExistsError struct {
	Message string
}

func (e *UsernameExistsError) Error() string {
	return e.Message
}

// EmailExistsError is a custom error type for email conflicts
type EmailExistsError struct {
	Message string
}

func (e *EmailExistsError) Error() string {
	return e.Message
}

// UserProfileResponse holds user profile data for API responses
type UserProfileResponse struct {
	ID          string                `json:"id"`
	Username    string                `json:"username"`
	FirstName   string                `json:"first_name,omitempty"`
	LastName    string                `json:"last_name,omitempty"`
	Email       string                `json:"email"`
	Phone       string                `json:"phone,omitempty"`
	DateOfBirth *time.Time            `json:"date_of_birth,omitempty"`
	Location    *UserLocationEditable `json:"location,omitempty"`
}

// UserLocationEditable holds editable location fields for users
type UserLocationEditable struct {
	City  string `json:"city,omitempty"`
	State string `json:"state,omitempty"`
	Zip   string `json:"zip,omitempty"`
}

func normalizeEmail(e string) string {
	return strings.ToLower(strings.TrimSpace(e))
}

func (s *userService) ValidateUserExists(username, email string) error {
	normalizedEmail := normalizeEmail(email)

	// 1) Username
	if err := s.db.Where("username = ?", username).First(&members.User{}).Error; err == nil {
		return &UsernameExistsError{"username already exists"}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("username check failed: %w", err)
	}

	// 2) Confirmed email
	if err := s.db.
		Where("LOWER(email) = ?", normalizedEmail).
		First(&members.User{}).
		Error; err == nil {
		return &EmailExistsError{"email already exists"}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("email check failed: %w", err)
	}

	// 3) Pending email
	if err := s.db.
		Where("LOWER(pending_email) = ?", normalizedEmail).
		First(&members.UserSecurityAttributes{}).
		Error; err == nil {
		return &EmailExistsError{"email is already pending verification"}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("pending email check failed: %w", err)
	}

	return nil
}

// isUserOver16 checks if the user is at least 16 years old
func isUserOver16(dob time.Time) bool {
	return time.Since(dob).Hours()/24/365 >= 16
}

// SignUp handles user registration
func (s *userService) SignUp(ctx context.Context, req dto.SignUpRequest, userID, encodedHash string) (*members.User, error) {
	// Parse and validate date of birth
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

	// Validate username and email availability
	if err := s.ValidateUserExists(req.Username, req.Email); err != nil {
		return nil, err
	}

	// Get default guest role
	var role authorization.Role
	if err := s.db.Where("role_name = ?", "guest").First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("default guest role not found")
		}
		return nil, fmt.Errorf("failed to get guest role: %w", err)
	}

	// Prepare user object
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

	// Get location from context and map to UserLocation
	var userLocation *members.UserLocation
	if loc, ok := ctx.Value("user_location").(*helpers.Location); ok && loc != nil {
		userLocation = &members.UserLocation{
			UserID: uuid.MustParse(userID),
			//City:      loc.City,
			//State:     loc.State,
			Country:   loc.Country,
			Continent: loc.Continent,
			Timezone:  loc.Timezone,
			//Zip:       loc.Zip,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// Prepare user preferences
	prefs := &members.UserPreferences{
		UserID:   uuid.MustParse(user.ID),
		Language: "en",
		Theme:    "light",
		Timezone: "UTC",
	}

	// Prepare security attributes
	expiry := time.Now().Add(24 * time.Hour)
	securityAttrs := &members.UserSecurityAttributes{
		UserID:           uuid.MustParse(user.ID),
		Password:         encodedHash,
		AuthType:         members.AuthTypePassword,
		TwoFactorEnabled: false,
		EmailTokenExpiry: &expiry,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Start database transaction
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, errors.New("failed to start transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create user and related records
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		if isDuplicateKeyError(err) {
			return nil, getDuplicateKeyMessage(err)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := tx.Create(securityAttrs).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create security attributes: %w", err)
	}

	if err := tx.Create(prefs).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user preferences: %w", err)
	}

	// Create user location if available
	if userLocation != nil {
		if err := tx.Create(userLocation).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create user location: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("failed to commit transaction")
	}

	return user, nil
}

// GetUserProfile retrieves user profile with editable location fields
func (s *userService) GetUserProfile(ctx context.Context, userID string) (*UserProfileResponse, error) {
	var user members.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	var userLocation members.UserLocation
	var location *UserLocationEditable
	if err := s.db.Where("user_id = ?", userID).First(&userLocation).Error; err == nil {
		// Only expose editable fields
		location = &UserLocationEditable{
			City:  userLocation.City,
			State: userLocation.State,
			Zip:   userLocation.Zip,
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to fetch user location: %w", err)
	}

	return &UserProfileResponse{
		ID:          user.ID,
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		Phone:       user.Phone,
		DateOfBirth: user.DateOfBirth,
		Location:    location,
	}, nil
}
