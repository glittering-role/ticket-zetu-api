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

// UserPreferencesService defines the interface for user preferences operations
type UserPreferencesService interface {
	GetUserPreferences(userID string) (*dto.UserPreferencesDto, error)
	UpdateUserPreferences(userID string, preferencesDto *dto.UpdateUserPreferencesDto) (*dto.UserPreferencesDto, error)
}

type userPreferencesService struct {
	db        *gorm.DB
	validator *validator.Validate
}

// NewUserPreferencesService initializes the user preferences service
func NewUserPreferencesService(db *gorm.DB) UserPreferencesService {
	return &userPreferencesService{
		db:        db,
		validator: validator.New(),
	}
}

// toUserPreferencesDto converts a members.UserPreferences to dto.UserPreferencesDto
func toUserPreferencesDto(preferences *members.UserPreferences) *dto.UserPreferencesDto {
	return &dto.UserPreferencesDto{
		ID:             preferences.ID,
		UserID:         preferences.UserID,
		ShowEmail:      preferences.ShowEmail,
		ShowPhone:      preferences.ShowPhone,
		ShowLocation:   preferences.ShowLocation,
		ShowGender:     preferences.ShowGender,
		ShowRole:       preferences.ShowRole,
		ShowProfile:    preferences.ShowProfile,
		AllowFollowing: preferences.AllowFollowing,
		Language:       preferences.Language,
		Theme:          preferences.Theme,
		Timezone:       preferences.Timezone,
		CreatedAt:      preferences.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      preferences.UpdatedAt.Format(time.RFC3339),
	}
}

// GetUserPreferences retrieves a user's preferences
func (s *userPreferencesService) GetUserPreferences(userID string) (*dto.UserPreferencesDto, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, errors.New("invalid user ID format")
	}

	var preferences members.UserPreferences
	if err := s.db.Where("user_id = ? AND deleted_at IS NULL", userID).
		First(&preferences).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create default preferences if none exist
			preferences = members.UserPreferences{
				ID:             uuid.New(),
				UserID:         uuid.MustParse(userID),
				ShowEmail:      false,
				ShowPhone:      false,
				ShowLocation:   false,
				ShowGender:     false,
				ShowRole:       false,
				ShowProfile:    true,
				AllowFollowing: true,
				Language:       "en",
				Theme:          "light",
				Timezone:       "UTC",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			if err := s.db.Create(&preferences).Error; err != nil {
				return nil, errors.New("failed to create default preferences: " + err.Error())
			}
			return toUserPreferencesDto(&preferences), nil
		}
		return nil, errors.New("failed to retrieve preferences: " + err.Error())
	}

	return toUserPreferencesDto(&preferences), nil
}

// UpdateUserPreferences updates a user's preferences
func (s *userPreferencesService) UpdateUserPreferences(userID string, preferencesDto *dto.UpdateUserPreferencesDto) (*dto.UserPreferencesDto, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, errors.New("invalid user ID format")
	}

	if err := s.validator.Struct(preferencesDto); err != nil {
		return nil, errors.New("validation failed: " + err.Error())
	}

	var updatedPreferences members.UserPreferences
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var preferences members.UserPreferences
		if err := tx.Where("user_id = ? AND deleted_at IS NULL", userID).
			First(&preferences).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new preferences if none exist
				preferences = members.UserPreferences{
					ID:             uuid.New(),
					UserID:         uuid.MustParse(userID),
					ShowEmail:      false,
					ShowPhone:      false,
					ShowLocation:   false,
					ShowGender:     false,
					ShowRole:       false,
					ShowProfile:    true,
					AllowFollowing: true,
					Language:       "en",
					Theme:          "light",
					Timezone:       "UTC",
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				}
			} else {
				return errors.New("failed to retrieve preferences: " + err.Error())
			}
		}

		updates := make(map[string]interface{})
		if preferencesDto.ShowEmail != nil {
			updates["show_email"] = *preferencesDto.ShowEmail
			preferences.ShowEmail = *preferencesDto.ShowEmail
		}
		if preferencesDto.ShowPhone != nil {
			updates["show_phone"] = *preferencesDto.ShowPhone
			preferences.ShowPhone = *preferencesDto.ShowPhone
		}
		if preferencesDto.ShowLocation != nil {
			updates["show_location"] = *preferencesDto.ShowLocation
			preferences.ShowLocation = *preferencesDto.ShowLocation
		}
		if preferencesDto.ShowGender != nil {
			updates["show_gender"] = *preferencesDto.ShowGender
			preferences.ShowGender = *preferencesDto.ShowGender
		}
		if preferencesDto.ShowRole != nil {
			updates["show_role"] = *preferencesDto.ShowRole
			preferences.ShowRole = *preferencesDto.ShowRole
		}
		if preferencesDto.ShowProfile != nil {
			updates["show_profile"] = *preferencesDto.ShowProfile
			preferences.ShowProfile = *preferencesDto.ShowProfile
		}
		if preferencesDto.AllowFollowing != nil {
			updates["allow_following"] = *preferencesDto.AllowFollowing
			preferences.AllowFollowing = *preferencesDto.AllowFollowing
		}
		if preferencesDto.Language != nil {
			updates["language"] = *preferencesDto.Language
			preferences.Language = *preferencesDto.Language
		}
		if preferencesDto.Theme != nil {
			updates["theme"] = *preferencesDto.Theme
			preferences.Theme = *preferencesDto.Theme
		}
		if preferencesDto.Timezone != nil {
			updates["timezone"] = *preferencesDto.Timezone
			preferences.Timezone = *preferencesDto.Timezone
		}
		updates["updated_at"] = time.Now()

		if len(updates) == 0 {
			return errors.New("no updates provided")
		}

		if preferences.ID == uuid.Nil {
			if err := tx.Create(&preferences).Error; err != nil {
				return errors.New("failed to create preferences: " + err.Error())
			}
		} else {
			if err := tx.Model(&preferences).Updates(updates).Error; err != nil {
				return errors.New("failed to update preferences: " + err.Error())
			}
		}

		updatedPreferences = preferences
		return nil
	})

	if err != nil {
		return nil, err
	}

	return toUserPreferencesDto(&updatedPreferences), nil
}
