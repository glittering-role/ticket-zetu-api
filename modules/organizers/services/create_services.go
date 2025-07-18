package organizers_services

import (
	"errors"
	organizer_dto "ticket-zetu-api/modules/organizers/dto"
	organizers "ticket-zetu-api/modules/organizers/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *organizerService) CreateOrganizer(userID, name, contactPerson, email, phone, companyName, taxID, bankAccountInfo string, commissionRate, balance float64, notes string) (*organizer_dto.OrganizerResponse, error) {
	var existingOrganizer organizers.Organizer
	if err := s.db.Where("email = ? AND deleted_at IS NULL", email).First(&existingOrganizer).Error; err == nil {
		return nil, errors.New("organizer email already exists")
	}

	dbOrganizer := organizers.Organizer{
		Name:            name,
		ContactPerson:   contactPerson,
		Email:           email,
		Phone:           phone,
		CompanyName:     companyName,
		TaxID:           taxID,
		BankAccountInfo: bankAccountInfo,
		CommissionRate:  commissionRate,
		Balance:         balance,
		Status:          "active",
		IsFlagged:       false,
		IsBanned:        false,
		CreatedBy:       userID,
		Notes:           notes,
	}

	if err := s.db.Create(&dbOrganizer).Error; err != nil {
		return nil, err
	}

	return s.toOrganizerResponse(&dbOrganizer, userID), nil
}

func (s *organizerService) UpdateOrganizer(userID string, id, name, contactPerson, email, phone, companyName, taxID, bankAccountInfo string, commissionRate, balance float64, notes string, allowSubscriptions bool) (*organizer_dto.OrganizerResponse, error) {
	_, err := s.HasPermission(userID, "update:organizers")
	if err != nil {
		return nil, err
	}
	// if !hasPerm {
	// 	return nil, errors.New("user lacks update:organizers permission")
	// }

	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid organizer ID format")
	}

	// Check if email already exists (excluding current organizer)
	var existingOrganizer organizers.Organizer
	if err := s.db.Where("email = ? AND id != ? AND deleted_at IS NULL", email, id).First(&existingOrganizer).Error; err == nil {
		return nil, errors.New("organizer email already exists")
	}

	var dbOrganizer organizers.Organizer
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&dbOrganizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organizer not found")
		}
		return nil, err
	}

	dbOrganizer.Name = name
	dbOrganizer.ContactPerson = contactPerson
	dbOrganizer.Email = email
	dbOrganizer.Phone = phone
	dbOrganizer.CompanyName = companyName
	dbOrganizer.TaxID = taxID
	dbOrganizer.BankAccountInfo = bankAccountInfo
	dbOrganizer.CommissionRate = commissionRate
	dbOrganizer.Balance = balance
	dbOrganizer.Notes = notes
	dbOrganizer.AllowSubscriptions = allowSubscriptions

	if err := s.db.Save(&dbOrganizer).Error; err != nil {
		return nil, err
	}

	return s.toOrganizerResponse(&dbOrganizer, userID), nil
}
