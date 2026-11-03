package previews

import (
	"image"
	"strings"
	"unicode"

	"github.com/duxlabs-dev/dux-docs/server/internal/models"
	"github.com/fogleman/gg"
)

func ComposeCard(p models.Plugin) (image.Image, error) {
	if err := loadFonts(); err != nil {
		return nil, err
	}

	dc := newCanvas()
	descLines, regionHeight, err := footerLayout(dc, p)
	if err != nil {
		return nil, err
	}

	dc.SetColor(colSurfaceContainer)
	dc.DrawRoundedRectangle(regionInset, regionInset, regionWidth, regionHeight, regionRadius)
	dc.Fill()

	if err := drawCardRegion(dc, p, regionHeight); err != nil {
		return nil, err
	}
	if err := drawStatusChips(dc, p.Status); err != nil {
		return nil, err
	}
	if err := drawFooter(dc, p, descLines, regionHeight); err != nil {
		return nil, err
	}
	return dc.Image(), nil
}

func drawCardRegion(dc *gg.Context, p models.Plugin, regionHeight float64) error {
	const (
		centerX  = cardWidth / 2.0
		circleR  = 76.0
		textMaxW = regionWidth - 80
	)
	circleY := regionInset + regionHeight*0.36
	nameY := regionInset + regionHeight*0.72

	dc.SetColor(withAlpha(colPrimary, 0.20))
	dc.DrawCircle(centerX, circleY, circleR)
	dc.Fill()

	letterFace, err := newFace(boldFont, 96)
	if err != nil {
		return err
	}
	dc.SetFontFace(letterFace)
	dc.SetColor(colPrimary)
	dc.DrawStringAnchored(initialLetter(p.Name), centerX, circleY, 0.5, 0.36)

	nameFace, err := newFace(boldFont, 44)
	if err != nil {
		return err
	}
	dc.SetFontFace(nameFace)
	dc.SetColor(colSurfaceText)
	dc.DrawStringAnchored(ellipsize(dc, p.Name, textMaxW), centerX, nameY, 0.5, 0.36)

	if p.Author != "" {
		authorFace, err := newFace(regularFont, 16)
		if err != nil {
			return err
		}
		dc.SetFontFace(authorFace)
		dc.SetColor(colOutline)
		dc.DrawStringAnchored("by "+p.Author, centerX, nameY+40, 0.5, 0.36)
	}

	markFace, err := newFace(boldFont, 14)
	if err != nil {
		return err
	}
	dc.SetFontFace(markFace)
	dc.SetColor(withAlpha(colPrimary, 0.75))
	dc.DrawStringAnchored("DMS", regionInset+regionWidth-16, regionInset+regionHeight-18, 1, 0.36)
	return nil
}

func initialLetter(name string) string {
	for _, r := range strings.TrimSpace(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return strings.ToUpper(string(r))
		}
	}
	return "?"
}
