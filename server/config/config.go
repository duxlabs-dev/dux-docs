package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                   string
	Environment            string
	GithubToken            string
	KlipyAPIKey            string
	PoeditorCallbackSecret string
	GithubWebhookSecret    string
	GithubModToken         string
	GithubAppID            int64
	GithubAppPrivateKey    string
	ModOrg                 string
	ModTeam                string
	OwnersTeam             string
	DiscordWebhookURL      string
	UploadToken            string
	UploadDir              string
	CacheDir               string
	BlogDir                string
	PublicBaseURL          string
	FedoraMessagingEnabled bool
	FedoraMessagingURL     string
	FedoraMessagingTLSDir  string
	QtRebuildRepo          string
	QtRebuildPackages      []string
	KofiVerificationToken  string
	SponsorOwners          []string
}

func NewConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8337"
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	klipyAPIKey := os.Getenv("KLIPY_API_KEY")
	poeditorSecret := os.Getenv("POEDITOR_CALLBACK_SECRET")
	githubWebhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	githubModToken := os.Getenv("GITHUB_MOD_TOKEN")
	githubAppID, _ := strconv.ParseInt(os.Getenv("GITHUB_APP_ID"), 10, 64)
	githubAppPrivateKey := os.Getenv("GITHUB_APP_PRIVATE_KEY")

	modOrg := os.Getenv("PLUGIN_MOD_ORG")
	if modOrg == "" {
		modOrg = "duxlabs-dev"
	}

	modTeam := os.Getenv("PLUGIN_MOD_TEAM")
	if modTeam == "" {
		modTeam = "plugin-moderators"
	}

	ownersTeam := os.Getenv("PLUGIN_OWNERS_TEAM")
	if ownersTeam == "" {
		ownersTeam = "owners"
	}

	discordWebhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	uploadToken := os.Getenv("UPLOAD_TOKEN")

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "/data/uploads"
	}

	cacheDir := os.Getenv("CACHE_DIR")

	blogDir := os.Getenv("BLOG_DIR")
	if blogDir == "" {
		blogDir = "../blog"
	}

	publicBaseURL := os.Getenv("PUBLIC_BASE_URL")
	if publicBaseURL == "" {
		publicBaseURL = "https://api.danklinux.com"
	}

	fedoraMessagingEnabled, _ := strconv.ParseBool(os.Getenv("FEDORA_MESSAGING_ENABLED"))
	fedoraMessagingURL := os.Getenv("FEDORA_MESSAGING_URL")
	fedoraMessagingTLSDir := os.Getenv("FEDORA_MESSAGING_TLS_DIR")

	qtRebuildRepo := os.Getenv("QT_REBUILD_REPO")
	if qtRebuildRepo == "" {
		qtRebuildRepo = "duxlabs-dev/dux-repo"
	}

	qtRebuildPackages := os.Getenv("QT_REBUILD_PACKAGES")
	if qtRebuildPackages == "" {
		qtRebuildPackages = "qt6-qtbase,qt6-qtdeclarative,qt6-qtwayland"
	}

	sponsorOwners := os.Getenv("SPONSOR_OWNERS")
	if sponsorOwners == "" {
		sponsorOwners = "AvengeMedia,duxlabs-dev"
	}

	return &Config{
		Port:                   port,
		Environment:            env,
		GithubToken:            githubToken,
		KlipyAPIKey:            klipyAPIKey,
		PoeditorCallbackSecret: poeditorSecret,
		GithubWebhookSecret:    githubWebhookSecret,
		GithubModToken:         githubModToken,
		GithubAppID:            githubAppID,
		GithubAppPrivateKey:    githubAppPrivateKey,
		ModOrg:                 modOrg,
		ModTeam:                modTeam,
		OwnersTeam:             ownersTeam,
		DiscordWebhookURL:      discordWebhookURL,
		UploadToken:            uploadToken,
		UploadDir:              uploadDir,
		CacheDir:               cacheDir,
		BlogDir:                blogDir,
		PublicBaseURL:          publicBaseURL,
		FedoraMessagingEnabled: fedoraMessagingEnabled,
		FedoraMessagingURL:     fedoraMessagingURL,
		FedoraMessagingTLSDir:  fedoraMessagingTLSDir,
		QtRebuildRepo:          qtRebuildRepo,
		QtRebuildPackages:      strings.Split(qtRebuildPackages, ","),
		KofiVerificationToken:  os.Getenv("KOFI_VERIFICATION_TOKEN"),
		SponsorOwners:          strings.Split(sponsorOwners, ","),
	}
}
