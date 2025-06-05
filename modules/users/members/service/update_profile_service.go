package service

import (
	"errors"
	"ticket-zetu-api/modules/users/members/dto"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *userService) UpdateUserDetails(id string, userDto *dto.UpdateUserDto, updaterID string) (*dto.UserProfileResponseDto, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid user ID format")
	}

	if err := s.validator.Struct(userDto); err != nil {
		return nil, errors.New("validation failed: " + err.Error())
	}

	var updatedUser members.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var user members.User
		if err := tx.Preload("Preferences").Preload("Location").Preload("Role").
			Where("id = ? AND deleted_at IS NULL", id).
			First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			return err
		}

		updates := make(map[string]interface{})
		if userDto.FirstName != nil {
			updates["first_name"] = *userDto.FirstName
		}
		if userDto.LastName != nil {
			updates["last_name"] = *userDto.LastName
		}
		if userDto.AvatarURL != nil {
			updates["avatar_url"] = *userDto.AvatarURL
		}
		if userDto.DateOfBirth != nil {
			var dob *time.Time
			if *userDto.DateOfBirth != "" {
				parsedDob, err := time.Parse("2006-01-02", *userDto.DateOfBirth)
				if err != nil {
					return errors.New("invalid date_of_birth format")
				}
				dob = &parsedDob
			}
			updates["date_of_birth"] = dob
		}
		if userDto.Gender != nil {
			updates["gender"] = *userDto.Gender
		}
		updates["last_modified_by"] = updaterID
		updates["updated_at"] = time.Now()

		if len(updates) > 0 {
			if err := tx.Model(&user).Updates(updates).Error; err != nil {
				return err
			}
		}

		if err := tx.Preload("Preferences").Preload("Location").Preload("Role").
			Where("id = ?", id).First(&updatedUser).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return toUserProfileResponseDto(&updatedUser, true), nil
}

func (s *userService) UpdateUserLocation(id string, locationDto *dto.UserLocationUpdateDto, updaterID string) (*dto.UserProfileResponseDto, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid user ID format")
	}

	if err := s.validator.Struct(locationDto); err != nil {
		return nil, errors.New("validation failed: " + err.Error())
	}

	var updatedUser members.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var user members.User
		if err := tx.Preload("Location").Where("id = ? AND deleted_at IS NULL", id).
			First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			return err
		}

		var location members.UserLocation
		isNewLocation := user.Location.ID == uuid.Nil
		if isNewLocation {
			location = members.UserLocation{
				ID:     uuid.New(),
				UserID: uuid.MustParse(user.ID),
			}
		} else {
			if err := tx.Where("id = ?", user.Location.ID).First(&location).Error; err != nil {
				return err
			}
		}

		updates := make(map[string]interface{})
		if locationDto.Country != nil {
			updates["country"] = *locationDto.Country
			location.Country = *locationDto.Country
		}
		if locationDto.State != nil {
			updates["state"] = *locationDto.State
			location.State = *locationDto.State
		}
		if locationDto.StateName != nil {
			updates["state_name"] = *locationDto.StateName
			location.StateName = *locationDto.StateName
		}
		if locationDto.Continent != nil {
			updates["continent"] = *locationDto.Continent
			location.Continent = *locationDto.Continent
		}
		if locationDto.City != nil {
			updates["city"] = *locationDto.City
			location.City = *locationDto.City
		}
		if locationDto.Zip != nil {
			updates["zip"] = *locationDto.Zip
			location.Zip = *locationDto.Zip
		}
		if locationDto.Timezone != nil {
			updates["timezone"] = *locationDto.Timezone
			location.Timezone = *locationDto.Timezone
		}
		if locationDto.LastActive != nil {
			var lastActive *time.Time
			if *locationDto.LastActive != "" {
				parsedTime, err := time.Parse(time.RFC3339, *locationDto.LastActive)
				if err != nil {
					return errors.New("invalid last_active format")
				}
				lastActive = &parsedTime
				location.LastActive = lastActive
			}
			updates["last_active"] = lastActive
		}

		if len(updates) == 0 {
			return errors.New("no location updates provided")
		}

		if isNewLocation {
			if err := tx.Create(&location).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&location).Updates(updates).Error; err != nil {
				return err
			}
		}

		if err := tx.Preload("Preferences").Preload("Location").Preload("Role").
			Where("id = ?", id).First(&updatedUser).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return toUserProfileResponseDto(&updatedUser, true), nil
}
