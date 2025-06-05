package organizers

import (
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Organizer struct {
	ID              string  `gorm:"type:char(36);primaryKey" json:"id"`
	Name            string  `gorm:"type:varchar(255);not null" json:"name"`
	ContactPerson   string  `gorm:"type:varchar(255);not null" json:"contact_person"`
	Email           string  `gorm:"type:varchar(255);not null" json:"email"`
	Phone           string  `gorm:"type:varchar(50)" json:"phone,omitempty"`
	CompanyName     string  `gorm:"type:varchar(255)" json:"company_name,omitempty"`
	TaxID           string  `gorm:"type:varchar(100)" json:"tax_id,omitempty"`
	BankAccountInfo string  `gorm:"type:text" json:"bank_account_info,omitempty"`
	ImageURL        string  `gorm:"type:varchar(255)" json:"image_url,omitempty"`
	CommissionRate  float64 `gorm:"type:numeric(5,2);default:10.00" json:"commission_rate"`
	Balance         float64 `gorm:"type:numeric(14,2);default:0.00" json:"balance"`
	Status          string  `gorm:"type:varchar(20);default:'active';check:status IN ('active','inactive','suspended')" json:"status"`
	IsFlagged       bool    `gorm:"default:false" json:"is_flagged"`
	IsBanned        bool    `gorm:"default:false" json:"is_banned"`
	Notes           string  `gorm:"type:text" json:"notes,omitempty"`

	AllowSubscriptions       bool  `gorm:"default:true" json:"allow_subscriptions"`
	SubscriberCount          int64 `gorm:"default:0" json:"subscriber_count"`
	IsAcceptingSubscriptions bool  `gorm:"default:true" json:"is_accepting_subscriptions"`

	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy     string         `gorm:"type:char(36);index" json:"created_by"`
	CreatedByUser members.User   `gorm:"foreignKey:CreatedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"created_by_user"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate assigns a UUID if not set
func (o *Organizer) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

func (Organizer) TableName() string {
	return "organizers"
}
