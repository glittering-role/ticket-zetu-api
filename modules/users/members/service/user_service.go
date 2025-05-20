package members_service

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"ticket-zetu-api/modules/users/models/members"
)

type UserService interface {
	GetUserProfile(identifier string, requesterID string) (*UserProfileResponse, error)
}

type userService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) UserService {
	return &userService{db: db}
}

type UserProfileResponse struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Location  string `json:"location,omitempty"`
	Gender    string `json:"gender,omitempty"`
	RoleName  string `json:"role_name,omitempty"`
}

func (s *userService) GetUserProfile(identifier string, requesterID string) (*UserProfileResponse, error) {
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

	log.Printf("GetUserProfile query: identifier=%s", identifier)
	if err := query.First(&user).Error; err != nil {
		log.Printf("GetUserProfile error: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Check if profile is visible (or requester is the user)
	if user.Preferences.ID == uuid.Nil || (!user.Preferences.ShowProfile && user.ID != requesterID) {
		return nil, errors.New("user profile is private or preferences not set")
	}

	// Build response based on preferences
	response := &UserProfileResponse{
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarURL: user.AvatarURL,
	}

	if user.Preferences.ShouldShow("email") {
		response.Email = user.Email
	}
	if user.Preferences.ShouldShow("phone") {
		response.Phone = user.Phone
	}
	if user.Preferences.ShouldShow("location") && user.Location.ID != uuid.Nil && user.Location.FullLocation() != "Unknown location" {
		response.Location = user.Location.FullLocation()
	}
	if user.Preferences.ShouldShow("gender") && user.Gender != "" {
		response.Gender = user.Gender
	}
	if user.Role.ID != "" {
		response.RoleName = user.Role.RoleName
	}

	return response, nil
}
