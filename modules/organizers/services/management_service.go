package organizers_services

import (
	"errors"
	organizers "ticket-zetu-api/modules/organizers/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *organizerService) authorizeOrganizerAction(userID, organizerID, action string) (*organizers.Organizer, error) {
	isSuperAdmin, err := s.HasPermission(userID, "master:super_admin")
	if err != nil {
		return nil, err
	}

	if !isSuperAdmin {
		hasPerm, err := s.HasPermission(userID, action)
		if err != nil {
			return nil, err
		}
		if !hasPerm {
			return nil, errors.New("user lacks " + action + " permission")
		}
	}

	if _, err := uuid.Parse(organizerID); err != nil {
		return nil, errors.New("invalid organizer ID format")
	}

	var org organizers.Organizer
	if err := s.db.Where("id = ? AND deleted_at IS NULL", organizerID).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organizer not found")
		}
		return nil, err
	}

	if !isSuperAdmin && org.CreatedBy != userID {
		return nil, errors.New("user is not authorized to perform this action on this organizer")
	}

	return &org, nil
}

func (s *organizerService) ToggleOrganizationsStatus(userID, id string) error {
	org, err := s.authorizeOrganizerAction(userID, id, "update:organizers")
	if err != nil {
		return err
	}

	if org.Status == "inactive" {
		org.Status = "active"
	} else {
		org.Status = "inactive"
	}

	return s.db.Save(org).Error
}

func (s *organizerService) FlagOrganization(userID, id string) error {
	org, err := s.authorizeOrganizerAction(userID, id, "update:organizers")
	if err != nil {
		return err
	}

	org.IsFlagged = !org.IsFlagged
	return s.db.Save(org).Error
}

func (s *organizerService) BanOrganization(userID, id string) error {
	org, err := s.authorizeOrganizerAction(userID, id, "update:organizers")
	if err != nil {
		return err
	}

	org.IsBanned = !org.IsBanned
	return s.db.Save(org).Error
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

	var dbOrganizer organizers.Organizer
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&dbOrganizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("organizer not found")
		}
		return err
	}

	if dbOrganizer.Status == "inactive" {
		return errors.New("organizer is already inactive")
	}

	dbOrganizer.Status = "inactive"
	dbOrganizer.CreatedBy = userID

	if err := s.db.Save(&dbOrganizer).Error; err != nil {
		return err
	}

	return nil
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

	var dbOrganizer organizers.Organizer
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&dbOrganizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("organizer not found")
		}
		return err
	}

	if dbOrganizer.Status == "active" {
		return errors.New("cannot delete an active organizer")
	}

	if err := s.db.Delete(&dbOrganizer).Error; err != nil {
		return err
	}

	return nil
}
