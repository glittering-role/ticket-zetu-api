package organizers_services

import (
	"errors"
	organizer_dto "ticket-zetu-api/modules/organizers/dto"
	organizers "ticket-zetu-api/modules/organizers/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *organizerService) GetOrganizer(userID, id string) (*organizer_dto.OrganizerResponse, error) {
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

	var dbOrganizer organizers.Organizer
	if err := s.db.
		Preload("CreatedByUser").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&dbOrganizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organizer not found")
		}
		return nil, err
	}

	return s.toOrganizerResponse(&dbOrganizer, userID), nil
}

func (s *organizerService) GetMyOrganizer(userID string) (*organizer_dto.OrganizerResponse, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, errors.New("invalid user ID format")
	}

	var dbOrganizer organizers.Organizer
	if err := s.db.
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, email, username, first_name, last_name, avatar_url")
		}).
		Where("created_by = ? AND deleted_at IS NULL", userID).
		First(&dbOrganizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organizer not found")
		}
		return nil, err
	}

	return s.toOrganizerResponse(&dbOrganizer, userID), nil
}

func (s *organizerService) SearchOrganizers(userID, searchTerm string, createdBy uuid.UUID, page, pageSize int) ([]organizer_dto.OrganizerResponse, int64, error) {
	hasPerm, err := s.HasPermission(userID, "view:organizers")
	if err != nil {
		return nil, 0, err
	}
	if !hasPerm {
		return nil, 0, errors.New("user lacks view:organizers permission")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	query := s.db.Model(&organizers.Organizer{}).
		Preload("CreatedByUser").
		Where("deleted_at IS NULL")

	if searchTerm != "" {
		searchTerm = "%" + searchTerm + "%"
		query = query.Where(
			"name LIKE ? OR email LIKE ? OR contact_person LIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	if createdBy != uuid.Nil {
		query = query.Where("created_by = ?", createdBy.String())
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var dbOrganizers []organizers.Organizer
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Find(&dbOrganizers).Error; err != nil {
		return nil, 0, err
	}

	organizers := make([]organizer_dto.OrganizerResponse, 0, len(dbOrganizers))
	for _, dbOrg := range dbOrganizers {
		if resp := s.toOrganizerResponse(&dbOrg, userID); resp != nil {
			organizers = append(organizers, *resp)
		}
	}

	return organizers, total, nil
}
