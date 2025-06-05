package organizers

import (
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationSubscription struct {
	ID           string       `gorm:"type:char(36);primaryKey" json:"id"`
	OrganizerID  string       `gorm:"type:char(36);not null;index" json:"organizer_id"`
	Organizer    Organizer    `gorm:"foreignKey:OrganizerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"organizer"`
	SubscriberID string       `gorm:"type:char(36);not null;index" json:"subscriber_id"`
	Subscriber   members.User `gorm:"foreignKey:SubscriberID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"subscriber"`

	// Subscription preferences
	ReceiveEventUpdates bool   `gorm:"default:true" json:"receive_event_updates"`
	ReceiveNewsletters  bool   `gorm:"default:true" json:"receive_newsletters"`
	ReceivePromotions   bool   `gorm:"default:false" json:"receive_promotions"`
	NotificationTypes   string `gorm:"type:varchar(50);default:'all';check:notification_types IN ('all','essential','none')" json:"notification_types"`

	// Status fields
	IsActive        bool   `gorm:"default:true" json:"is_active"`
	IsBlocked       bool   `gorm:"default:false" json:"is_blocked"` 
	BlockedReason   string `gorm:"type:varchar(255)" json:"blocked_reason,omitempty"` 
	SubscriberNotes string `gorm:"type:text" json:"subscriber_notes,omitempty"` 

	// Metadata
	SubscriptionDate time.Time      `gorm:"autoCreateTime" json:"subscription_date"`
	LastUpdated      time.Time      `gorm:"autoUpdateTime" json:"last_updated"`
	UnsubscribedAt   gorm.DeletedAt `gorm:"index" json:"unsubscribed_at,omitempty"`
}

// BeforeCreate assigns a UUID if not set
func (os *OrganizationSubscription) BeforeCreate(tx *gorm.DB) (err error) {
	if os.ID == "" {
		os.ID = uuid.New().String()
	}
	return nil
}

func (OrganizationSubscription) TableName() string {
	return "organization_subscriptions"
}
