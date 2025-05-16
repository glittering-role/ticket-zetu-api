package members

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type AuthType string

const (
	AuthTypePassword AuthType = "password"
	AuthTypeOAuth    AuthType = "oauth"
	AuthTypeSSO      AuthType = "sso"
)

type UserSecurityAttributes struct {
	ID                       uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	UserID                   uuid.UUID      `gorm:"type:char(36);uniqueIndex" json:"user_id"`
	Password                 string         `gorm:"type:text" json:"-"` // Always exclude from JSON
	AuthType                 AuthType       `gorm:"type:varchar(20);default:'password'" json:"auth_type"`
	PasswordResetToken       string         `gorm:"type:text" json:"-"`
	PasswordResetTokenExpiry *time.Time     `gorm:"type:timestamp" json:"-"`
	EmailVerificationToken   string         `gorm:"type:text" json:"-"`
	EmailTokenExpiry         *time.Time     `gorm:"type:timestamp" json:"-"`
	TwoFactorEnabled         bool           `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret          string         `gorm:"type:text" json:"-"`
	FailedLoginAttempts      int            `gorm:"default:0" json:"failed_login_attempts"`
	LockUntil                *time.Time     `gorm:"type:timestamp" json:"lock_until"`
	IsDeleted                bool           `gorm:"default:false" json:"is_deleted"`
	CreatedAt                time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

func (a *UserSecurityAttributes) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (UserSecurityAttributes) TableName() string {
	return "user_security_attributes"
}

// Helper Methods
func (a *UserSecurityAttributes) IsLocked() bool {
	if a.LockUntil == nil {
		return false
	}
	return a.LockUntil.After(time.Now())
}

func (a *UserSecurityAttributes) CanAttemptLogin() bool {
	return !a.IsLocked() && !a.IsDeleted
}

func (a *UserSecurityAttributes) NeedsEmailVerification() bool {
	return a.EmailVerificationToken != "" &&
		(a.EmailTokenExpiry == nil || a.EmailTokenExpiry.After(time.Now()))
}

// Added security helper methods
func (a *UserSecurityAttributes) ResetLoginAttempts() {
	a.FailedLoginAttempts = 0
	a.LockUntil = nil
}

func (a *UserSecurityAttributes) IncrementFailedAttempts(maxAttempts int) {
	a.FailedLoginAttempts++
	if a.FailedLoginAttempts >= maxAttempts {
		lockDuration := time.Duration(a.FailedLoginAttempts-maxAttempts+1) * time.Hour
		lockTime := time.Now().Add(lockDuration)
		a.LockUntil = &lockTime
	}
}
