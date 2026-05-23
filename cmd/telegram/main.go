package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/anwar/ai-content-engine/internal/env"
	"github.com/anwar/ai-content-engine/internal/telegram"
)

func main() {
	env.Load()
	text := os.Getenv("TELEGRAM_TEXT")
	photo := os.Getenv("TELEGRAM_PHOTO")

	if text == "" && photo == "" {
		log.Fatal("TELEGRAM_TEXT or TELEGRAM_PHOTO required")
	}

	ctx := context.Background()

	if photo != "" {
		if _, err := os.Stat(photo); err == nil {
			if err := telegram.SendPhoto(ctx, text, photo); err != nil {
				log.Fatalf("telegram photo: %v", err)
			}
			fmt.Println("  ✓ Telegram photo sent")
			return
		}
	}

	if text != "" {
		if err := telegram.SendMessage(ctx, text); err != nil {
			log.Fatalf("telegram message: %v", err)
		}
		fmt.Println("  ✓ Telegram message sent")
	}
}
