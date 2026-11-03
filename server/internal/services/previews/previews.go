package previews

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"image/png"
	"strings"
	"sync"

	"github.com/duxlabs-dev/dux-docs/server/internal/log"
	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

const syncWorkers = 4

type Generator struct {
	store         *Store
	thumbs        *Store
	fetcher       *imageFetcher
	publicBaseURL string
}

func NewGenerator(cacheDir, publicBaseURL string) (*Generator, error) {
	store, err := NewStore(cacheDir)
	if err != nil {
		return nil, err
	}

	if err := store.EnsurePlaceholder(renderPlaceholder); err != nil {
		return nil, err
	}

	thumbs, err := NewThumbStore(cacheDir)
	if err != nil {
		return nil, err
	}
	return &Generator{
		store:         store,
		thumbs:        thumbs,
		fetcher:       newImageFetcher(),
		publicBaseURL: strings.TrimSuffix(publicBaseURL, "/"),
	}, nil
}

func (g *Generator) Store() *Store {
	return g.store
}

func (g *Generator) ThumbStore() *Store {
	return g.thumbs
}

func (g *Generator) Sync(ctx context.Context, plugins []models.Plugin) []models.Plugin {
	out := make([]models.Plugin, len(plugins))
	copy(out, plugins)

	jobs := make(chan int)
	var wg sync.WaitGroup
	for range syncWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				g.syncPlugin(ctx, &out[i])
			}
		}()
	}
	for i := range out {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return out
}

func (g *Generator) syncPlugin(ctx context.Context, p *models.Plugin) {
	p.PreviewURL = g.publicBaseURL + "/previews/" + p.ID
	p.ThumbURL = p.PreviewURL + "/thumb"

	if p.Screenshot != "" && g.syncImageSource(ctx, *p, "screenshot", p.Screenshot) {
		return
	}
	g.syncCard(*p)
}

func (g *Generator) syncImageSource(ctx context.Context, p models.Plugin, kind, sourceURL string) bool {
	key := SourceKey(sourceURL, p)
	needCard := g.store.NeedsUpdate(p.ID, key)
	needThumb := g.thumbs.NeedsUpdate(p.ID, key)
	if !needCard && !needThumb {
		return true
	}

	src, err := g.fetcher.fetch(ctx, sourceURL)
	if err != nil {
		log.Warnf("Preview %s fetch failed for %s: %v", kind, p.ID, err)
		// A current card must survive a failed refetch; the thumb retries next sync
		return !needCard
	}

	card, err := ComposeScreenshot(src, p)
	if err != nil {
		log.Warnf("Preview composition failed for %s: %v", p.ID, err)
		return false
	}

	if !g.putJPEG(g.store, p.ID, kind, key, card) {
		return false
	}
	return g.putJPEG(g.thumbs, p.ID, kind, key, ComposeThumb(src))
}

func (g *Generator) putJPEG(store *Store, id, kind, key string, img image.Image) bool {
	data, err := encodeJPEG(img)
	if err != nil {
		log.Warnf("Preview encoding failed for %s: %v", id, err)
		return false
	}
	if err := store.Put(id, kind, key, "jpg", data); err != nil {
		log.Warnf("Preview store failed for %s: %v", id, err)
		return false
	}
	return true
}

func (g *Generator) syncCard(p models.Plugin) {
	key := SourceKey("", p)
	if g.store.NeedsUpdate(p.ID, key) {
		g.putCard(g.store, p, key, ComposeCard)
	}
	if g.thumbs.NeedsUpdate(p.ID, key) {
		g.putCard(g.thumbs, p, key, ComposeThumbCard)
	}
}

func (g *Generator) putCard(store *Store, p models.Plugin, key string, compose func(models.Plugin) (image.Image, error)) {
	card, err := compose(p)
	if err != nil {
		log.Warnf("Preview card render failed for %s: %v", p.ID, err)
		return
	}

	data, err := encodePNG(card)
	if err != nil {
		log.Warnf("Preview card encoding failed for %s: %v", p.ID, err)
		return
	}

	if err := store.Put(p.ID, "card", key, "png", data); err != nil {
		log.Warnf("Preview store failed for %s: %v", p.ID, err)
	}
}

func renderPlaceholder() ([]byte, error) {
	img, err := ComposeCard(models.Plugin{Name: "DMS Plugin"})
	if err != nil {
		return nil, err
	}
	return encodePNG(img)
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeJPEG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
