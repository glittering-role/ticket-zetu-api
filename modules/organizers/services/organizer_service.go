package organizers_services

import (
	"errors"
	"ticket-zetu-api/cloudinary"
	organizer_dto "ticket-zetu-api/modules/organizers/dto"
	organizers "ticket-zetu-api/modules/organizers/models"
	"ticket-zetu-api/modules/users/authorization/service"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizerService interface {
	CreateOrganizer(userID, name, contactPerson, email, phone, companyName, taxID, bankAccountInfo string, commissionRate, balance float64, notes string) (*organizer_dto.OrganizerResponse, error)
	UpdateOrganizer(userID, id, name, contactPerson, email, phone, companyName, taxID, bankAccountInfo string, commissionRate, balance float64, notes string, allowSubscriptions bool) (*organizer_dto.OrganizerResponse, error)
	DeleteOrganizer(userID, id string) error
	DeactivateOrganizer(userID, id string) error
	GetOrganizer(userID, id string) (*organizer_dto.OrganizerResponse, error)
	GetOrganizers(userID string, page, pageSize int) ([]interface{}, int64, error)
	GetMyOrganizer(userID string) (*organizer_dto.OrganizerResponse, error)
	HasPermission(userID, permission string) (bool, error)
	SearchOrganizers(userID, searchTerm string, createdBy uuid.UUID, page, pageSize int) ([]organizer_dto.OrganizerResponse, int64, error)
	ToggleOrganizationsStatus(userID, id string) error
	BanOrganization(userID, id string) error
	FlagOrganization(userID, id string) error
}

type organizerService struct {
	db                   *gorm.DB
	authorizationService authorization_service.PermissionService
	cloudinaryService    *cloudinary.CloudinaryService
}

func NewOrganizerService(db *gorm.DB, authService authorization_service.PermissionService) OrganizerService {
	return &organizerService{
		db:                   db,
		authorizationService: authService,
	}
}

func NewBaseService(db *gorm.DB, authService authorization_service.PermissionService, cloudinary *cloudinary.CloudinaryService) *organizerService {
	return &organizerService{
		db:                   db,
		authorizationService: authService,
		cloudinaryService:    cloudinary,
	}
}

func (s *organizerService) GetOrganizers(userID string, page, pageSize int) ([]interface{}, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	query := s.db.Model(&organizers.Organizer{}).
		Preload("CreatedByUser").
		Where("deleted_at IS NULL")

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

	isSuperadmin, err := s.HasPermission(userID, "superadmin")
	if err != nil {
		return nil, 0, err
	}

	organizers := make([]interface{}, 0, len(dbOrganizers))
	for _, dbOrg := range dbOrganizers {
		isOwner := dbOrg.CreatedBy == userID
		if isSuperadmin || isOwner {
			if resp := s.toOrganizerResponse(&dbOrg, userID); resp != nil {
				organizers = append(organizers, *resp)
			}
		} else {
			if resp := s.toBasicOrganizerResponse(&dbOrg, userID); resp != nil {
				organizers = append(organizers, *resp)
			}
		}
	}

	return organizers, total, nil
}

func (s *organizerService) toBasicOrganizerResponse(dbOrg *organizers.Organizer, currentUserID string) *organizer_dto.BasicOrganizerResponse {
	resp := &organizer_dto.BasicOrganizerResponse{
		ID:                     dbOrg.ID,
		Name:                   dbOrg.Name,
		CompanyName:            dbOrg.CompanyName,
		ImageURL:               dbOrg.ImageURL,
		IsActive:               dbOrg.Status == "active",
		IsAcceptingSubscribers: dbOrg.IsAcceptingSubscriptions,
		CreatedAt:              dbOrg.CreatedAt,
	}

	if dbOrg.IsAcceptingSubscriptions {
		var subscription organizers.OrganizationSubscription
		err := s.db.
			Where("organizer_id = ? AND subscriber_id = ? AND is_active = ? AND unsubscribed_at IS NULL",
				dbOrg.ID, currentUserID, true).
			First(&subscription).Error
		if err == nil {
			resp.CurrentUserSubscription = &organizer_dto.SubscriptionInfo{
				IsSubscribed:           true,
				SubscriptionDate:       subscription.SubscriptionDate,
				ReceiveEventUpdates:    subscription.ReceiveEventUpdates,
				ReceiveNewsletters:     subscription.ReceiveNewsletters,
				ReceivePromotions:      subscription.ReceivePromotions,
				NotificationPreference: subscription.NotificationTypes,
			}
		} else {
			resp.CurrentUserSubscription = &organizer_dto.SubscriptionInfo{
				IsSubscribed:           false,
				SubscriptionDate:       time.Time{},
				ReceiveEventUpdates:    false,
				ReceiveNewsletters:     false,
				ReceivePromotions:      false,
				NotificationPreference: "",
			}
		}
	}

	return resp
}

func (s *organizerService) toOrganizerResponse(dbOrganizer *organizers.Organizer, currentUserID string) *organizer_dto.OrganizerResponse {
	if dbOrganizer == nil {
		return nil
	}

	resp := &organizer_dto.OrganizerResponse{
		ID:                     dbOrganizer.ID,
		Name:                   dbOrganizer.Name,
		ContactPerson:          dbOrganizer.ContactPerson,
		Email:                  dbOrganizer.Email,
		Phone:                  dbOrganizer.Phone,
		CompanyName:            dbOrganizer.CompanyName,
		TaxID:                  dbOrganizer.TaxID,
		BankAccountInfo:        dbOrganizer.BankAccountInfo,
		ImageURL:               dbOrganizer.ImageURL,
		CommissionRate:         dbOrganizer.CommissionRate,
		Balance:                dbOrganizer.Balance,
		Status:                 dbOrganizer.Status,
		IsFlagged:              dbOrganizer.IsFlagged,
		IsBanned:               dbOrganizer.IsBanned,
		CreatedBy:              dbOrganizer.CreatedBy,
		SubscriberCount:        dbOrganizer.SubscriberCount,
		IsAcceptingSubscribers: dbOrganizer.IsAcceptingSubscriptions,
	}

	if dbOrganizer.CreatedByUser.ID != "" {
		resp.CreatedByUser = &organizer_dto.UserResponse{
			ID:        dbOrganizer.CreatedByUser.ID,
			Email:     dbOrganizer.CreatedByUser.Email,
			Username:  dbOrganizer.CreatedByUser.Username,
			FirstName: dbOrganizer.CreatedByUser.FirstName,
			LastName:  dbOrganizer.CreatedByUser.LastName,
			AvatarURL: dbOrganizer.CreatedByUser.AvatarURL,
		}
	}

	if currentUserID != "" {
		isOwner := dbOrganizer.CreatedBy == currentUserID
		if isOwner {
			var subscriptions []organizers.OrganizationSubscription
			if err := s.db.
				Where("organizer_id = ? AND is_active = ? AND unsubscribed_at IS NULL",
					dbOrganizer.ID, true).
				Find(&subscriptions).Error; err == nil {
				resp.CurrentUserSubscription = &organizer_dto.SubscriptionInfo{
					IsSubscribed: len(subscriptions) > 0,
				}
			}
		} else {
			var subscription organizers.OrganizationSubscription
			err := s.db.
				Where("organizer_id = ? AND subscriber_id = ? AND is_active = ? AND unsubscribed_at IS NULL",
					dbOrganizer.ID, currentUserID, true).
				First(&subscription).Error
			if err == nil {
				resp.CurrentUserSubscription = &organizer_dto.SubscriptionInfo{
					IsSubscribed:           true,
					SubscriptionDate:       subscription.SubscriptionDate,
					ReceiveEventUpdates:    subscription.ReceiveEventUpdates,
					ReceiveNewsletters:     subscription.ReceiveNewsletters,
					ReceivePromotions:      subscription.ReceivePromotions,
					NotificationPreference: subscription.NotificationTypes,
				}
			} else {
				resp.CurrentUserSubscription = &organizer_dto.SubscriptionInfo{
					IsSubscribed:           false,
					SubscriptionDate:       time.Time{},
					ReceiveEventUpdates:    false,
					ReceiveNewsletters:     false,
					ReceivePromotions:      false,
					NotificationPreference: "",
				}
			}
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
