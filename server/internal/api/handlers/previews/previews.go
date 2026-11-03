package previews_handler

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/duxlabs-dev/dux-docs/server/internal/services/previews"
)

var pluginIDPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{0,63}$`)

func ServePreview(store *previews.Store, pluginID string, w http.ResponseWriter, r *http.Request) {
	if !pluginIDPattern.MatchString(pluginID) {
		http.NotFound(w, r)
		return
	}

	path, etag, ok := store.Lookup(pluginID)
	if !ok {
		servePlaceholder(store, w, r)
		return
	}
	serveStored(path, etag, w, r)
}

// No placeholder: a 404 lets clients fall back to the full preview card until the thumb is generated.
func ServeThumb(store *previews.Store, pluginID string, w http.ResponseWriter, r *http.Request) {
	if !pluginIDPattern.MatchString(pluginID) {
		http.NotFound(w, r)
		return
	}

	path, etag, ok := store.Lookup(pluginID)
	if !ok {
		http.NotFound(w, r)
		return
	}
	serveStored(path, etag, w, r)
}

func serveStored(path, etag string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	quoted := `"` + etag + `"`
	w.Header().Set("ETag", quoted)

	if match := r.Header.Get("If-None-Match"); match == "*" || containsETag(match, quoted) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	http.ServeFile(w, r, path)
}

func servePlaceholder(store *previews.Store, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=60")
	http.ServeFile(w, r, store.PlaceholderPath())
}

func containsETag(header, quoted string) bool {
	for _, candidate := range strings.Split(header, ",") {
		if strings.TrimPrefix(strings.TrimSpace(candidate), "W/") == quoted {
			return true
		}
	}
	return false
}
