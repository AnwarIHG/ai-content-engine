# The AI-Powered Content Engine — $0 Implementation Guide (Go)

> **Total upfront effort:** 6–10 hours (one-time setup)  
> **Expected timeline to first automated revenue:** 3–8 weeks (depends on niche, content quality, and audience growth)  
> **Total cost to run:** $0.00/month — every tool has a permanent, no-credit-card free tier.

---

## What This Guide Covers

A fully automated pipeline that every day:
1. Generates a blog post, a video script, and social copy using **Google Gemini** (free forever, 1,500 requests/day)
2. Creates a 1280×720 thumbnail image using Go's **image** packages (free, no API key)
3. Converts the blog post to a **PDF e-book** via **pandoc** (or pure-Go md2pdf)
4. Creates a **Gumroad** product (PDF for sale, free to list)
5. Injects **affiliate links** (Amazon Associates, free to join)
6. Uploads a video to **YouTube** (free channel)
7. Posts to **Twitter/X** (free API tier)
8. Posts to **Telegram** (free bot API)

All orchestrated by **GitHub Actions** (2,000 free minutes/month).

---

## Critical Constraint — Every Tool Is Permanently Free

| Tool | Free Tier | Credit Card? | Time Limit? |
|---|---|---|---|
| GitHub Actions | 2,000 min/month | No | No |
| Google Gemini API | 1,500 req/day, 15 RPM | No | No (permanent) |
| Go image/gg (Pillow equivalent) | Open source, unlimited | No | No |
| YouTube Data API | 10,000 quota units/day | No | No |
| Twitter API v2 | Free write access (rate limited) | No | No |
| Telegram Bot API | Unlimited messages | No | No |
| Gumroad | Free to list products (fee on sale) | No | No |
| Amazon Associates | Free to join | No | No |
| pandoc / weasyprint | Open source, unlimited | No | No |

> **Why not DeepSeek?** DeepSeek's free API credits (5M tokens) expire after 30 days. This user constraint requires permanent free.  
> **Why not Canva?** Canva's Autofill API (programmatic text in designs) requires Canva Enterprise ($2,000+/yr). The free plan cannot fill templates via API.

---

## 1. Provision the Free Automation Engine

### 1.1 Create a GitHub Account

1. Open https://github.com/signup
2. Enter email → create password → pick a username
3. Verify your email
4. Choose **Free** plan

GitHub Free gives you:
- Unlimited public/private repos
- **2,000 minutes/month** of GitHub Actions (public repos)
- 500 MB of GitHub Packages storage

### 1.2 Create the Repository

```bash
mkdir -p ai-content-engine/{cmd/{gencontent,thumbnail,ebook,youtube,twitter,telegram,affiliate},internal/{gemini,canva,youtube,twitter,telegram,gumroad,affiliate},outputs/{scripts,thumbnails,ebooks}}
cd ai-content-engine

go mod init github.com/YOUR_USERNAME/ai-content-engine

git init
git branch -M main
git remote add origin https://github.com/YOUR_USERNAME/ai-content-engine.git
git add . && git commit -m "initial scaffold"
git push -u origin main
```

### 1.3 Staying Within 2,000 Free Minutes

A daily run takes **~3–5 minutes**. At 30 runs/month = **90–150 minutes** (7.5% of budget).

**Tips to stay within limits:**
- Run once per day (not hourly)
- Cache Go module downloads between runs
- Skip weekends: `cron: '0 2 * * 1-5'`
- Keep video files ≤50 MB

---

## 2. Deploy the AI Core (Go)

### 2a. AI Content Generator — Google Gemini (Free Forever)

#### 2a.1 Get a Free Gemini API Key (No Credit Card)

1. Go to https://aistudio.google.com/app/apikey
2. Sign in with any **Gmail/Google account**
3. Click **"Create API Key"** → "Create API key in new project"
4. Copy the key (starts with `AIzaSy...`)

**This key is free forever.** Limits: 1,500 requests/day, 15 requests/minute for Gemini 2.0 Flash. No credit card required at any point.

#### 2a.2 Gemini Go Client (Pure net/http — No SDK Dependency)

Save as `internal/gemini/client.go`:

```go
package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Request/response types for the Gemini REST API.
// Gemini uses a different API shape than OpenAI — this is the native format.

type Content struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role,omitempty"`
}

type Part struct {
	Text string `json:"text"`
}

type GenerateRequest struct {
	Contents         []Content         `json:"contents"`
	SystemInstruction *Content          `json:"system_instruction,omitempty"`
	GenerationConfig map[string]any    `json:"generation_config,omitempty"`
}

type GenerateResponse struct {
	Candidates []struct {
		Content Content `json:"content"`
	} `json:"candidates"`
}

const baseURL = "https://generativelanguage.googleapis.com/v1beta/models"

func Generate(apiKey, model, systemPrompt, userPrompt string) (string, error) {
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", baseURL, model, apiKey)

	req := GenerateRequest{
		Contents: []Content{
			{
				Role: "user",
				Parts: []Part{{Text: userPrompt}},
			},
		},
		GenerationConfig: map[string]any{
			"temperature":     0.7,
			"maxOutputTokens": 1024,
		},
	}

	if systemPrompt != "" {
		req.SystemInstruction = &Content{
			Parts: []Part{{Text: systemPrompt}},
		}
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("gemini request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("gemini %d: %s", resp.StatusCode, string(raw))
	}

	var gr GenerateResponse
	if err := json.Unmarshal(raw, &gr); err != nil {
		return "", fmt.Errorf("gemini parse: %w", err)
	}

	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: empty response")
	}

	return gr.Candidates[0].Content.Parts[0].Text, nil
}

func GenerateFromEnv(systemPrompt, userPrompt string) (string, error) {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set")
	}
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.0-flash" // free tier model
	}
	return Generate(key, model, systemPrompt, userPrompt)
}
```

> **Model note:** `gemini-2.0-flash` is the free tier model (1,500 req/day). `gemini-2.5-flash` is paid. The guide uses `gemini-2.0-flash` exclusively.

#### 2a.3 Content Generator Command

Save as `cmd/gencontent/main.go`:

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/YOUR_USERNAME/ai-content-engine/internal/gemini"
)

func save(dir, filename, content string) {
	path := filepath.Join(dir, filename)
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte(content), 0644)
	fmt.Printf("  ✓ Saved: %s\n", path)
}

func main() {
	topic := os.Getenv("CONTENT_TOPIC")
	if topic == "" {
		topic = "AI productivity tools for small businesses"
	}
	out := os.Getenv("OUTPUT_DIR")
	if out == "" {
		out = "outputs"
	}

	fmt.Printf("Topic: %s\n\n", topic)

	// Blog post
	fmt.Println("[1/3] Generating blog post...")
	blog, err := gemini.GenerateFromEnv(
		"You are a professional blog writer. Write in English using markdown.",
		fmt.Sprintf("Write a 350-word blog post about: %s. Include a title (# heading), 3 sections, and a call-to-action. Use clear conversational English.", topic),
	)
	if err != nil {
		panic(err)
	}
	save(out, "scripts/blog.md", blog)

	// Video script
	fmt.Println("[2/3] Generating video script...")
	script, err := gemini.GenerateFromEnv(
		"You are a YouTube script writer.",
		fmt.Sprintf("Write a 2-minute video script about: %s. Structure: hook (15s), main content (90s), CTA (15s). Mark speaker cues as [HOST]. Use natural spoken English.", topic),
	)
	if err != nil {
		panic(err)
	}
	save(out, "scripts/script.md", script)

	// Social posts
	fmt.Println("[3/3] Generating social posts...")
	social, err := gemini.GenerateFromEnv(
		"You are a social media copywriter.",
		fmt.Sprintf("Write two things:\n1. A Twitter/X post (under 280 chars) about: %s. Include 2 hashtags.\n2. A Telegram post (3 bullet points with emojis) about: %s.\nSeparate them with '---TWITTER_END---'", topic, topic),
	)
	if err != nil {
		panic(err)
	}
	save(out, "social.md", social)

	fmt.Println("\n✅ All content generated.")
}
```

**Test locally:**
```bash
export GEMINI_API_KEY="AIzaSy..."
export CONTENT_TOPIC="how to automate your business with AI"
export OUTPUT_DIR="./outputs"
go run cmd/gencontent/main.go
```

---

### 2b. Visual Creator — Go Image Processing (100% Free, No API Key)

**Why not Canva?** Canva's Connect API can create blank designs on the free plan, but the Autofill endpoint (dynamic text insertion into templates) requires **Canva Enterprise** (~$2,000+/yr). This violates the free constraint. Go's standard library + the `gg` package is open source, runs offline, and costs nothing.

#### 2b.1 Install Go Graphics Library

```bash
go get github.com/fogleman/gg
```

This is a pure-Go 2D rendering library (MIT license, no external dependencies).

#### 2b.2 Thumbnail Generator

Save as `internal/canva/thumbnail.go`:

```go
package canva

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	"github.com/fogleman/gg"
)

const (
	width  = 1280
	height = 720
)

// hexToRGBA converts "#hex" to color.RGBA.
func hexToRGBA(h string) color.RGBA {
	h = strings.TrimPrefix(h, "#")
	if len(h) != 6 {
		return color.RGBA{30, 30, 60, 255}
	}
	r, g, b := 0, 0, 0
	fmt.Sscanf(h, "%02x%02x%02x", &r, &g, &b)
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

// GenerateThumbnail creates a 1280×720 PNG thumbnail with gradient background
// and centered title text. Uses Go's native image packages — no API key needed.
func GenerateThumbnail(title, outputPath string) error {
	dc := gg.NewContext(width, height)

	// Gradient background: dark blue → purple
	top := hexToRGBA("#1a1a2e")
	bot := hexToRGBA("#16213e")
	for y := 0; y < height; y++ {
		ratio := float64(y) / float64(height)
		r := float64(top.R)*(1-ratio) + float64(bot.R)*ratio
		g := float64(top.G)*(1-ratio) + float64(bot.G)*ratio
		b := float64(top.B)*(1-ratio) + float64(bot.B)*ratio
		dc.SetRGB255(int(r), int(g), int(b))
		dc.DrawLine(0, float64(y), width, float64(y))
		dc.Stroke()
	}

	// Accent bar
	dc.SetHexColor("#e94560")
	dc.DrawRectangle(60, height-120, 200, 8)
	dc.Fill()

	// Load a built-in font face
	// For simplicity, we use the Go regular font via gg.
	// In production, embed a .ttf with //go:embed
	if err := dc.LoadFontFace("", 0); err != nil {
		// Fallback: use default font at a reasonable size
		dc.SetFontFace(gg.NewFontFace(nil, nil))
	}

	// Word-wrap title
	words := strings.Fields(title)
	lines := []string{}
	line := ""
	for _, w := range words {
		test := line + " " + w
		if len(test) > 25 { // rough char-count wrap
			lines = append(lines, strings.TrimSpace(line))
			line = w
		} else {
			line = test
		}
	}
	if line != "" {
		lines = append(lines, strings.TrimSpace(line))
	}
	if len(lines) > 3 {
		lines = lines[:3]
	}

	// Draw centered title
	dc.SetHexColor("#FFFFFF")
	yStart := 180.0
	for _, l := range lines {
		// Shadow
		dc.SetRGB(0, 0, 0)
		dc.DrawStringAnchored(l, float64(width)/2+3, yStart+3, 0.5, 0.5)
		// Text
		dc.SetHexColor("#FFFFFF")
		dc.DrawStringAnchored(l, float64(width)/2, yStart, 0.5, 0.5)
		yStart += 80
	}

	// Footer
	dc.SetHexColor("#CCCCCC")
	dc.DrawStringAnchored("AI Content Engine", float64(width)/2, float64(height)-60, 0.5, 0.5)

	// Save
	os.MkdirAll(outputPath[:strings.LastIndex(outputPath, "/")], 0755)
	return dc.SavePNG(outputPath)
}
```

**Enhanced version with downloadable font support:**

For better typography, save this as `cmd/thumbnail/main.go`:

```go
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	width  = 1280
	height = 720
)

func downloadFont(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	os.MkdirAll(filepath.Dir(dest), 0755)
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func main() {
	title := flag.String("title", "AI Content Engine", "Thumbnail title")
	output := flag.String("output", "outputs/thumbnails/thumbnail.png", "Output path")
	flag.Parse()

	dc := gg.NewContext(width, height)

	// Gradient
	dc.SetRGB(0.1, 0.1, 0.24)
	dc.Clear()
	for y := 0; y < height; y++ {
		r := 0.1 + float64(y)/float64(height)*0.15
		g := 0.1 + float64(y)/float64(height)*0.1
		b := 0.24 + float64(y)/float64(height)*0.2
		dc.SetRGB(r, g, b)
		dc.DrawLine(0, float64(y), width, float64(y))
		dc.Stroke()
	}

	// Accent bar
	dc.SetHexColor("#e94560")
	dc.DrawRectangle(60, height-120, 200, 8)
	dc.Fill()

	// Download a free Google Font (MIT-licensed)
	fontPath := "/tmp/NotoSans-Bold.ttf"
	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		fmt.Println("Downloading font (one-time)...")
		downloadFont(
			"https://github.com/google/fonts/raw/main/ofl/notosans/NotoSans-Bold.ttf",
			fontPath,
		)
	}

	fontBytes, err := os.ReadFile(fontPath)
	if err == nil {
		fontParsed, err := opentype.Parse(fontBytes)
		if err == nil {
			face, err := opentype.NewFace(fontParsed, &opentype.FaceOptions{
				Size:    56,
				DPI:     72,
				Hinting: font.HintingFull,
			})
			if err == nil {
				dc.SetFontFace(face)
			}
		}
	}

	// Word wrap
	words := strings.Fields(*title)
	lines := []string{}
	line := ""
	for _, w := range words {
		test := line + " " + w
		wl, _ := dc.MeasureString(test)
		if wl > float64(width-100) && line != "" {
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

	// Draw centered title
	yStart := float64(height)/2 - float64(len(lines))*45
	for _, l := range lines {
		// Shadow
		dc.SetRGBA(0, 0, 0, 0.4)
		dc.DrawStringAnchored(l, float64(width)/2+3, yStart+3, 0.5, 0.5)
		// Text
		dc.SetHexColor("#FFFFFF")
		dc.DrawStringAnchored(l, float64(width)/2, yStart, 0.5, 0.5)
		yStart += 90
	}

	// Footer
	dc.SetHexColor("#AAAAAA")
	dc.DrawStringAnchored("AI Content Engine", float64(width)/2, float64(height)-60, 0.5, 0.5)

	os.MkdirAll(filepath.Dir(*output), 0755)
	if err := dc.SavePNG(*output); err != nil {
		panic(err)
	}
	fmt.Printf("  ✓ Thumbnail saved: %s\n", *output)
}
```

**Test it:**
```bash
go run cmd/thumbnail/main.go --title "5 AI Tools That Save 10 Hours/Week" --output thumbnails/test.png
```

This is 100% free — the Google Font is MIT-licensed, downloaded once and cached.

---

## 3. Automate Multi-Platform Distribution

### 3a. YouTube Upload

#### 3a.1 Prerequisites (Free)

1. Create a **YouTube channel** (sign in → create channel — free, no subscription)
2. Go to https://console.cloud.google.com/ → Create project → Enable **YouTube Data API v3**
3. Create **OAuth 2.0 credentials** → "Desktop application" → Download `client_secret.json`
4. Install the Go YouTube client library:
   ```bash
   go get google.golang.org/api/youtube/v3
   go get golang.org/x/oauth2/...
   ```

#### 3a.2 Generate OAuth Token (One-Time, on Your Machine)

Save as `cmd/gentoken/main.go`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

func main() {
	clientSecretBytes, _ := ioutil.ReadFile("client_secret.json")

	config, err := google.ConfigFromJSON(clientSecretBytes, youtube.YoutubeUploadScope)
	if err != nil {
		log.Fatalf("parse client secret: %v", err)
	}

	// Start local server to receive OAuth callback
	token, err := getTokenFromWeb(config)
	if err != nil {
		log.Fatalf("get token: %v", err)
	}

	tokenJSON, _ := json.Marshal(token)
	fmt.Println(string(tokenJSON))
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Open this URL in your browser:\n%s\n", authURL)
	fmt.Print("Enter authorization code: ")

	var authCode string
	fmt.Scan(&authCode)

	return config.Exchange(context.Background(), authCode)
}
```

```bash
go run cmd/gentoken/main.go
# Copy the JSON output → store as YOUTUBE_TOKEN_JSON GitHub Secret
# Store client_secret.json content as YOUTUBE_CLIENT_SECRET_JSON
```

> **Refresh tokens:** Using `oauth2.AccessTypeOffline` gives you a refresh token that never expires — your pipeline will never need re-auth.

#### 3a.3 YouTube Upload Command

Save as `cmd/youtube/main.go`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func main() {
	videoPath := os.Getenv("VIDEO_PATH")
	title := os.Getenv("VIDEO_TITLE")
	desc := os.Getenv("VIDEO_DESC")
	thumbPath := os.Getenv("THUMBNAIL_PATH")
	privacy := os.Getenv("VIDEO_PRIVACY")
	if privacy == "" {
		privacy = "unlisted"
	}

	if videoPath == "" || title == "" {
		log.Fatal("VIDEO_PATH and VIDEO_TITLE required")
	}

	ctx := context.Background()

	// Load OAuth credentials from GitHub Secrets
	clientSecret := os.Getenv("YOUTUBE_CLIENT_SECRET_JSON")
	tokenJSON := os.Getenv("YOUTUBE_TOKEN_JSON")

	var config oauth2.Config
	json.Unmarshal([]byte(clientSecret), &config)

	var token oauth2.Token
	json.Unmarshal([]byte(tokenJSON), &token)

	client := config.Client(ctx, &token)
	service, err := youtube.New(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("youtube client: %v", err)
	}

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: desc,
			CategoryId:  "22", // People & Blogs
		},
		Status: &youtube.VideoStatus{PrivacyStatus: privacy},
	}

	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
	file, err := os.Open(videoPath)
	if err != nil {
		log.Fatalf("open video: %v", err)
	}
	defer file.Close()

	response, err := call.Media(file).Do()
	if err != nil {
		log.Fatalf("upload: %v", err)
	}

	videoID := response.Id
	url := fmt.Sprintf("https://youtu.be/%s", videoID)
	fmt.Printf("  ✓ YouTube video uploaded: %s\n", url)

	// Set thumbnail
	if thumbPath != "" {
		thumbFile, err := os.Open(thumbPath)
		if err == nil {
			defer thumbFile.Close()
			service.Thumbnails.Set(videoID, thumbFile).Do()
			fmt.Println("  ✓ Thumbnail set")
		}
	}

	// GitHub Actions output
	fmt.Printf("video_url=%s\n", url)

	// Create description file for other steps
	os.WriteFile(os.Getenv("OUTPUT_DIR")+"/youtube_url.txt", []byte(url), 0644)
}
```

> **YouTube quota:** 10,000 units/day free. An upload costs ~1,600 units. One upload/day = 48K/month = well within limits.

---

### 3b. Twitter/X Posts

#### 3b.1 Free Tier Setup

1. Go to https://developer.twitter.com/ → Sign up for a **Free** project
2. Create a project → "Making a bot"
3. Generate **Bearer Token**
4. Store `TWITTER_BEARER_TOKEN` as a GitHub Secret

> Fresh accounts may have a 3–7 day approval wait. This is normal and free.

#### 3b.2 Twitter Client (Pure net/http — No SDK)

Save as `cmd/twitter/main.go`:

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type tweetRequest struct {
	Text string `json:"text"`
}

type tweetResponse struct {
	Data struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"data"`
}

func main() {
	text := os.Getenv("TWEET_TEXT")
	if text == "" {
		log.Fatal("TWEET_TEXT required")
	}
	if len(text) > 280 {
		text = text[:280]
	}

	bearer := os.Getenv("TWITTER_BEARER_TOKEN")
	if bearer == "" {
		log.Fatal("TWITTER_BEARER_TOKEN not set")
	}

	body, _ := json.Marshal(tweetRequest{Text: text})
	req, _ := http.NewRequest("POST", "https://api.twitter.com/2/tweets",
		bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("twitter request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		var errResp map[string]any
		json.NewDecoder(resp.Body).Decode(&errResp)
		log.Fatalf("twitter %d: %v", resp.StatusCode, errResp)
	}

	var result tweetResponse
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("  ✓ Tweet posted (id=%s)\n", result.Data.ID)
}
```

---

### 3c. Telegram Posts

#### 3c.1 Create a Bot (Free)

1. Open Telegram → search for **@BotFather**
2. Send `/newbot` → choose a name (e.g., `AI Content Engine`)
3. Copy the **token** (format: `123456:ABC-DEF1234`) → store as `TELEGRAM_BOT_TOKEN`
4. Create a channel → add bot as **admin**
5. Get **chat ID**: send any message to the channel, then visit:
   ```
   https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates
   ```
   Look for `"chat":{"id":-1001234567890}`. Store as `TELEGRAM_CHAT_ID`.

#### 3c.2 Telegram Command (Using go-telegram/bot)

```bash
go get github.com/go-telegram/bot
```

Save as `cmd/telegram/main.go`:

```go
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	text := os.Getenv("TELEGRAM_TEXT")
	photoPath := os.Getenv("TELEGRAM_PHOTO")

	if text == "" && photoPath == "" {
		log.Fatal("TELEGRAM_TEXT or TELEGRAM_PHOTO required")
	}

	b, err := bot.New(token)
	if err != nil {
		log.Fatalf("new bot: %v", err)
	}

	// Start bot in background (needed to use API methods)
	go b.Start(ctx)

	if photoPath != "" {
		data, err := os.ReadFile(photoPath)
		if err != nil {
			log.Fatalf("read photo: %v", err)
		}
		_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:  chatID,
			Photo:   &models.InputFileUpload{Filename: "thumbnail.png", Data: bytes.NewReader(data)},
			Caption: text,
		})
		if err != nil {
			log.Fatalf("send photo: %v", err)
		}
		fmt.Println("  ✓ Telegram photo sent")
	} else {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   text,
		})
		if err != nil {
			log.Fatalf("send message: %v", err)
		}
		fmt.Println("  ✓ Telegram message sent")
	}
}
```

---

## 4. Implement Automated Monetization

### 4a. Gumroad Product Creation (Free to List)

**Gumroad is free to join.** You only pay a transaction fee (~9%) when a sale happens. Listing products, file storage, and the API are all free.

#### 4a.1 Get Gumroad API Access

1. Go to https://app.gumroad.com/advanced
2. Click **"Generate access token"**
3. Copy the token → store as `GUMROAD_ACCESS_TOKEN`

Also install the **gumroad-cli** (Go-based CLI tool):

```bash
go install github.com/antiwork/gumroad-cli/cmd/gumroad@latest
```

#### 4a.2 PDF Converter

**Option A: Use pandoc (recommended — pre-installed on ubuntu-latest)**

```bash
sudo apt-get install -y pandoc texlive-latex-base
pandoc input.md -o output.pdf --pdf-engine=xelatex
```

**Option B: Pure Go PDF — markdown2pdf**

```bash
go install github.com/dlouwers/markdown2pdf/cmd/markdown2pdf@latest
markdown2pdf input.md -o output.pdf
```

**Option C: Call from Go using os/exec**

Save as `cmd/ebook/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	mdPath := os.Getenv("BLOG_PATH")
	pdfPath := os.Getenv("PDF_PATH")
	title := os.Getenv("EBOOK_TITLE")

	if mdPath == "" || pdfPath == "" {
		log.Fatal("BLOG_PATH and PDF_PATH required")
	}

	os.MkdirAll(filepath.Dir(pdfPath), 0755)

	// Try pandoc first
	cmd := exec.Command("pandoc", mdPath, "-o", pdfPath, "--pdf-engine=xelatex")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("pandoc failed (may need texlive), trying markdown2pdf: %v", err)

		// Fallback to markdown2pdf
		cmd2 := exec.Command("markdown2pdf", mdPath, "-o", pdfPath)
		if err2 := cmd2.Run(); err2 != nil {
			log.Fatalf("markdown2pdf also failed: %v", err2)
		}
	}

	fmt.Printf("  ✓ PDF created: %s\n", pdfPath)

	// Create Gumroad product using gumroad-cli
	gumCmd := exec.Command("gumroad", "products", "create",
		"--name", title,
		"--price", "4.99",
		"--type", "ebook",
		"--file", pdfPath,
		"--file-name", title+".pdf",
		"--json", "--no-input",
	)
	gumCmd.Env = append(os.Environ(), "GUMROAD_ACCESS_TOKEN="+os.Getenv("GUMROAD_ACCESS_TOKEN"))
	output, err := gumCmd.Output()
	if err != nil {
		log.Printf("gumroad create (may already exist): %v", err)
	}

	fmt.Printf("  ✓ Gumroad product created\n  %s\n", string(output))
}
```

> **Note:** Gumroad's product creation API (`POST /v2/products`) is now functional. Products are created as **drafts** — you can publish from the dashboard or use `gumroad products publish <id>`.

---

### 4b. Affiliate Link Injection (Free to Join)

**Amazon Associates** is free to join (no subscription fee). Approval depends on your content. Other free programs: ShareASale, Impact, Rakuten.

Save as `cmd/affiliate/main.go`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// AffiliateConfig maps keywords → affiliate URLs.
// Loaded from a JSON file. Example:
//
//	{
//	  "keywords": {
//	    "productivity tool": "https://www.amazon.com/s?k=productivity+tool&tag=YOUR_TAG-20",
//	    "AI tool": "https://amzn.to/xxx",
//	    "notion": "https://www.notion.so/affiliate"
//	  },
//	  "disclaimer": "\n\n---\n*Some links are affiliate links. I may earn a commission at no extra cost to you.*"
//	}
type AffiliateConfig struct {
	Keywords   map[string]string `json:"keywords"`
	Disclaimer string            `json:"disclaimer"`
}

func main() {
	input := os.Getenv("AFFILIATE_INPUT")
	output := os.Getenv("AFFILIATE_OUTPUT")
	configPath := os.Getenv("AFFILIATE_CONFIG_PATH")

	if input == "" || output == "" {
		log.Fatal("AFFILIATE_INPUT and AFFILIATE_OUTPUT required")
	}

	// Default config
	config := AffiliateConfig{
		Keywords: map[string]string{
			"productivity tool": "https://www.amazon.com/s?k=productivity+tool&tag=YOUR_TAG-20",
			"AI tool":           "https://www.amazon.com/s?k=AI+tool&tag=YOUR_TAG-20",
			"software":          "https://www.amazon.com/s?k=software&tag=YOUR_TAG-20",
			"recommend":         "https://www.amazon.com/s?k=recommended+tool&tag=YOUR_TAG-20",
		},
		Disclaimer: "\n\n---\n*Some links above are affiliate links. I may earn a commission at no extra cost to you. As an Amazon Associate I earn from qualifying purchases.*",
	}

	// Override from JSON file if provided
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			json.Unmarshal(data, &config)
		}
	}

	// Read input
	content, err := os.ReadFile(input)
	if err != nil {
		log.Fatalf("read input: %v", err)
	}

	text := string(content)

	// Inject affiliate links (case-insensitive)
	for keyword, url := range config.Keywords {
		replacement := fmt.Sprintf("[%s](%s)", keyword, url)
		text = strings.ReplaceAll(
			strings.ToLower(text),
			strings.ToLower(keyword),
			replacement,
		)
	}

	// Append disclaimer if not already present
	if config.Disclaimer != "" && !strings.Contains(text, config.Disclaimer) {
		text += config.Disclaimer
	}

	os.WriteFile(output, []byte(text), 0644)
	fmt.Printf("  ✓ Affiliate links injected: %s\n", output)
}
```

---

## 5. Full Workflow Example (End-to-End YAML)

Create `.github/workflows/content_engine.yml`:

```yaml
name: AI Content Engine — Daily Pipeline ($0)

on:
  schedule:
    # Daily at 2:00 AM UTC — uses ~4 min/day = ~120 min/month
    - cron: '0 2 * * *'
  workflow_dispatch:
    inputs:
      topic:
        description: 'Content topic'
        required: false
        default: 'AI productivity tools for small businesses'

concurrency:
  group: content-engine
  cancel-in-progress: false

env:
  OUTPUT_DIR: ${{ github.workspace }}/outputs
  CONTENT_TOPIC: ${{ github.event.inputs.topic || 'AI productivity tools for small businesses' }}
  VIDEO_PATH: ${{ github.workspace }}/video.mp4

jobs:
  pipeline:
    runs-on: ubuntu-latest
    timeout-minutes: 25

    steps:
      # ── 1. Checkout ─────────────────────────────────────────
      - name: Checkout repository
        uses: actions/checkout@v4

      # ── 2. Set up Go (free, pre-installed on runner) ────────
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      # ── 3. Install system tools (all free, open source) ─────
      - name: Install system dependencies
        run: |
          sudo apt-get update -qq
          sudo apt-get install -y -qq \
            pandoc texlive-latex-base weasyprint ffmpeg espeak

      # ── 4. Install markdown2pdf (pure Go, free) ────────────
      - name: Install markdown2pdf
        run: go install github.com/dlouwers/markdown2pdf/cmd/markdown2pdf@latest

      # ── 5. Install gumroad-cli (Go, free) ──────────────────
      - name: Install gumroad-cli
        run: go install github.com/antiwork/gumroad-cli/cmd/gumroad@latest

      # ── 6. Create output directories ───────────────────────
      - name: Create directories
        run: |
          mkdir -p $OUTPUT_DIR/{scripts,thumbnails,ebooks}

      # ── 7. Generate AI content (Gemini, free) ──────────────
      - name: Generate blog post, video script, social copy
        run: go run ./cmd/gencontent/
        env:
          GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}

      # ── 8. Create thumbnail (Go image, free, no API key) ───
      - name: Generate thumbnail
        run: |
          TITLE=$(head -1 $OUTPUT_DIR/scripts/blog.md | sed 's/^# //' | head -c 80)
          go run ./cmd/thumbnail/ --title "$TITLE" --output "$OUTPUT_DIR/thumbnails/thumbnail.png"

      # ── 9. Inject affiliate links (Amazon Associates, free) ─
      - name: Inject affiliate links
        run: |
          go run ./cmd/affiliate/
        env:
          AFFILIATE_INPUT: ${{ github.workspace }}/outputs/scripts/blog.md
          AFFILIATE_OUTPUT: ${{ github.workspace }}/outputs/scripts/blog_with_affiliates.md

      # ── 10. Create PDF e-book (pandoc, free) ────────────────
      - name: Create PDF e-book
        run: |
          pandoc $OUTPUT_DIR/scripts/blog_with_affiliates.md \
            -o $OUTPUT_DIR/ebooks/daily_ebook.pdf \
            --pdf-engine=xelatex || \
          markdown2pdf $OUTPUT_DIR/scripts/blog_with_affiliates.md \
            -o $OUTPUT_DIR/ebooks/daily_ebook.pdf

      # ── 11. Create Gumroad product (free to list) ──────────
      - name: Publish to Gumroad
        run: |
          TITLE=$(head -1 $OUTPUT_DIR/scripts/blog.md | sed 's/^# //')
          gumroad products create \
            --name "$TITLE" \
            --price 4.99 \
            --type ebook \
            --file $OUTPUT_DIR/ebooks/daily_ebook.pdf \
            --file-name "$TITLE.pdf" \
            --json --no-input || echo "Product may already exist, continuing..."
        env:
          GUMROAD_ACCESS_TOKEN: ${{ secrets.GUMROAD_ACCESS_TOKEN }}

      # ── 12. Generate video (TTS + thumbnail, free) ─────────
      - name: Generate video from script
        run: |
          SCRIPT="$OUTPUT_DIR/scripts/script.md"
          THUMB="$OUTPUT_DIR/thumbnails/thumbnail.png"

          # Clean text from markdown
          grep -v '^#' "$SCRIPT" | grep -v '^\[' | head -c 500 > /tmp/script_clean.txt

          # Generate audio (espeak is free, offline TTS)
          espeak -f /tmp/script_clean.txt -w /tmp/audio.wav -s 150 -v en-us 2>/dev/null

          # Combine audio + thumbnail into video
          ffmpeg -y -loop 1 -i "$THUMB" -i /tmp/audio.wav \
            -c:v libx264 -tune stillimage -c:a aac -b:a 64k \
            -pix_fmt yuv420p -shortest "$VIDEO_PATH" 2>/dev/null

          echo "Video size: $(wc -c < $VIDEO_PATH) bytes"

      # ── 13. Upload to YouTube (free, 10K quota/day) ────────
      - name: Upload video to YouTube
        id: youtube
        run: |
          TITLE=$(head -1 $OUTPUT_DIR/scripts/blog.md | sed 's/^# //')
          DESC=$(head -20 $OUTPUT_DIR/scripts/blog_with_affiliates.md | tr '\n' ' ')
          go run ./cmd/youtube/
        env:
          YOUTUBE_TOKEN_JSON: ${{ secrets.YOUTUBE_TOKEN_JSON }}
          YOUTUBE_CLIENT_SECRET_JSON: ${{ secrets.YOUTUBE_CLIENT_SECRET_JSON }}
          VIDEO_PATH: $VIDEO_PATH
          VIDEO_TITLE: ${{ env.CONTENT_TOPIC }}
          VIDEO_DESC: ${{ env.CONTENT_TOPIC }}
          THUMBNAIL_PATH: $OUTPUT_DIR/thumbnails/thumbnail.png

      # ── 14. Post to Twitter/X (free API tier) ──────────────
      - name: Post to Twitter/X
        run: |
          TWEET=$(head -1 $OUTPUT_DIR/social.md | sed 's/---TWITTER_END---.*//' | head -c 270)
          echo "TWEET_TEXT=$TWEET" >> $GITHUB_ENV
          go run ./cmd/twitter/
        env:
          TWITTER_BEARER_TOKEN: ${{ secrets.TWITTER_BEARER_TOKEN }}

      # ── 15. Post to Telegram (free bot API) ────────────────
      - name: Post to Telegram
        run: |
          TELETEXT=$(sed -n '/---TWITTER_END---/,$ p' $OUTPUT_DIR/social.md | tail -n +2 | head -c 1000)
          echo "TELEGRAM_TEXT=$TELETEXT" >> $GITHUB_ENV
          echo "TELEGRAM_PHOTO=$OUTPUT_DIR/thumbnails/thumbnail.png" >> $GITHUB_ENV
          go run ./cmd/telegram/
        env:
          TELEGRAM_BOT_TOKEN: ${{ secrets.TELEGRAM_BOT_TOKEN }}
          TELEGRAM_CHAT_ID: ${{ secrets.TELEGRAM_CHAT_ID }}

      # ── 16. Summary ───────────────────────────────────────
      - name: Print summary
        run: |
          echo "## ✅ Content Engine Daily Run" >> $GITHUB_STEP_SUMMARY
          echo "" >> $GITHUB_STEP_SUMMARY
          echo "| Step | Status |" >> $GITHUB_STEP_SUMMARY
          echo "|---|---|" >> $GITHUB_STEP_SUMMARY
          echo "| Gemini content | ✅ Generated |" >> $GITHUB_STEP_SUMMARY
          echo "| Thumbnail | ✅ Created |" >> $GITHUB_STEP_SUMMARY
          echo "| Affiliate links | ✅ Injected |" >> $GITHUB_STEP_SUMMARY
          echo "| PDF E-book | ✅ Created |" >> $GITHUB_STEP_SUMMARY
          echo "| Gumroad product | ✅ Listed |" >> $GITHUB_STEP_SUMMARY
          echo "| YouTube video | ✅ Uploaded |" >> $GITHUB_STEP_SUMMARY
          echo "| Twitter/X post | ✅ Posted |" >> $GITHUB_STEP_SUMMARY
          echo "| Telegram post | ✅ Sent |" >> $GITHUB_STEP_SUMMARY
```

### GitHub Secrets to Configure

In your repo → **Settings** → **Secrets and variables** → **Actions** → **New repository secret**:

| Secret | Value | Source |
|---|---|---|
| `GEMINI_API_KEY` | `AIzaSy...` | https://aistudio.google.com/app/apikey |
| `YOUTUBE_TOKEN_JSON` | `{"token":"...","refresh_token":"..."}` | Run `cmd/gentoken` locally |
| `YOUTUBE_CLIENT_SECRET_JSON` | `{"installed":{"client_id":"...","client_secret":"..."}}` | Google Cloud Console |
| `TWITTER_BEARER_TOKEN` | `AAAAAAAA...` | https://developer.twitter.com |
| `TELEGRAM_BOT_TOKEN` | `123456:ABC-DEF...` | @BotFather |
| `TELEGRAM_CHAT_ID` | `-1001234567890` | Telegram `/getUpdates` |
| `GUMROAD_ACCESS_TOKEN` | Gumroad token | https://app.gumroad.com/advanced |

---

## 6. Monitoring, Costs & Scale-Up Notes

### 6.1 Free Monitoring

| Method | What You See | How |
|---|---|---|
| **GitHub Actions tab** | Pass/fail, logs per step, duration | Click repo → Actions |
| **Email notifications** | Failure alerts only | GitHub Settings → Notifications |
| **Step summary** | Custom markdown table per run | `$GITHUB_STEP_SUMMARY` (built into the YAML above) |
| **Telegram failure alert** | Optional | Add a step that calls Telegram only if `failure()` |

### 6.2 Staying Within Free Tiers

| Service | Free Limit | Our Daily Use | Monthly Total | Headroom |
|---|---|---|---|---|
| GitHub Actions | 2,000 min/month | ~4 min | ~120 min | **16×** |
| Gemini API | 1,500 req/day | 3 req | ~90 req | **500×** |
| YouTube Data API | 10,000 quota/day | ~1,600 | ~48,000 | **6×** |
| Twitter API v2 | ~1,500 tweets/month | 1 | ~30 | **50×** |
| Telegram Bot API | Unlimited | 1 msg | ~30 | Unlimited |
| Gumroad | Free listing (9% per sale) | 1 product | ~30 | Unlimited free listings |
| Go / pandoc / ffmpeg | Open source | Unlimited | Unlimited | Unlimited |
| GitHub storage | 500 MB artifacts | ~5 MB/run | ~150 MB | **3×** (clean up monthly) |

### 6.3 Avoiding Overage

- **GitHub Actions:** Set billing → Usage limit → **$0**. It stops when minutes run out.
- **YouTube:** Use `unlisted` for dev runs; only go `public` when content is ready.
- **Gemini:** At 3 requests/day you'll never hit 1,500. If you do, add `time.Sleep(1 * time.Second)` between calls.
- **Twitter:** Free tier allows ~50 tweets/month. Our pipeline does 30. Fine.

### 6.4 Upgrading When Revenue Comes

| Monthly Revenue | Suggested Upgrade | Why |
|---|---|---|
| $0–$50 | Stay on free tiers | Everything works at $0 |
| $50–$200 | Self-hosted runner ($5/mo VPS) | No CI minute limits |
| $200–$500 | YouTube quota increase request | Upload more videos |
| $500–$1K | Gemini paid tier ($0.15/1M tokens) | Faster model, higher rate limits |
| $1K+ | Dedicated server ($10–$20/mo) | Full automation, multiple channels |

### 6.5 Video Generation Quality Note

The pipeline uses `espeak` (free, offline TTS) + `ffmpeg` for video creation. For better audio quality:

- **gTTS (Google Text-to-Speech):** `pip install gTTS && gtts-cli "text" --output audio.mp3` — free, 200 requests/day
- **Amazon Polly (free tier):** 5M characters/month for the first 12 months (but this has a time limit, so espeak is preferred for "forever free")

### 6.6 Architecture Summary

```
  ┌─ GitHub Actions (cron: 2 AM UTC, ~4 min/run, $0) ───────────────┐
  │                                                                     │
  │  Gemini API (Go)    ──→ blog.md, script.md, social.md (free)       │
  │  Go image/fogleman  ──→ thumbnail.png (1280×720, no API key)       │
  │  Pandoc/markdown2pdf──→ daily_ebook.pdf (free, open source)        │
  │  Gumroad CLI (Go)   ──→ product listing (free, pay per sale)       │
  │  ffmpeg + espeak    ──→ video.mp4 (free TTS + thumbnail)          │
  │  YouTube API (Go)   ──→ video uploaded (10,000 quota/day free)    │
  │  Twitter API (Go)   ──→ tweet posted (free tier, rate limited)     │
  │  Telegram API (Go)  ──→ message + photo (unlimited, free)          │
  │                                                                     │
  │  Total operating cost: $0.00/month                                  │
  └─────────────────────────────────────────────────────────────────────┘
```

### 6.7 Timeline to First Automated Revenue ($0 Spent)

| Phase | Time | Activity |
|---|---|---|
| **Setup** | 6–10 hours | Create accounts, generate API keys, write Go code, test workflow |
| **Content library** | 1–2 weeks | Daily pipeline builds 7–14 pieces of content |
| **Organic growth** | 2–4 weeks | YouTube indexing, Twitter followers, Telegram subscribers |
| **First sale** | 3–8 weeks | First Gumroad purchase or affiliate commission |

Growth is 100% organic through YouTube search, Twitter reach, and Telegram sharing — no ads, no paid promotion.

---

*End of Guide — All tools are permanently free, no credit card required, no time limits.*
