package affiliate

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Keywords   map[string]string `json:"keywords"`
	Disclaimer string            `json:"disclaimer"`
}

var defaultConfig = Config{
	Keywords: map[string]string{
		"productivity tool": "https://www.amazon.com/s?k=productivity+tool&tag=YOUR_TAG-20",
		"AI tool":           "https://www.amazon.com/s?k=AI+tool&tag=YOUR_TAG-20",
		"software":          "https://www.amazon.com/s?k=productivity+software&tag=YOUR_TAG-20",
	},
	Disclaimer: "\n\n---\n*Some links above are affiliate links. I may earn a commission at no extra cost to you. As an Amazon Associate I earn from qualifying purchases.*",
}

func Inject(text string, cfg *Config) (string, error) {
	if cfg == nil {
		cfg = &defaultConfig
	}

	// Load from env JSON if provided
	if envCfg := os.Getenv("AFFILIATE_CONFIG"); envCfg != "" {
		var custom Config
		if err := json.Unmarshal([]byte(envCfg), &custom); err == nil {
			if custom.Keywords != nil {
				cfg = &custom
			}
		}
	}

	result := text
	for keyword, url := range cfg.Keywords {
		replacement := fmt.Sprintf("[%s](%s)", keyword, url)
		result = strings.ReplaceAll(
			strings.ToLower(result),
			strings.ToLower(keyword),
			replacement,
		)
	}

	if cfg.Disclaimer != "" && !strings.Contains(result, cfg.Disclaimer) {
		result += cfg.Disclaimer
	}

	return result, nil
}
