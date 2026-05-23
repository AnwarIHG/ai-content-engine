# AI Content Engine

Zero-cost automated content pipeline: generates, publishes, and monetizes daily content using only permanently-free tools.

## Pipeline (daily at 2 AM UTC, ~4 min/run)

```
Gemini API  ──→ blog.md + script.md + social.md
Go image/gg ──→ thumbnail.png (1280×720)
pandoc      ──→ daily_ebook.pdf
Gumroad CLI ──→ product listing ($4.99)
ffmpeg+TTS  ──→ video.mp4
YouTube API ──→ video uploaded (unlisted/public)
Twitter API ──→ tweet posted
Telegram API ──→ message + photo sent
```

## Setup (6–10 hours)

1. Fork this repo
2. Set 8 secrets in **Settings → Secrets → Actions**:

| Secret | Source |
|---|---|
| `GEMINI_API_KEY` | https://aistudio.google.com/app/apikey |
| `YOUTUBE_TOKEN_JSON` | `go run ./cmd/gentoken/` (run locally once) |
| `YOUTUBE_CLIENT_SECRET_JSON` | Google Cloud Console |
| `TWITTER_BEARER_TOKEN` | https://developer.twitter.com |
| `TELEGRAM_BOT_TOKEN` | @BotFather |
| `TELEGRAM_CHAT_ID` | Telegram `/getUpdates` |
| `GUMROAD_ACCESS_TOKEN` | https://app.gumroad.com/advanced |
| `AFFILIATE_CONFIG` | (optional) JSON override of keyword→link map |

3. Push — the YAML in `.github/workflows/content_engine.yml` runs automatically

## Cost

**$0.00/month.** Every tool has a permanent free tier with no credit card required. GitHub Actions (2,000 min/mo free) uses ~120 min/mo.

## Structure

```
cmd/gencontent/   — Gemini content generation
cmd/thumbnail/    — 1280×720 PNG renderer
cmd/youtube/      — YouTube Data API v3 upload
cmd/twitter/      — Twitter API v2 tweet
cmd/telegram/     — Telegram Bot API message
cmd/ebook/        — pandoc PDF + Gumroad product
cmd/affiliate/    — keyword→affiliate link injection
cmd/gentoken/     — one-time YouTube OAuth token
```

## Requirement

Go 1.22+ (for local testing). All dependencies are fetched automatically by `go mod tidy`.

## Revenue Timeline

- Content library → 1–2 weeks
- Organic growth → 2–4 weeks
- First sale → 3–8 weeks ($0 spent on ads)
