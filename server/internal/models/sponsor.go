package models

import "time"

type Sponsor struct {
	Name      string `json:"name"`
	Login     string `json:"login,omitempty"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	URL       string `json:"url,omitempty"`
	// Source is "github" or "kofi".
	Source string `json:"source"`
}

type SponsorsFeed struct {
	Current   []Sponsor `json:"current"`
	Past      []Sponsor `json:"past"`
	UpdatedAt time.Time `json:"updatedAt"`
}
