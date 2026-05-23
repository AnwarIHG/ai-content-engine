package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

func main() {
	clientSecretBytes, err := os.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("read client_secret.json: %v", err)
	}

	config, err := google.ConfigFromJSON(clientSecretBytes, youtube.YoutubeUploadScope)
	if err != nil {
		log.Fatalf("parse client secret: %v", err)
	}

	// Use local server OAuth flow (opens browser)
	// Fall back to console flow if browser unavailable
	token, err := getToken(config)
	if err != nil {
		log.Fatalf("get token: %v", err)
	}

	tokenJSON, _ := json.MarshalIndent(token, "", "  ")
	fmt.Println(string(tokenJSON))
}

func getToken(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline,
		oauth2.ApprovalForce)
	fmt.Printf("\nOpen this URL in your browser:\n%s\n", authURL)
	fmt.Print("\nEnter the authorization code: ")

	var code string
	fmt.Scan(&code)

	return config.Exchange(context.Background(), code)
}
