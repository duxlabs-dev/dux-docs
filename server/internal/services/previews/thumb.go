package previews

import (
	"image"

	"github.com/duxlabs-dev/dux-docs/server/internal/models"
)

// Thumbs are the screenshot alone at the 16:10 ratio the shell's plugin cards use.
// Clients draw their own title and badges, so no footer or chips.
func ComposeThumb(src image.Image) image.Image {
	dc := newCanvasSized(thumbWidth, thumbHeight)
	drawFitted(dc, src, 0, 0, thumbWidth, thumbHeight)
	return dc.Image()
}

func ComposeThumbCard(p models.Plugin) (image.Image, error) {
	if err := loadFonts(); err != nil {
		return nil, err
	}
	dc := newCanvasSized(thumbWidth, thumbHeight)
	if err := drawCardRegion(dc, p, thumbHeight-2*regionInset); err != nil {
		return nil, err
	}
	return dc.Image(), nil
}
