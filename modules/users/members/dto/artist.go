package dto

import (
	"encoding/json"
	"strings"
	"time"
)

// CommaSeparatedString handles JSON input as either a single string, comma-separated string, or array of strings
type CommaSeparatedString string

// UnmarshalJSON implements custom JSON unmarshaling
func (c *CommaSeparatedString) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*c = CommaSeparatedString(single)
		return nil
	}

	var array []string
	if err := json.Unmarshal(data, &array); err == nil {
		*c = CommaSeparatedString(strings.Join(array, ", "))
		return nil
	}

	return json.Unmarshal(data, (*string)(c))
}

// ToArray converts the comma-separated string to an array
func (c CommaSeparatedString) ToArray() []string {
	if c == "" {
		return nil
	}
	return strings.Split(string(c), ", ")
}

// CreateArtistProfileDTO represents the data required to create a new artist profile
type CreateArtistProfileDTO struct {
	StageName      string               `json:"stage_name" example:"DJ Wave" validate:"required,max=100"`
	Type           string               `json:"type" example:"musician" validate:"required,max=50"`
	Bio            string               `json:"bio" example:"Emerging EDM artist known for high-energy live shows." validate:"max=65535"`
	Website        string               `json:"website,omitempty" example:"https://djwave.com" validate:"omitempty,url,max=255"`
	Location       string               `json:"location,omitempty" example:"Berlin, Germany" validate:"omitempty,max=100"`
	Collaboration  bool                 `json:"open_to_collaboration" example:"true"`
	SpotifyURL     string               `json:"spotify_url,omitempty" example:"https://spotify.com/artist/123" validate:"omitempty,url,max=255"`
	YouTubeURL     string               `json:"youtube_url,omitempty" example:"https://youtube.com/channel/abc" validate:"omitempty,url,max=255"`
	Instagram      string               `json:"instagram_url,omitempty" example:"https://instagram.com/djwave" validate:"omitempty,url,max=255"`
	TikTok         string               `json:"tiktok_url,omitempty" example:"https://tiktok.com/@djwave" validate:"omitempty,url,max=255"`
	Twitter        string               `json:"twitter_url,omitempty" example:"https://twitter.com/djwave" validate:"omitempty,url,max=255"`
	Reddit         string               `json:"reddit_url,omitempty" example:"https://reddit.com/u/djwave" validate:"omitempty,url,max=255"`
	Snapchat       string               `json:"snapchat_url,omitempty" example:"https://snapchat.com/add/djwave" validate:"omitempty,url,max=255"`
	Patreon        string               `json:"patreon_url,omitempty" example:"https://patreon.com/djwave" validate:"omitempty,url,max=255"`
	SoundCloud     string               `json:"soundcloud_url,omitempty" example:"https://soundcloud.com/djwave" validate:"omitempty,url,max=255"`
	Behance        string               `json:"behance_url,omitempty" example:"https://behance.net/djwave" validate:"omitempty,url,max=255"`
	Dribbble       string               `json:"dribbble_url,omitempty" example:"https://dribbble.com/djwave" validate:"omitempty,url,max=255"`
	Vimeo          string               `json:"vimeo_url,omitempty" example:"https://vimeo.com/djwave" validate:"omitempty,url,max=255"`
	Goodreads      string               `json:"goodreads_url,omitempty" example:"https://goodreads.com/djwave" validate:"omitempty,url,max=255"`
	LinkedIn       string               `json:"linkedin_url,omitempty" example:"https://linkedin.com/in/djwave" validate:"omitempty,url,max=255"`
	Pinterest      string               `json:"pinterest_url,omitempty" example:"https://pinterest.com/djwave" validate:"omitempty,url,max=255"`
	Twitch         string               `json:"twitch_url,omitempty" example:"https://twitch.tv/djwave" validate:"omitempty,url,max=255"`
	DeviantArt     string               `json:"deviantart_url,omitempty" example:"https://deviantart.com/djwave" validate:"omitempty,url,max=255"`
	PortfolioURL   string               `json:"portfolio_url,omitempty" example:"https://djwave.com/portfolio" validate:"omitempty,url,max=255"`
	Genres         CommaSeparatedString `json:"genres,omitempty" example:"EDM, Techno" validate:"omitempty,max=65535"`
	Skills         CommaSeparatedString `json:"skills,omitempty" example:"DJing, Music Production" validate:"omitempty,max=65535"`
	Availability   string               `json:"availability,omitempty" example:"Weekends and evenings" validate:"omitempty,max=100"`
	ContactEmail   string               `json:"contact_email,omitempty" example:"contact@djwave.com" validate:"omitempty,email,max=255"`
	Representation string               `json:"representation,omitempty" example:"Wave Talent Agency" validate:"omitempty,max=255"`
}

// UpdateArtistProfileDTO represents the data for updating an existing artist profile
type UpdateArtistProfileDTO struct {
	StageName      string               `json:"stage_name,omitempty" example:"DJ Wave Updated" validate:"omitempty,max=100"`
	Type           string               `json:"type,omitempty" example:"producer" validate:"omitempty,max=50"`
	Bio            string               `json:"bio,omitempty" example:"Updated bio with new achievements." validate:"omitempty,max=65535"`
	Website        string               `json:"website,omitempty" example:"https://djwaveupdated.com" validate:"omitempty,url,max=255"`
	Location       string               `json:"location,omitempty" example:"London, UK" validate:"omitempty,max=100"`
	Collaboration  *bool                `json:"open_to_collaboration,omitempty" example:"false"`
	SpotifyURL     string               `json:"spotify_url,omitempty" example:"https://spotify.com/artist/456" validate:"omitempty,url,max=255"`
	YouTubeURL     string               `json:"youtube_url,omitempty" example:"https://youtube.com/channel/def" validate:"omitempty,url,max=255"`
	Instagram      string               `json:"instagram_url,omitempty" example:"https://instagram.com/djwaveupdated" validate:"omitempty,url,max=255"`
	TikTok         string               `json:"tiktok_url,omitempty" example:"https://tiktok.com/@djwaveupdated" validate:"omitempty,url,max=255"`
	Twitter        string               `json:"twitter_url,omitempty" example:"https://twitter.com/djwaveupdated" validate:"omitempty,url,max=255"`
	Reddit         string               `json:"reddit_url,omitempty" example:"https://reddit.com/u/djwaveupdated" validate:"omitempty,url,max=255"`
	Snapchat       string               `json:"snapchat_url,omitempty" example:"https://snapchat.com/add/djwaveupdated" validate:"omitempty,url,max=255"`
	Patreon        string               `json:"patreon_url,omitempty" example:"https://patreon.com/djwaveupdated" validate:"omitempty,url,max=255"`
	SoundCloud     string               `json:"soundcloud_url,omitempty" example:"https://soundcloud.com/djwaveupdated" validate:"omitempty,url,max=255"`
	Behance        string               `json:"behance_url,omitempty" example:"https://behance.net/djwaveupdated" validate:"omitempty,url,max=255"`
	Dribbble       string               `json:"dribbble_url,omitempty" example:"https://dribbble.com/djwaveupdated" validate:"omitempty,url,max=255"`
	Vimeo          string               `json:"vimeo_url,omitempty" example:"https://vimeo.com/djwaveupdated" validate:"omitempty,url,max=255"`
	Goodreads      string               `json:"goodreads_url,omitempty" example:"https://goodreads.com/djwaveupdated" validate:"omitempty,url,max=255"`
	LinkedIn       string               `json:"linkedin_url,omitempty" example:"https://linkedin.com/in/djwaveupdated" validate:"omitempty,url,max=255"`
	Pinterest      string               `json:"pinterest_url,omitempty" example:"https://pinterest.com/djwaveupdated" validate:"omitempty,url,max=255"`
	Twitch         string               `json:"twitch_url,omitempty" example:"https://twitch.tv/djwaveupdated" validate:"omitempty,url,max=255"`
	DeviantArt     string               `json:"deviantart_url,omitempty" example:"https://deviantart.com/djwaveupdated" validate:"omitempty,url,max=255"`
	PortfolioURL   string               `json:"portfolio_url,omitempty" example:"https://djwaveupdated.com/portfolio" validate:"omitempty,url,max=255"`
	Genres         CommaSeparatedString `json:"genres,omitempty" example:"House, Trance" validate:"omitempty,max=65535"`
	Skills         CommaSeparatedString `json:"skills,omitempty" example:"Mixing, Mastering" validate:"omitempty,max=65535"`
	Availability   string               `json:"availability,omitempty" example:"Full-time" validate:"omitempty,max=100"`
	ContactEmail   string               `json:"contact_email,omitempty" example:"updated@djwave.com" validate:"omitempty,email,max=255"`
	Representation string               `json:"representation,omitempty" example:"Updated Talent Agency" validate:"omitempty,max=255"`
}

// ReadArtistProfileDTO represents the response data for reading an artist profile
type ReadArtistProfileDTO struct {
	ID             string     `json:"id" example:"123e4567-e89b-12d3-a456-426614174001"`
	UserID         string     `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	StageName      string     `json:"stage_name" example:"DJ Wave"`
	Type           string     `json:"type" example:"musician"`
	Bio            string     `json:"bio" example:"Emerging EDM artist known for high-energy live shows."`
	Website        string     `json:"website,omitempty" example:"https://djwave.com"`
	Location       string     `json:"location,omitempty" example:"Berlin, Germany"`
	Collaboration  bool       `json:"open_to_collaboration" example:"true"`
	SpotifyURL     string     `json:"spotify_url,omitempty" example:"https://spotify.com/artist/123"`
	YouTubeURL     string     `json:"youtube_url,omitempty" example:"https://youtube.com/channel/abc"`
	Instagram      string     `json:"instagram_url,omitempty" example:"https://instagram.com/djwave"`
	TikTok         string     `json:"tiktok_url,omitempty" example:"https://tiktok.com/@djwave"`
	Twitter        string     `json:"twitter_url,omitempty" example:"https://twitter.com/djwave"`
	Reddit         string     `json:"reddit_url,omitempty" example:"https://reddit.com/u/djwave"`
	Snapchat       string     `json:"snapchat_url,omitempty" example:"https://snapchat.com/add/djwave"`
	Patreon        string     `json:"patreon_url,omitempty" example:"https://patreon.com/djwave"`
	SoundCloud     string     `json:"soundcloud_url,omitempty" example:"https://soundcloud.com/djwave"`
	Behance        string     `json:"behance_url,omitempty" example:"https://behance.net/djwave"`
	Dribbble       string     `json:"dribbble_url,omitempty" example:"https://dribbble.com/djwave"`
	Vimeo          string     `json:"vimeo_url,omitempty" example:"https://vimeo.com/djwave"`
	Goodreads      string     `json:"goodreads_url,omitempty" example:"https://goodreads.com/djwave"`
	LinkedIn       string     `json:"linkedin_url,omitempty" example:"https://linkedin.com/in/djwave"`
	Pinterest      string     `json:"pinterest_url,omitempty" example:"https://pinterest.com/djwave"`
	Twitch         string     `json:"twitch_url,omitempty" example:"https://twitch.tv/djwave"`
	DeviantArt     string     `json:"deviantart_url,omitempty" example:"https://deviantart.com/djwave"`
	PortfolioURL   string     `json:"portfolio_url,omitempty" example:"https://djwave.com/portfolio"`
	Genres         string     `json:"genres,omitempty" example:"EDM, Techno"`
	GenresArray    []string   `json:"genres_array,omitempty" gorm:"-" example:"[\"EDM\",\"Techno\"]"`
	Skills         string     `json:"skills,omitempty" example:"DJing, Music Production"`
	SkillsArray    []string   `json:"skills_array,omitempty" gorm:"-" example:"[\"DJing\",\"Music Production\"]"`
	Availability   string     `json:"availability,omitempty" example:"Weekends and evenings"`
	ContactEmail   string     `json:"contact_email,omitempty" example:"contact@djwave.com"`
	Representation string     `json:"representation,omitempty" example:"Wave Talent Agency"`
	CreatedAt      time.Time  `json:"created_at" example:"2025-06-08T23:46:39Z"`
	UpdatedAt      time.Time  `json:"updated_at" example:"2025-06-08T23:46:39Z"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" example:"null"`
}

type PublicArtistProfileDto struct {
	StageName     string   `json:"stage_name"`
	Type          string   `json:"type"`
	Bio           string   `json:"bio,omitempty"`
	Website       string   `json:"website,omitempty"`
	Location      string   `json:"location,omitempty"`
	Collaboration bool     `json:"open_to_collaboration,omitempty"`
	SpotifyURL    string   `json:"spotify_url,omitempty"`
	YouTubeURL    string   `json:"youtube_url,omitempty"`
	Instagram     string   `json:"instagram_url,omitempty"`
	TikTok        string   `json:"tiktok_url,omitempty"`
	Twitter       string   `json:"twitter_url,omitempty"`
	PortfolioURL  string   `json:"portfolio_url,omitempty"`
	Genres        []string `json:"genres,omitempty"`
	Skills        []string `json:"skills,omitempty"`
	Availability  string   `json:"availability,omitempty"`
}
