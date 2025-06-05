package organizer_dto

import "time"

type OrganizerResponse struct {
	ID                      string            `json:"id"`
	Name                    string            `json:"name"`
	ContactPerson           string            `json:"contact_person"`
	Email                   string            `json:"email"`
	Phone                   string            `json:"phone,omitempty"`
	CompanyName             string            `json:"company_name,omitempty"`
	TaxID                   string            `json:"tax_id,omitempty"`
	BankAccountInfo         string            `json:"bank_account_info,omitempty"`
	ImageURL                string            `json:"image_url,omitempty"`
	CommissionRate          float64           `json:"commission_rate"`
	Balance                 float64           `json:"balance"`
	Status                  string            `json:"status"`
	IsFlagged               bool              `json:"is_flagged"`
	IsBanned                bool              `json:"is_banned"`
	CreatedBy               string            `json:"created_by"`
	CreatedByUser           *UserResponse     `json:"created_by_user,omitempty"`
	SubscriberCount         int64             `json:"subscriber_count"`
	IsAcceptingSubscribers  bool              `json:"is_accepting_subscribers"`
	CurrentUserSubscription *SubscriptionInfo `json:"current_user_subscription,omitempty"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type SubscriptionInfo struct {
	IsSubscribed           bool      `json:"is_subscribed"`
	SubscriptionDate       time.Time `json:"subscription_date,omitempty"`
	ReceiveEventUpdates    bool      `json:"receive_event_updates"`
	ReceiveNewsletters     bool      `json:"receive_newsletters"`
	ReceivePromotions      bool      `json:"receive_promotions"`
	NotificationPreference string    `json:"notification_preference"`
}

type OrganizerSubscriptionInfo struct {
	OrganizerID  string           `json:"organizer_id"`
	Name         string           `json:"name"`
	ImageURL     string           `json:"image_url,omitempty"`
	SubscribedAt time.Time        `json:"subscribed_at"`
	Preferences  SubscriptionInfo `json:"preferences"`
}

type SubscriberInfo struct {
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	SubscribedAt time.Time `json:"subscribed_at"`
	IsBanned     bool      `json:"is_banned"`
	IsFlagged    bool      `json:"is_flagged"`
}

type CreateOrganizerRequest struct {
	Name            string  `json:"name" validate:"required,min=2,max=255"`
	ContactPerson   string  `json:"contact_person" validate:"required,min=2,max=255"`
	Email           string  `json:"email" validate:"required,email"`
	Phone           string  `json:"phone,omitempty" validate:"max=50"`
	CompanyName     string  `json:"company_name,omitempty" validate:"max=255"`
	TaxID           string  `json:"tax_id,omitempty" validate:"max=100"`
	BankAccountInfo string  `json:"bank_account_info,omitempty"`
	ImageURL        string  `json:"image_url,omitempty" validate:"max=255"`
	CommissionRate  float64 `json:"commission_rate" validate:"gte=0,lte=100"`
	Balance         float64 `json:"balance" validate:"gte=0"`
	Notes           string  `json:"notes,omitempty"`
}

type UpdateOrganizerRequest struct {
	Name               string  `json:"name" validate:"required,min=2,max=255"`
	ContactPerson      string  `json:"contact_person" validate:"required,min=2,max=255"`
	Email              string  `json:"email" validate:"required,email"`
	Phone              string  `json:"phone,omitempty" validate:"max=50"`
	CompanyName        string  `json:"company_name,omitempty" validate:"max=255"`
	TaxID              string  `json:"tax_id,omitempty" validate:"max=100"`
	BankAccountInfo    string  `json:"bank_account_info,omitempty"`
	CommissionRate     float64 `json:"commission_rate" validate:"gte=0,lte=100"`
	Balance            float64 `json:"balance" validate:"gte=0"`
	Notes              string  `json:"notes,omitempty"`
	AllowSubscriptions bool    `json:"allow_subscriptions" validate:"omitempty"`
}

type BasicOrganizerResponse struct {
	ID                      string            `json:"id"`
	Name                    string            `json:"name"`
	CompanyName             string            `json:"company_name,omitempty"`
	ImageURL                string            `json:"image_url,omitempty"`
	IsActive                bool              `json:"is_active"`
	IsAcceptingSubscribers  bool              `json:"is_accepting_subscribers"`
	CreatedAt               time.Time         `json:"created_at"`
	CurrentUserSubscription *SubscriptionInfo `json:"current_user_subscription,omitempty"`
}
