package models

import "time"

type CodeContributor struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	URL       string `json:"url"`
	// Contributions is the commit count summed across Repos.
	Contributions int      `json:"contributions"`
	Repos         []string `json:"repos"`
}

type TranslatorLanguage struct {
	ProjectID string `json:"projectId"`
	Project   string `json:"project"`
	Code      string `json:"code"`
	Name      string `json:"name"`
}

// Translator is a POEditor contributor. POEditor records who is assigned to a
// language, not who translated what, so this is membership, not activity.
type Translator struct {
	Name        string               `json:"name"`
	Proofreader bool                 `json:"proofreader,omitempty"`
	Languages   []TranslatorLanguage `json:"languages"`
}

type ContributorsFeed struct {
	Code        []CodeContributor `json:"code"`
	Translators []Translator      `json:"translators"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}
