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

// TranslationLanguage counts the POEditor contributors assigned to a language
// across every project. POEditor records assignment, not activity, and no
// names are published: translators have not agreed to be listed.
type TranslationLanguage struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Translators int    `json:"translators"`
}

type ContributorsFeed struct {
	Code      []CodeContributor     `json:"code"`
	Languages []TranslationLanguage `json:"languages"`
	// TranslatorCount is the number of distinct people across Languages.
	TranslatorCount int       `json:"translatorCount"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
