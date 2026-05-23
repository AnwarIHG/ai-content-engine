package twitter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type tweetReq struct {
	Text string `json:"text"`
}

type tweetResp struct {
	Data struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"data"`
}

func PostTweet(text string) (string, error) {
	bearer := os.Getenv("TWITTER_BEARER_TOKEN")
	if bearer == "" {
		return "", fmt.Errorf("TWITTER_BEARER_TOKEN not set")
	}

	if len(text) > 280 {
		text = text[:280]
	}

	body, _ := json.Marshal(tweetReq{Text: text})
	req, _ := http.NewRequest("POST", "https://api.twitter.com/2/tweets",
		bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("twitter request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		var errResp map[string]any
		json.NewDecoder(resp.Body).Decode(&errResp)
		return "", fmt.Errorf("twitter %d: %v", resp.StatusCode, errResp)
	}

	var result tweetResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("twitter decode: %w", err)
	}

	return result.Data.ID, nil
}
