package main

import (
	"fmt"
	"log"
	"os"

	"github.com/anwar/ai-content-engine/internal/twitter"
)

func main() {
	text := os.Getenv("TWEET_TEXT")
	if text == "" {
		log.Fatal("TWEET_TEXT environment variable required")
	}

	id, err := twitter.PostTweet(text)
	if err != nil {
		log.Fatalf("twitter post: %v", err)
	}

	fmt.Printf("  ✓ Tweet posted (id=%s)\n", id)
}
