package service

import (
	"errors"
	"ticket-zetu-api/modules/users/members/dto"
	"ticket-zetu-api/modules/users/models/artist"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ArtistService defines the interface for artist profile operations
type ArtistService interface {
	CreateArtistProfile(artistDto *dto.CreateArtistProfileDTO, userID string) (*dto.ReadArtistProfileDTO, error)
	UpdateArtistProfile(userID string, artistDto *dto.UpdateArtistProfileDTO) (*dto.ReadArtistProfileDTO, error)
	DeleteArtistProfile(userID string) error
	GetArtistProfileByUserID(userID string) (*dto.ReadArtistProfileDTO, error)
}

// artistService implements the ArtistService interface
type artistService struct {
	db        *gorm.DB
	validator *validator.Validate
}

// NewArtistService initializes a new ArtistService
func NewArtistService(db *gorm.DB) ArtistService {
	v := validator.New()
	return &artistService{
		db:        db,
		validator: v,
	}
}

// toReadArtistProfileDTO converts an artist.ArtistProfile to dto.ReadArtistProfileDTO
func toReadArtistProfileDTO(artistProfile *artist.ArtistProfile) *dto.ReadArtistProfileDTO {
	var deletedAt *time.Time
	if !artistProfile.DeletedAt.Time.IsZero() {
		deletedAt = &artistProfile.DeletedAt.Time
	}

	return &dto.ReadArtistProfileDTO{
		ID:             artistProfile.ID,
		UserID:         artistProfile.UserID,
		StageName:      artistProfile.StageName,
		Type:           string(artistProfile.Type),
		Bio:            artistProfile.Bio,
		Website:        artistProfile.Website,
		Location:       artistProfile.Location,
		Collaboration:  artistProfile.Collaboration,
		SpotifyURL:     artistProfile.SpotifyURL,
		YouTubeURL:     artistProfile.YouTubeURL,
		Instagram:      artistProfile.Instagram,
		TikTok:         artistProfile.TikTok,
		Twitter:        artistProfile.Twitter,
		Reddit:         artistProfile.Reddit,
		Snapchat:       artistProfile.Snapchat,
		Patreon:        artistProfile.Patreon,
		SoundCloud:     artistProfile.SoundCloud,
		Behance:        artistProfile.Behance,
		Dribbble:       artistProfile.Dribbble,
		Vimeo:          artistProfile.Vimeo,
		Goodreads:      artistProfile.Goodreads,
		LinkedIn:       artistProfile.LinkedIn,
		Pinterest:      artistProfile.Pinterest,
		Twitch:         artistProfile.Twitch,
		DeviantArt:     artistProfile.DeviantArt,
		PortfolioURL:   artistProfile.PortfolioURL,
		Genres:         artistProfile.Genres,
		Skills:         artistProfile.Skills,
		Availability:   artistProfile.Availability,
		ContactEmail:   artistProfile.ContactEmail,
		Representation: artistProfile.Representation,
		CreatedAt:      artistProfile.CreatedAt,
		UpdatedAt:      artistProfile.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}

// CreateArtistProfile creates a new artist profile
func (s *artistService) CreateArtistProfile(artistDto *dto.CreateArtistProfileDTO, creatorID string) (*dto.ReadArtistProfileDTO, error) {
	// Validate DTO
	if err := s.validator.Struct(artistDto); err != nil {
		return nil, errors.New("validation failed: " + err.Error())
	}

	// Check if UserID exists
	var user members.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", creatorID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Check if UserID already has an artist profile
	var count int64
	if err := s.db.Model(&artist.ArtistProfile{}).Where("user_id = ? AND deleted_at IS NULL", creatorID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("user already has an artist profile")
	}

	// Begin transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Map DTO to model
	artistProfile := &artist.ArtistProfile{
		UserID:         creatorID,
		StageName:      artistDto.StageName,
		Type:           artist.ArtistType(artistDto.Type),
		Bio:            artistDto.Bio,
		Website:        artistDto.Website,
		Location:       artistDto.Location,
		Collaboration:  artistDto.Collaboration,
		SpotifyURL:     artistDto.SpotifyURL,
		YouTubeURL:     artistDto.YouTubeURL,
		Instagram:      artistDto.Instagram,
		TikTok:         artistDto.TikTok,
		Twitter:        artistDto.Twitter,
		Reddit:         artistDto.Reddit,
		Snapchat:       artistDto.Snapchat,
		Patreon:        artistDto.Patreon,
		SoundCloud:     artistDto.SoundCloud,
		Behance:        artistDto.Behance,
		Dribbble:       artistDto.Dribbble,
		Vimeo:          artistDto.Vimeo,
		Goodreads:      artistDto.Goodreads,
		LinkedIn:       artistDto.LinkedIn,
		Pinterest:      artistDto.Pinterest,
		Twitch:         artistDto.Twitch,
		DeviantArt:     artistDto.DeviantArt,
		PortfolioURL:   artistDto.PortfolioURL,
		Genres:         string(artistDto.Genres),
		Skills:         string(artistDto.Skills),
		Availability:   artistDto.Availability,
		ContactEmail:   artistDto.ContactEmail,
		Representation: artistDto.Representation,
	}

	// Save to database
	if err := tx.Create(artistProfile).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return toReadArtistProfileDTO(artistProfile), nil
}

// UpdateArtistProfile updates an existing artist profile
func (s *artistService) UpdateArtistProfile(userID string, artistDto *dto.UpdateArtistProfileDTO) (*dto.ReadArtistProfileDTO, error) {
	if err := s.validator.Struct(artistDto); err != nil {
		return nil, errors.New("validation failed: " + err.Error())
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var artistProfile artist.ArtistProfile
	if err := tx.Where("user_id = ? AND deleted_at IS NULL", userID).First(&artistProfile).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("artist profile not found")
		}
		return nil, err
	}

	// Update fields (same as before)
	if artistDto.StageName != "" {
		artistProfile.StageName = artistDto.StageName
	}

	if artistDto.Type != "" {
		artistProfile.Type = artist.ArtistType(artistDto.Type)
	}
	if artistDto.Bio != "" {
		artistProfile.Bio = artistDto.Bio
	}
	if artistDto.Website != "" {
		artistProfile.Website = artistDto.Website
	}
	if artistDto.Location != "" {
		artistProfile.Location = artistDto.Location
	}
	if artistDto.Collaboration != nil {
		artistProfile.Collaboration = *artistDto.Collaboration
	}
	if artistDto.SpotifyURL != "" {
		artistProfile.SpotifyURL = artistDto.SpotifyURL
	}
	if artistDto.YouTubeURL != "" {
		artistProfile.YouTubeURL = artistDto.YouTubeURL
	}
	if artistDto.Instagram != "" {
		artistProfile.Instagram = artistDto.Instagram
	}
	if artistDto.TikTok != "" {
		artistProfile.TikTok = artistDto.TikTok
	}
	if artistDto.Twitter != "" {
		artistProfile.Twitter = artistDto.Twitter
	}
	if artistDto.Reddit != "" {
		artistProfile.Reddit = artistDto.Reddit
	}
	if artistDto.Snapchat != "" {
		artistProfile.Snapchat = artistDto.Snapchat
	}
	if artistDto.Patreon != "" {
		artistProfile.Patreon = artistDto.Patreon
	}
	if artistDto.SoundCloud != "" {
		artistProfile.SoundCloud = artistDto.SoundCloud
	}
	if artistDto.Behance != "" {
		artistProfile.Behance = artistDto.Behance
	}
	if artistDto.Dribbble != "" {
		artistProfile.Dribbble = artistDto.Dribbble
	}
	if artistDto.Vimeo != "" {
		artistProfile.Vimeo = artistDto.Vimeo
	}
	if artistDto.Goodreads != "" {
		artistProfile.Goodreads = artistDto.Goodreads
	}
	if artistDto.LinkedIn != "" {
		artistProfile.LinkedIn = artistDto.LinkedIn
	}
	if artistDto.Pinterest != "" {
		artistProfile.Pinterest = artistDto.Pinterest
	}
	if artistDto.Twitch != "" {
		artistProfile.Twitch = artistDto.Twitch
	}
	if artistDto.DeviantArt != "" {
		artistProfile.DeviantArt = artistDto.DeviantArt
	}
	if artistDto.PortfolioURL != "" {
		artistProfile.PortfolioURL = artistDto.PortfolioURL
	}
	if len(artistDto.Genres) > 0 {
		artistProfile.Genres = string(artistDto.Genres)
	}
	if len(artistDto.Skills) > 0 {
		artistProfile.Skills = string(artistDto.Skills)
	}

	if artistDto.Availability != "" {
		artistProfile.Availability = artistDto.Availability
	}
	if artistDto.ContactEmail != "" {
		artistProfile.ContactEmail = artistDto.ContactEmail
	}
	if artistDto.Representation != "" {
		artistProfile.Representation = artistDto.Representation
	}

	if err := tx.Save(&artistProfile).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return toReadArtistProfileDTO(&artistProfile), nil
}

// DeleteArtistProfile soft-deletes an artist profile
func (s *artistService) DeleteArtistProfile(userID string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var artistProfile artist.ArtistProfile
	if err := tx.Where("user_id = ? AND deleted_at IS NULL", userID).First(&artistProfile).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("artist profile not found")
		}
		return err
	}

	if err := tx.Delete(&artistProfile).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// GetArtistProfile retrieves an artist profile by ID or UserID
func (s *artistService) GetArtistProfileByUserID(userID string) (*dto.ReadArtistProfileDTO, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, errors.New("invalid user ID format")
	}

	var artistProfile artist.ArtistProfile
	err := s.db.
		Where("user_id = ? AND deleted_at IS NULL", userID).
		First(&artistProfile).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("artist profile not found")
		}
		return nil, err
	}

	return toReadArtistProfileDTO(&artistProfile), nil
}
