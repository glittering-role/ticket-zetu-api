package service

import (
	"errors"
	"ticket-zetu-api/modules/users/authentication/mail"
	auth_utils "ticket-zetu-api/modules/users/authentication/utils"
	"ticket-zetu-api/modules/users/members/dto"
	"ticket-zetu-api/modules/users/models/artist"
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
	UpdateUsername(id string, usernameDto *dto.UpdateUsernameDto, updaterID string) (*dto.UserProfileResponseDto, error)
	SetNewPassword(userID string, userDto *dto.NewPasswordDto, updaterID string) (*dto.UserProfileResponseDto, error)
}

type userService struct {
	db            *gorm.DB
	validator     *validator.Validate
	emailService  mail_service.EmailService
	usernameCheck *auth_utils.UsernameCheck
}

func NewUserService(db *gorm.DB, emailService mail_service.EmailService, usernameCheck *auth_utils.UsernameCheck) UserService {
	return &userService{
		db:            db,
		validator:     validator.New(),
		emailService:  emailService,
		usernameCheck: usernameCheck,
	}
}

// toSafeArtistProfileDto optimized with pointer checks and early returns
func toSafeArtistProfileDto(artistProfile *artist.ArtistProfile, preferences *members.UserPreferences) *dto.ReadArtistProfileDTO {
	if artistProfile == nil {
		return nil
	}

	safeDto := &dto.ReadArtistProfileDTO{
		ID:           artistProfile.ID,
		StageName:    artistProfile.StageName,
		Type:         string(artistProfile.Type),
		Bio:          artistProfile.Bio,
		Website:      artistProfile.Website,
		Genres:       artistProfile.Genres,
		Skills:       artistProfile.Skills,
		SpotifyURL:   artistProfile.SpotifyURL,
		YouTubeURL:   artistProfile.YouTubeURL,
		Instagram:    artistProfile.Instagram,
		TikTok:       artistProfile.TikTok,
		Twitter:      artistProfile.Twitter,
		Reddit:       artistProfile.Reddit,
		Snapchat:     artistProfile.Snapchat,
		Patreon:      artistProfile.Patreon,
		SoundCloud:   artistProfile.SoundCloud,
		Behance:      artistProfile.Behance,
		Dribbble:     artistProfile.Dribbble,
		Vimeo:        artistProfile.Vimeo,
		Goodreads:    artistProfile.Goodreads,
		LinkedIn:     artistProfile.LinkedIn,
		Pinterest:    artistProfile.Pinterest,
		Twitch:       artistProfile.Twitch,
		DeviantArt:   artistProfile.DeviantArt,
		PortfolioURL: artistProfile.PortfolioURL,
	}

	if preferences == nil || preferences.ID == uuid.Nil {
		return safeDto
	}

	if preferences.ShouldShow("artist_contact_email") {
		safeDto.ContactEmail = artistProfile.ContactEmail
	}
	if preferences.ShouldShow("artist_representation") {
		safeDto.Representation = artistProfile.Representation
	}
	if preferences.ShouldShow("artist_availability") {
		safeDto.Availability = artistProfile.Availability
	}
	if preferences.ShouldShow("artist_collaboration") {
		safeDto.Collaboration = artistProfile.Collaboration
	}

	return safeDto
}

// toUserProfileResponseDto optimized with minimal allocations and early returns
func toUserProfileResponseDto(user *members.User, isOwner bool, artistProfile *artist.ArtistProfile) *dto.UserProfileResponseDto {
	if user == nil {
		return nil
	}

	response := &dto.UserProfileResponseDto{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	// Only show verification status if user is verified OR it's the owner viewing their own profile
	if isOwner || user.IsVerified {
		response.IsVerified = user.IsVerified
	}

	if artistProfile != nil {
		if isOwner {
			response.ArtistProfile = convertFullArtistProfile(artistProfile)
		} else if user.Preferences.ID != uuid.Nil && user.Preferences.ShowProfile {
			response.ArtistProfile = toSafeArtistProfileDto(artistProfile, &user.Preferences)
		}
	}

	if !isOwner {
		if user.Preferences.ID == uuid.Nil || !user.Preferences.ShowProfile {
			return response
		}
		return addNonOwnerFields(response, user)
	}

	// Owner-specific fields
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
		response.LocationDetail = convertUserLocation(user.Location)
	}

	return response
}

// Helper functions to reduce cognitive complexity and improve readability
func convertFullArtistProfile(profile *artist.ArtistProfile) *dto.ReadArtistProfileDTO {
	if profile == nil {
		return nil
	}

	fullProfile := &dto.ReadArtistProfileDTO{
		ID:             profile.ID,
		UserID:         profile.UserID,
		StageName:      profile.StageName,
		Type:           string(profile.Type),
		Bio:            profile.Bio,
		Website:        profile.Website,
		Location:       profile.Location,
		Collaboration:  profile.Collaboration,
		SpotifyURL:     profile.SpotifyURL,
		YouTubeURL:     profile.YouTubeURL,
		Instagram:      profile.Instagram,
		TikTok:         profile.TikTok,
		Twitter:        profile.Twitter,
		Reddit:         profile.Reddit,
		Snapchat:       profile.Snapchat,
		Patreon:        profile.Patreon,
		SoundCloud:     profile.SoundCloud,
		Behance:        profile.Behance,
		Dribbble:       profile.Dribbble,
		Vimeo:          profile.Vimeo,
		Goodreads:      profile.Goodreads,
		LinkedIn:       profile.LinkedIn,
		Pinterest:      profile.Pinterest,
		Twitch:         profile.Twitch,
		DeviantArt:     profile.DeviantArt,
		PortfolioURL:   profile.PortfolioURL,
		Genres:         profile.Genres,
		Skills:         profile.Skills,
		Availability:   profile.Availability,
		ContactEmail:   profile.ContactEmail,
		Representation: profile.Representation,
		CreatedAt:      profile.CreatedAt,
		UpdatedAt:      profile.UpdatedAt,
	}

	if !profile.DeletedAt.Time.IsZero() {
		deletedAt := profile.DeletedAt.Time
		fullProfile.DeletedAt = &deletedAt
	}

	return fullProfile
}

func convertUserLocation(location members.UserLocation) *dto.UserLocationDto {
	locDto := &dto.UserLocationDto{
		Country:   location.Country,
		State:     location.State,
		StateName: location.StateName,
		Continent: location.Continent,
		City:      location.City,
		Zip:       location.Zip,
		Timezone:  location.Timezone,
	}

	if location.LastActive != nil {
		lastActive := location.LastActive.Format(time.RFC3339)
		locDto.LastActive = &lastActive
	}

	return locDto
}

func addNonOwnerFields(response *dto.UserProfileResponseDto, user *members.User) *dto.UserProfileResponseDto {
	if user.Preferences.ShouldShow("email") {
		response.Email = user.Email
	}
	if user.Preferences.ShouldShow("phone") {
		response.Phone = user.Phone
	}
	if user.Preferences.ShouldShow("location") && user.Location.ID != uuid.Nil {
		response.Location = user.Location.FullLocation()
		response.LocationDetail = convertUserLocation(user.Location)
	}
	if user.Preferences.ShouldShow("gender") {
		response.Gender = user.Gender
	}
	if user.Preferences.ShouldShow("role") && user.Role.ID != "" {
		response.RoleName = user.Role.RoleName
	}

	return response
}

// GetUserProfile optimized with select fields to reduce database load
func (s *userService) GetUserProfile(identifier string, requesterID string) (*dto.UserProfileResponseDto, error) {
	if identifier == "" {
		return nil, errors.New("identifier is required")
	}

	if err := s.validateIdentifier(identifier); err != nil {
		return nil, errors.New("invalid identifier format: " + err.Error())
	}

	var user members.User
	query := s.db.
		Select("id, username, first_name, last_name, avatar_url, email, phone, date_of_birth, gender, is_verified, created_at, updated_at").
		Preload("Role", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, role_name")
		}).
		Preload("Preferences").
		Preload("Location").
		Preload("ArtistProfile", "deleted_at IS NULL").
		Where("deleted_at IS NULL")

	if uuid, err := uuid.Parse(identifier); err == nil {
		query = query.Where("id = ?", uuid.String())
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
	var artistProfile *artist.ArtistProfile
	if user.ArtistProfile.ID != "" {
		artistProfile = &user.ArtistProfile
	}

	return toUserProfileResponseDto(&user, isOwner, artistProfile), nil
}

func (s *userService) checkUniqueField(tx *gorm.DB, field string, value string, excludeID string) error {
	var exists bool
	err := tx.Model(&members.User{}).
		Select("1").
		Where(field+" = ? AND id != ? AND deleted_at IS NULL", value, excludeID).
		Limit(1).
		Find(&exists).Error

	if err != nil {
		return err
	}
	if exists {
		return errors.New(field + " already in use")
	}
	return nil
}

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
