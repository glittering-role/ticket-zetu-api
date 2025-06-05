package service

import (
	"errors"
	"ticket-zetu-api/modules/users/members/dto"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService interface {
	GetUserProfile(identifier string, requesterID string) (*dto.UserProfileResponseDto, error)
	UpdateUserDetails(id string, userDto *dto.UpdateUserDto, updaterID string) (*dto.UserProfileResponseDto, error)
	UpdateUserLocation(id string, locationDto *dto.UserLocationUpdateDto, updaterID string) (*dto.UserProfileResponseDto, error)
	UpdatePhone(id string, phoneDto *dto.UpdatePhoneDto, updaterID string) (*dto.UserProfileResponseDto, error)
	UpdateUserEmail(id string, emailDto *dto.UpdateEmailDto, updaterID string) (*dto.UserProfileResponseDto, error)
}

type userService struct {
	db        *gorm.DB
	validator *validator.Validate
}

// NewUserService initializes the user service
func NewUserService(db *gorm.DB) UserService {
	return &userService{
		db:        db,
		validator: validator.New(),
	}
}

// toUserProfileResponseDto converts a members.User to dto.UserProfileResponseDto
func toUserProfileResponseDto(user *members.User, isOwner bool) *dto.UserProfileResponseDto {
	response := &dto.UserProfileResponseDto{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	if isOwner {
		response.Email = user.Email
		response.Phone = user.Phone
		if user.DateOfBirth != nil {
			dob := user.DateOfBirth.Format("2006-01-02")
			response.DateOfBirth = &dob
		}
		response.Gender = user.Gender
		if user.Role.ID != "" {
			response.RoleName = user.Role.RoleName
		}
		if user.Location.ID != uuid.Nil {
			response.Location = user.Location.FullLocation()
			response.LocationDetail = &dto.UserLocationDto{
				Country:   user.Location.Country,
				State:     user.Location.State,
				StateName: user.Location.StateName,
				Continent: user.Location.Continent,
				City:      user.Location.City,
				Zip:       user.Location.Zip,
				Timezone:  user.Location.Timezone,
			}
			if user.Location.LastActive != nil {
				lastActive := user.Location.LastActive.Format(time.RFC3339)
				response.LocationDetail.LastActive = &lastActive
			}
		}
		return response
	}

	// Non-owner: Only show fields explicitly allowed by UserPreferences
	if user.Preferences.ID == uuid.Nil || !user.Preferences.ShowProfile {
		return response
	}

	if user.Preferences.ShouldShow("email") {
		response.Email = user.Email
	}
	if user.Preferences.ShouldShow("phone") {
		response.Phone = user.Phone
	}
	if user.Preferences.ShouldShow("location") && user.Location.ID != uuid.Nil {
		response.Location = user.Location.FullLocation()
		response.LocationDetail = &dto.UserLocationDto{
			Country:   user.Location.Country,
			State:     user.Location.State,
			StateName: user.Location.StateName,
			Continent: user.Location.Continent,
			City:      user.Location.City,
			Zip:       user.Location.Zip,
			Timezone:  user.Location.Timezone,
		}
		if user.Location.LastActive != nil {
			lastActive := user.Location.LastActive.Format(time.RFC3339)
			response.LocationDetail.LastActive = &lastActive
		}
	}
	if user.Preferences.ShouldShow("gender") {
		response.Gender = user.Gender
	}
	if user.Preferences.ShouldShow("role") && user.Role.ID != "" {
		response.RoleName = user.Role.RoleName
	}

	return response
}

// GetUserProfile retrieves a user's profile based on identifier
func (s *userService) GetUserProfile(identifier string, requesterID string) (*dto.UserProfileResponseDto, error) {
	if identifier == "" {
		return nil, errors.New("identifier is required")
	}

	if err := s.validateIdentifier(identifier); err != nil {
		return nil, errors.New("invalid identifier format: " + err.Error())
	}

	var user members.User
	query := s.db.
		Preload("Preferences").
		Preload("Location").
		Preload("Role").
		Where("deleted_at IS NULL")

	if _, err := uuid.Parse(identifier); err == nil {
		query = query.Where("id = ?", identifier)
	} else {
		query = query.Where("username = ? OR email = ?", identifier, identifier)
	}

	if err := query.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	isOwner := user.ID == requesterID
	return toUserProfileResponseDto(&user, isOwner), nil
}

// checkUniqueField checks if a value for a given field is unique among users (excluding the given user ID)
func (s *userService) checkUniqueField(tx *gorm.DB, field string, value string, excludeID string) error {
	var count int64
	query := tx.Model(&members.User{}).Where(field+" = ? AND id != ? AND deleted_at IS NULL", value, excludeID)
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New(field + " already in use")
	}
	return nil
}

// validateIdentifier checks if the identifier is a valid UUID, username, or email format
func (s *userService) validateIdentifier(identifier string) error {
	if _, err := uuid.Parse(identifier); err == nil {
		return nil
	}
	if s.validator.Var(identifier, "email") == nil {
		return nil
	}
	if s.validator.Var(identifier, "alphanum,min=3,max=30") == nil {
		return nil
	}
	return errors.New("identifier must be a valid UUID, username, or email")
}
