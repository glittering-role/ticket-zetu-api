package organizers_services

import (
	"errors"
	organizers "ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/users/authorization"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizerService interface {
	CreateOrganizer(userID, name, contactPerson, email, phone, companyName, taxID, bankAccountInfo, imageURL string, commissionRate, balance float64) (*organizers.Organizer, error)
	UpdateOrganizer(userID, id, name, contactPerson, email, phone, companyName, taxID, bankAccountInfo, imageURL string, commissionRate, balance float64, isFlagged, isBanned bool) (*organizers.Organizer, error)
	DeleteOrganizer(userID, id string) error
	DeactivateOrganizer(userID, id string) error
	GetOrganizer(userID, id string) (*organizers.Organizer, error)
	GetOrganizers(userID string) ([]organizers.Organizer, error)
	GetMyOrganizer(userID string) (*organizers.Organizer, error)
	HasPermission(userID, permission string) (bool, error)
}

type organizerService struct {
	db                   *gorm.DB
	authorizationService authorization.PermissionService
}

func NewOrganizerService(db *gorm.DB, authService authorization.PermissionService) OrganizerService {
	return &organizerService{
		db:                   db,
		authorizationService: authService,
	}
}

func (s *organizerService) HasPermission(userID, permission string) (bool, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return false, errors.New("invalid user ID format")
	}
	hasPerm, err := s.authorizationService.HasPermission(userID, permission)
	if err != nil {
		return false, err
	}
	return hasPerm, nil
}

func (s *organizerService) CreateOrganizer(userID, name, contactPerson, email, phone, companyName, taxID, bankAccountInfo, imageURL string, commissionRate, balance float64) (*organizers.Organizer, error) {
	// hasPerm, err := s.HasPermission(userID, "create:organizers")
	// if err != nil {
	// 	return nil, err
	// }
	// if !hasPerm {
	// 	return nil, errors.New("user lacks create:organizers permission")
	// }

	// Check if email already exists
	var existingOrganizer organizers.Organizer
	if err := s.db.Where("email = ? AND deleted_at IS NULL", email).First(&existingOrganizer).Error; err == nil {
		return nil, errors.New("organizer email already exists")
	}

	organizer := organizers.Organizer{
		Name:            name,
		ContactPerson:   contactPerson,
		Email:           email,
		Phone:           phone,
		CompanyName:     companyName,
		TaxID:           taxID,
		BankAccountInfo: bankAccountInfo,
		ImageURL:        imageURL,
		CommissionRate:  commissionRate,
		Balance:         balance,
		Status:          "active",
		IsFlagged:       false,
		IsBanned:        false,
		CreatedBy:       userID,
	}

	if err := s.db.Create(&organizer).Error; err != nil {
		return nil, err
	}

	return &organizer, nil
}

func (s *organizerService) UpdateOrganizer(userID, id, name, contactPerson, email, phone, companyName, taxID, bankAccountInfo, imageURL string, commissionRate, balance float64, isFlagged, isBanned bool) (*organizers.Organizer, error) {
	hasPerm, err := s.HasPermission(userID, "update:organizers")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks update:organizers permission")
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid organizer ID format")
	}

	// Check if email already exists (excluding current organizer)
	var existingOrganizer organizers.Organizer
	if err := s.db.Where("email = ? AND id != ? AND deleted_at IS NULL", email, id).First(&existingOrganizer).Error; err == nil {
		return nil, errors.New("organizer email already exists")
	}

	var organizer organizers.Organizer
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&organizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organizer not found")
		}
		return nil, err
	}

	organizer.Name = name
	organizer.ContactPerson = contactPerson
	organizer.Email = email
	organizer.Phone = phone
	organizer.CompanyName = companyName
	organizer.TaxID = taxID
	organizer.BankAccountInfo = bankAccountInfo
	organizer.ImageURL = imageURL
	organizer.CommissionRate = commissionRate
	organizer.Balance = balance
	organizer.IsFlagged = isFlagged
	organizer.IsBanned = isBanned
	organizer.CreatedBy = userID

	if err := s.db.Save(&organizer).Error; err != nil {
		return nil, err
	}

	return &organizer, nil
}

func (s *organizerService) DeleteOrganizer(userID, id string) error {
	hasPerm, err := s.HasPermission(userID, "delete:organizers")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks delete:organizers permission")
	}

	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid organizer ID format")
	}

	var organizer organizers.Organizer
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&organizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("organizer not found")
		}
		return err
	}

	if organizer.Status == "active" {
		return errors.New("cannot delete an active organizer")
	}

	if err := s.db.Delete(&organizer).Error; err != nil {
		return err
	}

	return nil
}

func (s *organizerService) DeactivateOrganizer(userID, id string) error {
	hasPerm, err := s.HasPermission(userID, "update:organizers")
	if err != nil {
		return err
	}
	if !hasPerm {
		return errors.New("user lacks update:organizers permission")
	}

	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid organizer ID format")
	}

	var organizer organizers.Organizer
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&organizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("organizer not found")
		}
		return err
	}

	if organizer.Status == "inactive" {
		return errors.New("organizer is already inactive")
	}

	organizer.Status = "inactive"
	organizer.CreatedBy = userID

	if err := s.db.Save(&organizer).Error; err != nil {
		return err
	}

	return nil
}

func (s *organizerService) GetOrganizer(userID, id string) (*organizers.Organizer, error) {
	hasPerm, err := s.HasPermission(userID, "view:organizers")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks view:organizers permission")
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid organizer ID format")
	}

	var organizer organizers.Organizer
	if err := s.db.
		Preload("CreatedByUser").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&organizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organizer not found")
		}
		return nil, err
	}

	return &organizer, nil
}

func (s *organizerService) GetOrganizers(userID string) ([]organizers.Organizer, error) {
	hasPerm, err := s.HasPermission(userID, "view:organizers")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks view:organizers permission")
	}

	var organizers []organizers.Organizer
	if err := s.db.
		Preload("CreatedByUser").
		Where("deleted_at IS NULL").
		Find(&organizers).Error; err != nil {
		return nil, err
	}

	return organizers, nil
}

func (s *organizerService) GetMyOrganizer(userID string) (*organizers.Organizer, error) {
	// hasPerm, err := s.HasPermission(userID, "view:organizers")
	// if err != nil {
	// 	return nil, err
	// }
	// if !hasPerm {
	// 	return nil, errors.New("user lacks view:organizers permission")
	// }

	if _, err := uuid.Parse(userID); err != nil {
		return nil, errors.New("invalid user ID format")
	}

	var organizer organizers.Organizer
	if err := s.db.
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, email, username, first_name, last_name")
		}).Where("created_by = ? AND deleted_at IS NULL", userID).
		First(&organizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organizer not found")
		}
		return nil, err
	}

	return &organizer, nil
}
