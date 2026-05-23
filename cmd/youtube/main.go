package main

import (
	"fmt"
	"log"
	"os"

	"github.com/anwar/ai-content-engine/internal/youtube"
	"golang.org/x/net/context"
)

func main() {
	videoPath := os.Getenv("VIDEO_PATH")
	title := os.Getenv("VIDEO_TITLE")
	desc := os.Getenv("VIDEO_DESC")
	thumbPath := os.Getenv("THUMBNAIL_PATH")
	privacy := os.Getenv("VIDEO_PRIVACY")

	if videoPath == "" || title == "" {
		log.Fatal("VIDEO_PATH and VIDEO_TITLE environment variables required")
	}
	if privacy == "" {
		privacy = "unlisted"
	}

	url, err := youtube.UploadVideo(context.Background(), videoPath, title, desc, thumbPath, privacy)
	if err != nil {
		log.Fatalf("youtube upload: %v", err)
	}

	fmt.Printf("  ✓ YouTube video uploaded: %s\n", url)
	fmt.Printf("video_url=%s\n", url)
}
