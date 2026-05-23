package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/anwar/ai-content-engine/internal/canva"
	"github.com/anwar/ai-content-engine/internal/env"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

func main() {
	env.Load()
	title := flag.String("title", "AI Content Engine", "Thumbnail title text")
	output := flag.String("output", "outputs/thumbnails/thumbnail.png", "Output PNG path")
	flag.Parse()

	// Try system font first; fall back to downloading Noto Sans Bold
	face, err := canva.LoadSystemFont()
	if err != nil {
		fmt.Println("  ℹ No system font found, downloading Noto Sans Bold (MIT license)...")
		face = downloadFont()
	}

	if err := canva.Generate(*title, *output, face); err != nil {
		panic(err)
	}
	fmt.Printf("  ✓ Thumbnail saved: %s\n", *output)
}

func downloadFont() font.Face {
	fontPath := "/tmp/NotoSans-Bold.ttf"
	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		urls := []string{
			"https://github.com/google/fonts/raw/main/ofl/notosans/NotoSans-Bold.ttf",
			"https://raw.githubusercontent.com/google/fonts/main/ofl/notosans/NotoSans-Bold.ttf",
		}
		var data []byte
		for _, u := range urls {
			resp, err := http.Get(u)
			if err == nil && resp.StatusCode == 200 {
				data, _ = io.ReadAll(resp.Body)
				resp.Body.Close()
				break
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
		if data == nil {
			fmt.Println("  ⚠ Could not download font, using default")
			return nil
		}
		os.MkdirAll(filepath.Dir(fontPath), 0755)
		os.WriteFile(fontPath, data, 0644)
	}

	fBytes, err := os.ReadFile(fontPath)
	if err != nil {
		return nil
	}
	fParsed, err := opentype.Parse(fBytes)
	if err != nil {
		return nil
	}
	face, err := opentype.NewFace(fParsed, &opentype.FaceOptions{
		Size:    56,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil
	}
	return face
}
