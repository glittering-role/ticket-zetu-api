package artist

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ArtistType string

type ArtistProfile struct {
	ID     string `gorm:"type:char(36);primaryKey" json:"id"`
	UserID string `gorm:"type:char(36);uniqueIndex" json:"user_id"`

	// Core artist info
	StageName     string     `gorm:"size:100;index" json:"stage_name"`
	Type          ArtistType `gorm:"type:varchar(50);index" json:"type"`
	Bio           string     `gorm:"type:text" json:"bio"`
	Website       string     `gorm:"size:255" json:"website,omitempty"`
	Location      string     `gorm:"size:100" json:"location,omitempty"`
	Collaboration bool       `gorm:"default:false" json:"open_to_collaboration"`

	// Expanded social media links
	SpotifyURL string `gorm:"size:255" json:"spotify_url,omitempty"`
	YouTubeURL string `gorm:"size:255" json:"youtube_url,omitempty"`
	Instagram  string `gorm:"size:255" json:"instagram_url,omitempty"`
	TikTok     string `gorm:"size:255" json:"tiktok_url,omitempty"`
	Twitter    string `gorm:"size:255" json:"twitter_url,omitempty"`
	Reddit     string `gorm:"size:255" json:"reddit_url,omitempty"`
	Snapchat   string `gorm:"size:255" json:"snapchat_url,omitempty"`
	Patreon    string `gorm:"size:255" json:"patreon_url,omitempty"`
	SoundCloud string `gorm:"size:255" json:"soundcloud_url,omitempty"`
	Behance    string `gorm:"size:255" json:"behance_url,omitempty"`
	Dribbble   string `gorm:"size:255" json:"dribbble_url,omitempty"`
	Vimeo      string `gorm:"size:255" json:"vimeo_url,omitempty"`
	Goodreads  string `gorm:"size:255" json:"goodreads_url,omitempty"`
	LinkedIn   string `gorm:"size:255" json:"linkedin_url,omitempty"`
	Pinterest  string `gorm:"size:255" json:"pinterest_url,omitempty"`
	Twitch     string `gorm:"size:255" json:"twitch_url,omitempty"`
	DeviantArt string `gorm:"size:255" json:"deviantart_url,omitempty"`

	// Portfolio and media
	PortfolioURL string `gorm:"size:255" json:"portfolio_url,omitempty"`
	Genres       string `gorm:"type:text" json:"genres,omitempty"`
	Skills       string `gorm:"type:text" json:"skills,omitempty"`

	// Professional details
	Availability   string `gorm:"size:100" json:"availability,omitempty"`
	ContactEmail   string `gorm:"size:255" json:"contact_email,omitempty"`
	Representation string `gorm:"size:255" json:"representation,omitempty"`

	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (a *ArtistProfile) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

func (ArtistProfile) TableName() string {
	return "artist_profiles"
}
