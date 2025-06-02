package organizers_services

import (
	"errors"
	"ticket-zetu-api/cloudinary"
	organizer_dto "ticket-zetu-api/modules/organizers/dto"
	organizers "ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/users/authorization"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizerService interface {
	CreateOrganizer(userID, name, contactPerson, email, phone, companyName, taxID, bankAccountInfo string, commissionRate, balance float64, notes string) (*organizer_dto.OrganizerResponse, error)
	UpdateOrganizer(userID, id, name, contactPerson, email, phone, companyName, taxID, bankAccountInfo string, commissionRate, balance float64, notes string) (*organizer_dto.OrganizerResponse, error)
	DeleteOrganizer(userID, id string) error
	DeactivateOrganizer(userID, id string) error
	GetOrganizer(userID, id string) (*organizer_dto.OrganizerResponse, error)
	GetOrganizers(userID string) ([]organizer_dto.OrganizerResponse, error)
	GetMyOrganizer(userID string) (*organizer_dto.OrganizerResponse, error)
	HasPermission(userID, permission string) (bool, error)
	SearchOrganizers(userID, searchTerm string, createdBy uuid.UUID, page, pageSize int) ([]organizer_dto.OrganizerResponse, int64, error)
	ToggleOrganizationsStatus(userID, id string) error
	BanOrganization(userID, id string) error
	FlagOrganization(userID, id string) error
}

type organizerService struct {
	db                   *gorm.DB
	authorizationService authorization.PermissionService
	cloudinaryService    *cloudinary.CloudinaryService
}

func NewOrganizerService(db *gorm.DB, authService authorization.PermissionService) OrganizerService {
	return &organizerService{
		db:                   db,
		authorizationService: authService,
	}
}

func NewBaseService(db *gorm.DB, authService authorization.PermissionService, cloudinary *cloudinary.CloudinaryService) *organizerService {
	return &organizerService{
		db:                   db,
		authorizationService: authService,
	}
}

func (s *organizerService) toOrganizerResponse(dbOrganizer *organizers.Organizer) *organizer_dto.OrganizerResponse {
	if dbOrganizer == nil {
		return nil
	}

	resp := &organizer_dto.OrganizerResponse{
		ID:              dbOrganizer.ID,
		Name:            dbOrganizer.Name,
		ContactPerson:   dbOrganizer.ContactPerson,
		Email:           dbOrganizer.Email,
		Phone:           dbOrganizer.Phone,
		CompanyName:     dbOrganizer.CompanyName,
		TaxID:           dbOrganizer.TaxID,
		BankAccountInfo: dbOrganizer.BankAccountInfo,
		ImageURL:        dbOrganizer.ImageURL,
		CommissionRate:  dbOrganizer.CommissionRate,
		Balance:         dbOrganizer.Balance,
		Status:          dbOrganizer.Status,
		IsFlagged:       dbOrganizer.IsFlagged,
		IsBanned:        dbOrganizer.IsBanned,
		CreatedBy:       dbOrganizer.CreatedBy,
	}

	if dbOrganizer.CreatedByUser.ID != "" {
		resp.CreatedByUser = &organizer_dto.UserResponse{
			ID:        dbOrganizer.CreatedByUser.ID,
			Email:     dbOrganizer.CreatedByUser.Email,
			Username:  dbOrganizer.CreatedByUser.Username,
			FirstName: dbOrganizer.CreatedByUser.FirstName,
			LastName:  dbOrganizer.CreatedByUser.LastName,
		}
	}

	return resp
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

	return s.toOrganizerResponse(&dbOrganizer), nil
}

func (s *organizerService) GetOrganizers(userID string) ([]organizer_dto.OrganizerResponse, error) {
	hasPerm, err := s.HasPermission(userID, "view:organizers")
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return nil, errors.New("user lacks view:organizers permission")
	}

	var dbOrganizers []organizers.Organizer
	if err := s.db.
		Preload("CreatedByUser").
		Where("deleted_at IS NULL").
		Find(&dbOrganizers).Error; err != nil {
		return nil, err
	}

	organizers := make([]organizer_dto.OrganizerResponse, 0, len(dbOrganizers))
	for _, dbOrg := range dbOrganizers {
		if resp := s.toOrganizerResponse(&dbOrg); resp != nil {
			organizers = append(organizers, *resp)
		}
	}

	return organizers, nil
}

func (s *organizerService) GetMyOrganizer(userID string) (*organizer_dto.OrganizerResponse, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, errors.New("invalid user ID format")
	}

	var dbOrganizer organizers.Organizer
	if err := s.db.
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, email, username, first_name, last_name")
		}).
		Where("created_by = ? AND deleted_at IS NULL", userID).
		First(&dbOrganizer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organizer not found")
		}
		return nil, err
	}

	return s.toOrganizerResponse(&dbOrganizer), nil
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
		if resp := s.toOrganizerResponse(&dbOrg); resp != nil {
			organizers = append(organizers, *resp)
		}
	}

	return organizers, total, nil
}
