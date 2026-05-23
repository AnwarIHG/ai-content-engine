package canva

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	Width  = 1280
	Height = 720
)

func hexToRGBA(h string) color.RGBA {
	h = strings.TrimPrefix(h, "#")
	if len(h) != 6 {
		return color.RGBA{30, 30, 60, 255}
	}
	r, g, b := 0, 0, 0
	fmt.Sscanf(h, "%02x%02x%02x", &r, &g, &b)
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

// Generate creates a 1280x720 PNG thumbnail with gradient background and centered title.
// fontFace is optional; pass nil to use built-in fallback.
func Generate(title, outputPath string, fontFace font.Face) error {
	dc := gg.NewContext(Width, Height)

	// Gradient background
	top := hexToRGBA("#1a1a2e")
	bot := hexToRGBA("#16213e")
	for y := 0; y < Height; y++ {
		ratio := float64(y) / float64(Height)
		r := float64(top.R)*(1-ratio) + float64(bot.R)*ratio
		g := float64(top.G)*(1-ratio) + float64(bot.G)*ratio
		b := float64(top.B)*(1-ratio) + float64(bot.B)*ratio
		dc.SetRGB255(int(r), int(g), int(b))
		dc.DrawLine(0, float64(y), float64(Width), float64(y))
		dc.Stroke()
	}

	// Accent bar
	dc.SetHexColor("#e94560")
	dc.DrawRectangle(60, float64(Height)-120, 200, 8)
	dc.Fill()

	if fontFace != nil {
		dc.SetFontFace(fontFace)
	}

	// Word wrap
	words := strings.Fields(title)
	lines := []string{}
	line := ""
	for _, w := range words {
		test := line + " " + w
		wl, _ := dc.MeasureString(test)
		if wl > float64(Width-100) && line != "" {
			lines = append(lines, strings.TrimSpace(line))
			line = w
		} else {
			line = test
		}
	}
	if line != "" {
		lines = append(lines, strings.TrimSpace(line))
	}
	if len(lines) > 4 {
		lines = lines[:4]
	}

	// Draw centered title with shadow
	yStart := float64(Height)/2 - float64(len(lines))*45
	for _, l := range lines {
		dc.SetRGBA(0, 0, 0, 0.4)
		dc.DrawStringAnchored(l, float64(Width)/2+3, yStart+3, 0.5, 0.5)
		dc.SetHexColor("#FFFFFF")
		dc.DrawStringAnchored(l, float64(Width)/2, yStart, 0.5, 0.5)
		yStart += 90
	}

	// Footer
	dc.SetHexColor("#AAAAAA")
	dc.DrawStringAnchored("AI Content Engine", float64(Width)/2, float64(Height)-60, 0.5, 0.5)

	os.MkdirAll(filepath.Dir(outputPath), 0755)
	return dc.SavePNG(outputPath)
}

// LoadSystemFont tries common system font paths and returns a 56pt face.
func LoadSystemFont() (font.Face, error) {
	paths := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSans-Bold.ttf",
		"/System/Library/Fonts/Helvetica.ttc",
		"/System/Library/Fonts/Supplemental/Arial.ttf",
		"C:\\Windows\\Fonts\\Arial.ttf",
	}
	for _, fp := range paths {
		if _, err := os.Stat(fp); err == nil {
			fBytes, err := os.ReadFile(fp)
			if err != nil {
				continue
			}
			fParsed, err := opentype.Parse(fBytes)
			if err != nil {
				continue
			}
			face, err := opentype.NewFace(fParsed, &opentype.FaceOptions{
				Size:    56,
				DPI:     72,
				Hinting: font.HintingFull,
			})
			if err == nil {
				return face, nil
			}
		}
	}
	return nil, fmt.Errorf("no system font found")
}
