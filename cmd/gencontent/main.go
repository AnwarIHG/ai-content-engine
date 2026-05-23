package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anwar/ai-content-engine/internal/env"
	"github.com/anwar/ai-content-engine/internal/gemini"
)

func main() {
	env.Load()
	topic := os.Getenv("CONTENT_TOPIC")
	if topic == "" {
		topic = "AI productivity tools for small businesses"
	}
	out := os.Getenv("OUTPUT_DIR")
	if out == "" {
		out = "outputs"
	}

	os.MkdirAll(filepath.Join(out, "scripts"), 0755)

	fmt.Printf("Topic: %s\n\n", topic)

	// Blog post
	fmt.Println("[1/3] Generating blog post...")
	blog, err := gemini.GenerateFromEnv(
		"You are a professional blog writer. Write in English using markdown.",
		fmt.Sprintf("Write a 350-word blog post about: %s. Include a title (# heading), 3 sections, and a call-to-action conclusion. Use clear conversational English.", topic),
	)
	if err != nil {
		panic(err)
	}
	os.WriteFile(filepath.Join(out, "scripts", "blog.md"), []byte(blog), 0644)
	fmt.Println("  ✓ Saved: blog.md")

	// Video script
	fmt.Println("[2/3] Generating video script...")
	script, err := gemini.GenerateFromEnv(
		"You are a YouTube script writer. Write a 2-minute spoken script.",
		fmt.Sprintf("Write a 2-minute video script about: %s. Structure: hook (15s), main content (90s), CTA (15s). Mark speaker cues as [HOST]. Use natural spoken English.", topic),
	)
	if err != nil {
		panic(err)
	}
	os.WriteFile(filepath.Join(out, "scripts", "script.md"), []byte(script), 0644)
	fmt.Println("  ✓ Saved: script.md")

	// Social posts
	fmt.Println("[3/3] Generating social posts...")
	social, err := gemini.GenerateFromEnv(
		"You are a social media copywriter.",
		fmt.Sprintf("Write two things:\n1. A Twitter/X post (under 280 chars) about: %s. Include 2 hashtags.\n2. A Telegram post (3 bullet points with emojis) about: %s.\nSeparate them with '---TWITTER_END---'", topic, topic),
	)
	if err != nil {
		panic(err)
	}
	os.WriteFile(filepath.Join(out, "social.md"), []byte(social), 0644)
	fmt.Println("  ✓ Saved: social.md")

	fmt.Println("\n✅ All content generated.")
}
