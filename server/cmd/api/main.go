package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/duxlabs-dev/dux-docs/server/config"
	gifs_handler "github.com/duxlabs-dev/dux-docs/server/internal/api/handlers/gifs"
	plugins_handler "github.com/duxlabs-dev/dux-docs/server/internal/api/handlers/plugins"
	"github.com/duxlabs-dev/dux-docs/server/internal/api/handlers/poeditor"
	previews_handler "github.com/duxlabs-dev/dux-docs/server/internal/api/handlers/previews"
	releases_handler "github.com/duxlabs-dev/dux-docs/server/internal/api/handlers/releases"
	sponsors_handler "github.com/duxlabs-dev/dux-docs/server/internal/api/handlers/sponsors"
	stickers_handler "github.com/duxlabs-dev/dux-docs/server/internal/api/handlers/stickers"
	themes_handler "github.com/duxlabs-dev/dux-docs/server/internal/api/handlers/themes"
	uploads_handler "github.com/duxlabs-dev/dux-docs/server/internal/api/handlers/uploads"
	"github.com/duxlabs-dev/dux-docs/server/internal/api/handlers/webhooks"
	"github.com/duxlabs-dev/dux-docs/server/internal/api/middleware"
	"github.com/duxlabs-dev/dux-docs/server/internal/api/server"
	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/fedoramessaging"
	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/github"
	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/githubapp"
	"github.com/duxlabs-dev/dux-docs/server/internal/integrations/klipy"
	"github.com/duxlabs-dev/dux-docs/server/internal/log"
	"github.com/duxlabs-dev/dux-docs/server/internal/services/previews"
	"github.com/duxlabs-dev/dux-docs/server/internal/services/qtrebuild"
	"github.com/duxlabs-dev/dux-docs/server/internal/services/registry"
	"github.com/duxlabs-dev/dux-docs/server/internal/services/releases"
	"github.com/duxlabs-dev/dux-docs/server/internal/services/sponsors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"
)

var Version = "development"

func NewHumaConfig(title, version string) huma.Config {
	schemaPrefix := "#/components/schemas/"
	schemasPath := "/schemas"

	registry := huma.NewMapRegistry(schemaPrefix, huma.DefaultSchemaNamer)

	cfg := huma.Config{
		OpenAPI: &huma.OpenAPI{
			OpenAPI: "3.1.0",
			Info: &huma.Info{
				Title:   title,
				Version: version,
			},
			Components: &huma.Components{
				Schemas: registry,
			},
		},
		OpenAPIPath:   "/openapi",
		DocsPath:      "/docs",
		SchemasPath:   schemasPath,
		Formats:       huma.DefaultFormats,
		DefaultFormat: "application/json",
	}

	return cfg
}

func startAPI(cfg *config.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signalCh
		slog.Info("Received shutdown signal", "signal", sig)
		cancel()
	}()

	var pluginCacheFile, themeCacheFile, releasesCacheFile string
	if cfg.CacheDir != "" {
		pluginCacheFile = filepath.Join(cfg.CacheDir, "plugins.json")
		themeCacheFile = filepath.Join(cfg.CacheDir, "themes.json")
		releasesCacheFile = filepath.Join(cfg.CacheDir, "releases.json")
	}
	pluginCache := registry.NewCache(cfg.GithubToken, pluginCacheFile)
	themeCache := registry.NewThemeCache(cfg.GithubToken, themeCacheFile)
	releasesCache := releases.NewCache(cfg.GithubToken, cfg.BlogDir, releasesCacheFile)
	var sponsorsGQL sponsors.GraphQL
	if cfg.GithubToken != "" {
		sponsorsGQL = github.NewClient(cfg.GithubToken)
	}
	sponsorsCache := sponsors.NewCache(sponsorsGQL, cfg.SponsorOwners, cfg.CacheDir)

	var previewGen *previews.Generator
	if cfg.CacheDir != "" {
		gen, err := previews.NewGenerator(cfg.CacheDir, cfg.PublicBaseURL)
		if err != nil {
			log.Error("Failed to initialize preview generator", "err", err)
		} else {
			previewGen = gen
			pluginCache.SetPreviewSyncer(gen)
			log.Info("Preview generator initialized")
		}
	}

	srvImpl := &server.Server{
		PluginCache:   pluginCache,
		ThemeCache:    themeCache,
		ReleasesCache: releasesCache,
	}

	go func() {
		log.Info("Initializing plugin cache...")
		if err := pluginCache.Initialize(ctx); err != nil {
			log.Error("Failed to initialize plugin cache", "err", err)
		}
	}()

	go func() {
		log.Info("Initializing theme cache...")
		if err := themeCache.Initialize(ctx); err != nil {
			log.Error("Failed to initialize theme cache", "err", err)
		}
	}()

	go func() {
		log.Info("Initializing releases cache...")
		if err := releasesCache.Initialize(ctx); err != nil {
			log.Error("Failed to initialize releases cache", "err", err)
		}
	}()

	go func() {
		log.Info("Initializing sponsors cache...")
		if err := sponsorsCache.Refresh(ctx); err != nil {
			log.Error("Failed to initialize sponsors cache", "err", err)
		}
	}()

	var klipyClient *klipy.Client
	if cfg.KlipyAPIKey != "" {
		klipyClient = klipy.NewClient(cfg.KlipyAPIKey)
		log.Info("Klipy client initialized")
	}

	gifRateLimiter := middleware.NewRateLimiter(100.0/60.0, 100)
	// Every API route gets a per-client ceiling; the Klipy-backed routes keep their tighter one on top.
	apiRateLimiter := middleware.NewRateLimiter(20, 100)

	r := chi.NewRouter()

	var allowedOrigins []string
	if cfg.Environment == "development" {
		allowedOrigins = []string{
			"http://localhost:3000",
			"https://duxlabs.dev",
		}
	} else {
		allowedOrigins = []string{
			"https://duxlabs.dev",
			"https://danklinux.com",
		}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"version": "` + Version + `"}`))
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Get("/uploads/{filename}", func(w http.ResponseWriter, r *http.Request) {
		uploads_handler.ServeFile(cfg.UploadDir, chi.URLParam(r, "filename"), w, r)
	})

	if previewGen != nil {
		servePreview := func(w http.ResponseWriter, r *http.Request) {
			previews_handler.ServePreview(previewGen.Store(), chi.URLParam(r, "pluginId"), w, r)
		}
		serveThumb := func(w http.ResponseWriter, r *http.Request) {
			previews_handler.ServeThumb(previewGen.ThumbStore(), chi.URLParam(r, "pluginId"), w, r)
		}
		r.Get("/previews/{pluginId}", servePreview)
		r.Head("/previews/{pluginId}", servePreview)
		r.Get("/previews/{pluginId}/thumb", serveThumb)
		r.Head("/previews/{pluginId}/thumb", serveThumb)
	}

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RealIP)
		r.Use(apiRateLimiter.Middleware)
		r.Use(middleware.Logger)
		r.Use(middleware.RequestID)

		humaConfig := NewHumaConfig("DankLinux Docs API", "1.0.0")
		humaConfig.DocsPath = ""
		api := humachi.New(r, humaConfig)

		r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<!doctype html>
<html>
	<head>
		<title>API Reference</title>
		<meta charset="utf-8" />
		<meta
			name="viewport"
			content="width=device-width, initial-scale=1" />
	</head>
	<body>
		<script
			id="api-reference"
			data-url="/openapi.json"></script>
		<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@latest"></script>
	</body>
</html>`))
		})

		pluginsGroup := huma.NewGroup(api, "/plugins")
		pluginsGroup.UseModifier(func(op *huma.Operation, next func(*huma.Operation)) {
			op.Tags = []string{"Plugins"}
			next(op)
		})
		plugins_handler.RegisterHandlers(srvImpl, pluginsGroup)

		themesGroup := huma.NewGroup(api, "/themes")
		themesGroup.UseModifier(func(op *huma.Operation, next func(*huma.Operation)) {
			op.Tags = []string{"Themes"}
			next(op)
		})
		themes_handler.RegisterHandlers(srvImpl, themesGroup)

		releasesGroup := huma.NewGroup(api, "/dms/releases")
		releasesGroup.UseSimpleModifier(func(op *huma.Operation) {
			op.Tags = []string{"Releases"}
		})
		releases_handler.RegisterHandlers(srvImpl, releasesGroup)

		sponsorsGroup := huma.NewGroup(api, "/sponsors")
		sponsorsGroup.UseSimpleModifier(func(op *huma.Operation) {
			op.Tags = []string{"Sponsors"}
		})
		sponsors_handler.RegisterHandlers(sponsorsCache, sponsorsGroup)

		if cfg.KofiVerificationToken != "" {
			r.Post("/webhooks/kofi", sponsors_handler.KofiWebhook(cfg.KofiVerificationToken, sponsorsCache))
		} else {
			log.Warn("KOFI_VERIFICATION_TOKEN not set; Ko-fi webhook disabled")
		}

		gifsGroup := huma.NewGroup(api, "/gifs")
		gifsGroup.UseSimpleModifier(func(op *huma.Operation) {
			op.Tags = []string{"GIFs"}
		})
		gifsGroup.UseMiddleware(gifRateLimiter.HumaMiddleware)
		gifs_handler.RegisterHandlers(srvImpl, klipyClient, gifsGroup)

		stickersGroup := huma.NewGroup(api, "/stickers")
		stickersGroup.UseSimpleModifier(func(op *huma.Operation) {
			op.Tags = []string{"Stickers"}
		})
		stickersGroup.UseMiddleware(gifRateLimiter.HumaMiddleware)
		stickers_handler.RegisterHandlers(srvImpl, klipyClient, stickersGroup)

		poeditorGroup := huma.NewGroup(api, "/poeditor")
		poeditorGroup.UseSimpleModifier(func(op *huma.Operation) {
			op.Tags = []string{"POEditor"}
		})
		poeditor.RegisterHandlers(cfg.PoeditorCallbackSecret, cfg.DiscordWebhookURL, poeditorGroup)

		uploadsGroup := huma.NewGroup(api, "/uploads")
		uploadsGroup.UseSimpleModifier(func(op *huma.Operation) {
			op.Tags = []string{"Uploads"}
		})
		uploads_handler.RegisterHandlers(cfg.UploadDir, cfg.UploadToken, uploadsGroup)

		webhooksGroup := huma.NewGroup(api, "/webhooks/github")
		webhooksGroup.UseSimpleModifier(func(op *huma.Operation) {
			op.Tags = []string{"Webhooks"}
		})
		const registryOwner, registryRepo = "duxlabs-dev", "dux-plugin-registry"
		var moderator webhooks.Moderator
		switch {
		case cfg.GithubAppID != 0 && cfg.GithubAppPrivateKey != "":
			appClient, err := githubapp.NewApp(cfg.GithubAppID, cfg.GithubAppPrivateKey, registryOwner, registryRepo)
			if err != nil {
				log.Error("Failed to init GitHub App moderator", "err", err)
			} else {
				moderator = appClient
			}
		case cfg.GithubModToken != "":
			moderator = githubapp.NewToken(cfg.GithubModToken, registryOwner, registryRepo)
		}
		webhooks.RegisterHandlers(webhooks.Config{
			Secret:     cfg.GithubWebhookSecret,
			Owner:      registryOwner,
			Repo:       registryRepo,
			Org:        cfg.ModOrg,
			Team:       cfg.ModTeam,
			OwnersTeam: cfg.OwnersTeam,
			Cache:      pluginCache,
			Moderator:  moderator,
			Authors:    pluginCache,
		}, webhooksGroup)
	})

	if cfg.FedoraMessagingEnabled {
		startQtRebuildWatcher(ctx, cfg)
	}

	addr := ":" + cfg.Port
	log.Infof("Starting server on %s", addr)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	scheduler, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		log.Fatal("Failed to create scheduler", "err", err)
	}

	_, err = scheduler.NewJob(
		gocron.DurationJob(10*time.Minute),
		gocron.NewTask(
			func(ctx context.Context) {
				log.Info("Running scheduled cache refresh")
				if err := pluginCache.Refresh(ctx); err != nil {
					log.Error("Failed to refresh plugin cache", "err", err)
				}
				if err := themeCache.Refresh(ctx); err != nil {
					log.Error("Failed to refresh theme cache", "err", err)
				}
			},
			ctx,
		),
	)
	if err != nil {
		log.Fatal("Failed to schedule cache refresh", "err", err)
	}

	// Releases change a few times a month; a slower tick keeps GitHub calls down.
	_, err = scheduler.NewJob(
		gocron.DurationJob(30*time.Minute),
		gocron.NewTask(
			func(ctx context.Context) {
				if err := releasesCache.Refresh(ctx); err != nil {
					log.Error("Failed to refresh releases cache", "err", err)
				}
			},
			ctx,
		),
	)
	if err != nil {
		log.Fatal("Failed to create cache refresh job", "err", err)
	}

	// Sponsorships change rarely; Ko-fi updates arrive by webhook in between.
	_, err = scheduler.NewJob(
		gocron.DurationJob(time.Hour),
		gocron.NewTask(
			func(ctx context.Context) {
				if err := sponsorsCache.Refresh(ctx); err != nil {
					log.Error("Failed to refresh sponsors cache", "err", err)
				}
			},
			ctx,
		),
	)
	if err != nil {
		log.Fatal("Failed to create sponsors refresh job", "err", err)
	}

	scheduler.Start()
	defer func() {
		if err := scheduler.Shutdown(); err != nil {
			log.Error("Scheduler shutdown error", "err", err)
		} else {
			log.Info("Scheduler gracefully stopped")
		}
	}()

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Info("Server gracefully stopped")
}

func newGithubDispatcher(cfg *config.Config, owner, repo string) (qtrebuild.Dispatcher, error) {
	switch {
	case cfg.GithubAppID != 0 && cfg.GithubAppPrivateKey != "":
		return githubapp.NewApp(cfg.GithubAppID, cfg.GithubAppPrivateKey, owner, repo)
	case cfg.GithubModToken != "":
		return githubapp.NewToken(cfg.GithubModToken, owner, repo), nil
	}
	return nil, nil
}

func startQtRebuildWatcher(ctx context.Context, cfg *config.Config) {
	owner, repo, ok := strings.Cut(cfg.QtRebuildRepo, "/")
	if !ok {
		log.Error("QT_REBUILD_REPO must be owner/repo", "value", cfg.QtRebuildRepo)
		return
	}

	dispatcher, err := newGithubDispatcher(cfg, owner, repo)
	if err != nil {
		log.Error("Failed to init GitHub dispatcher for qt rebuilds", "err", err)
		return
	}
	if dispatcher == nil {
		log.Warn("Fedora messaging enabled but no GitHub App or GITHUB_MOD_TOKEN configured, qt rebuild dispatch disabled")
		return
	}

	watcher := qtrebuild.NewWatcher(owner, repo, cfg.QtRebuildPackages, dispatcher)
	consumer, err := fedoramessaging.New(cfg.FedoraMessagingURL, cfg.FedoraMessagingTLSDir, []string{qtrebuild.Topic}, watcher.Handle)
	if err != nil {
		log.Error("Failed to init Fedora messaging consumer", "err", err)
		return
	}

	log.Info("Starting Fedora messaging qt rebuild watcher", "repo", cfg.QtRebuildRepo, "packages", cfg.QtRebuildPackages)
	go consumer.Run(ctx)
}

func main() {
	log.Infof("Starting DankLinux Docs API version %s", Version)

	err := godotenv.Load()
	if err != nil {
		log.Warn("Error loading .env file:", err)
	}

	cfg := config.NewConfig()
	startAPI(cfg)
}
