package telegram

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func SendMessage(ctx context.Context, text string) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if token == "" || chatID == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID required")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := fmt.Sprintf(`{"chat_id":%s,"text":%q}`, chatID, text)

	req, _ := http.NewRequestWithContext(ctx, "POST", url,
		bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram %d", resp.StatusCode)
	}
	return nil
}

func SendPhoto(ctx context.Context, caption, photoPath string) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if token == "" || chatID == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID required")
	}

	file, err := os.Open(photoPath)
	if err != nil {
		return fmt.Errorf("open photo: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	w.WriteField("chat_id", chatID)
	w.WriteField("caption", caption)

	fw, err := w.CreateFormFile("photo", filepath.Base(photoPath))
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(fw, file); err != nil {
		return fmt.Errorf("copy photo: %w", err)
	}
	w.Close()

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", token)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram photo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram photo %d", resp.StatusCode)
	}
	return nil
}
