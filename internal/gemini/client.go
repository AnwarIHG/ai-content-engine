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

type Part struct {
	Text string `json:"text"`
}

type Content struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role,omitempty"`
}

type GenerateRequest struct {
	Contents          []Content      `json:"contents"`
	SystemInstruction *Content       `json:"system_instruction,omitempty"`
	GenerationConfig  map[string]any `json:"generation_config,omitempty"`
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
				Role:  "user",
				Parts: []Part{{Text: userPrompt}},
			},
		},
		GenerationConfig: map[string]any{
			"temperature":     0.7,
			"maxOutputTokens": 2048,
		},
	}

	if systemPrompt != "" {
		req.SystemInstruction = &Content{
			Parts: []Part{{Text: systemPrompt}},
		}
	}

	body, _ := json.Marshal(req)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
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
		model = "gemini-2.0-flash"
	}
	return Generate(key, model, systemPrompt, userPrompt)
}
